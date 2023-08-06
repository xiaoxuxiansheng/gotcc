package txmanager

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/xiaoxuxiansheng/gotcc/component"
)

// 1. 事务日志存储模块

// 2. TCC 组件注册模块

// 3. 串联两个流程
type TXManager struct {
	ctx            context.Context
	stop           context.CancelFunc
	opts           *Options
	txStore        TXStore
	registryCenter TCCRegistyCenter
}

func NewTXManager(txStore TXStore, registryCenter TCCRegistyCenter, opts ...Option) *TXManager {
	ctx, cancel := context.WithCancel(context.Background())
	txManager := TXManager{
		opts:           &Options{},
		txStore:        txStore,
		registryCenter: registryCenter,
		ctx:            ctx,
		stop:           cancel,
	}

	for _, opt := range opts {
		opt(txManager.opts)
	}

	go txManager.run()
	return &txManager
}

func (t *TXManager) Stop() {
	t.stop()
}

// 事务
func (t *TXManager) Transaction(ctx context.Context, reqs ...*RequestEntity) (bool, error) {
	tctx, cancel := context.WithTimeout(ctx, t.opts.Timeout)
	defer cancel()

	// 获得所有的组件
	componentEntities, err := t.getComponents(tctx, reqs...)
	if err != nil {
		return false, err
	}

	// 1 先创建事务明细记录，并取得全局唯一的事务 id
	txID, err := t.txStore.CreateTX(tctx, componentEntities.ToComponents()...)
	if err != nil {
		return false, err
	}

	// 2. 两阶段提交， try-confirm/cancel
	return t.twoPhaseCommit(ctx, txID, componentEntities)
}

func (t *TXManager) backOffTick(tick time.Duration) time.Duration {
	tick <<= 1
	if tick > t.opts.MonitorTick<<3 {
		return t.opts.MonitorTick << 3
	}
	return tick
}

func (t *TXManager) run() {
	var tick time.Duration
	var err error
	for {
		// 如果出现了失败，tick 需要避让
		if err == nil {
			tick = t.opts.MonitorTick
		} else {
			tick = t.backOffTick(tick)
		}
		select {
		case <-t.ctx.Done():
			return
		case <-time.After(tick):
			// 加锁，避免监控任务重复执行
			if err = t.txStore.Lock(t.ctx, t.opts.MonitorTick); err != nil {
				continue
			}

			// 获取仍然处于 hanging 状态的事务
			var txs []*Transaction
			if txs, err = t.txStore.GetHangingTXs(t.ctx); err != nil {
				_ = t.txStore.Unlock(t.ctx)
				continue
			}

			err = t.batchAdvanceProgress(txs)
			_ = t.txStore.Unlock(t.ctx)
		}
	}
}

func (t *TXManager) batchAdvanceProgress(txs []*Transaction) error {
	// 对每笔事务进行状态推进
	errCh := make(chan error)
	go func() {
		var wg sync.WaitGroup
		for _, tx := range txs {
			txStatus := tx.getStatus(time.Now().Add(-t.opts.Timeout))
			// hanging 状态的暂时不处理
			if txStatus == TXHanging {
				continue
			}

			// shadow
			tx := tx
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := t.advanceProgress(tx, txStatus == TXSucceesful); err != nil {
					errCh <- err
				}
			}()
		}
		wg.Wait()
		close(errCh)
	}()

	var firstErr error
	for err := range errCh {
		if firstErr != nil {
			continue
		}
		firstErr = err
	}

	return firstErr
}

func (t *TXManager) advanceProgress(tx *Transaction, success bool) error {
	// 推进为成功
	var confirmOrCancel func(ctx context.Context, component component.TCCComponent) (*component.TCCResp, error)
	var txAdvanceProgress func(ctx context.Context, componentID string) error
	if success {
		confirmOrCancel = func(ctx context.Context, component component.TCCComponent) (*component.TCCResp, error) {
			return component.Confirm(ctx, tx.TXID)
		}
		txAdvanceProgress = func(ctx context.Context, componentID string) error {
			return t.txStore.TXUpdate(ctx, tx.TXID, componentID, true)
		}

	} else {
		confirmOrCancel = func(ctx context.Context, component component.TCCComponent) (*component.TCCResp, error) {
			return component.Cancel(ctx, tx.TXID)
		}

		txAdvanceProgress = func(ctx context.Context, componentID string) error {
			return t.txStore.TXUpdate(ctx, tx.TXID, componentID, false)
		}
	}

	var componentID string
	for _, component := range tx.Components {
		if componentID == "" {
			componentID = component.Component.ID()
		}
		resp, err := confirmOrCancel(t.ctx, component.Component)
		if err != nil {
			return err
		}
		if !resp.ACK {
			return fmt.Errorf("component: %s ack failed", component.Component.ID())
		}
	}

	// 都执行完成后，对状态进行更新
	return txAdvanceProgress(t.ctx, componentID)
}

func (t *TXManager) twoPhaseCommit(ctx context.Context, txID string, componentEntities ComponentEntities) (bool, error) {
	// 开启轮次，分别执行 try
	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error)
	go func() {
		var wg sync.WaitGroup
		for _, componentEntity := range componentEntities {
			// shadow
			componentEntity := componentEntity
			wg.Add(1)
			go func() {
				defer wg.Done()
				resp, err := componentEntity.Component.Try(cctx, &component.TCCReq{
					ComponentID: componentEntity.Component.ID(),
					TXID:        txID,
					Data:        componentEntity.Request,
				})
				// 但凡有一个 component try 报错或者拒绝，就要进行 cancel 回滚
				if err != nil || !resp.ACK {
					// 对对应的事务进行更新
					_ = t.txStore.TXUpdate(cctx, txID, componentEntity.Component.ID(), false)
					errCh <- fmt.Errorf("component: %s try failed", componentEntity.Component.ID())
					return
				}
				_ = t.txStore.TXUpdate(cctx, txID, componentEntity.Component.ID(), true)
			}()
		}

		wg.Wait()
		close(errCh)
	}()

	successful := true
	if err := <-errCh; err != nil {
		// 只要有一笔 try 请求失败了，其他的都进行终止
		cancel()
		successful = false
	}

	// 执行二阶段
	go t.advanceProgress(NewTransaction(txID, componentEntities), successful)
	return successful, nil
}

// 并发执行，只要中间某次出现了失败，直接终止流程进行 cancel

// 如果全量执行成功，则返回成功的 ack，然后批量执行 confirm

func (t *TXManager) getComponents(ctx context.Context, reqs ...*RequestEntity) (ComponentEntities, error) {
	if len(reqs) == 0 {
		return nil, errors.New("emtpy task")
	}

	// 调一下接口，确认这些都是合法的
	idToReq := make(map[string]*RequestEntity, len(reqs))
	componentIDs := make([]string, 0, len(reqs))
	for _, req := range reqs {
		if _, ok := idToReq[req.ComponentID]; ok {
			return nil, fmt.Errorf("repeat component: %s", req.ComponentID)
		}
		idToReq[req.ComponentID] = req
		componentIDs = append(componentIDs, req.ComponentID)
	}

	// 校验其合法性
	components, err := t.registryCenter.Components(ctx, componentIDs...)
	if err != nil {
		return nil, err
	}
	if len(componentIDs) != len(components) {
		return nil, errors.New("invalid componentIDs ")
	}

	entities := make(ComponentEntities, 0, len(components))
	for _, component := range components {
		entities = append(entities, &ComponentEntity{
			Request:   idToReq[component.ID()].Request,
			Component: component,
		})
	}

	return entities, nil
}
