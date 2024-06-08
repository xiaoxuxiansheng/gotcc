package example

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/xiaoxuxiansheng/gotcc"
	expdao "github.com/xiaoxuxiansheng/gotcc/example/dao"
	"github.com/xiaoxuxiansheng/gotcc/example/pkg"

	"github.com/demdxx/gocast"
	"github.com/xiaoxuxiansheng/redis_lock"
)

type MockTXStore struct {
	client *redis_lock.Client
	dao    TXRecordDAO
}

func NewMockTXStore(dao TXRecordDAO, client *redis_lock.Client) *MockTXStore {
	return &MockTXStore{
		dao:    dao,
		client: client,
	}
}

func (m *MockTXStore) CreateTX(ctx context.Context, components ...gotcc.TCCComponent) (string, error) {
	// 创建一项内容，里面以唯一事务 id 为 key
	componentTryStatuses := make(map[string]*expdao.ComponentTryStatus, len(components))
	for _, component := range components {
		componentTryStatuses[component.ID()] = &expdao.ComponentTryStatus{
			ComponentID: component.ID(),
			TryStatus:   gotcc.TryHanging.String(),
		}
	}

	statusesBody, _ := json.Marshal(componentTryStatuses)
	txID, err := m.dao.CreateTXRecord(ctx, &expdao.TXRecordPO{
		Status:               gotcc.TXHanging.String(),
		ComponentTryStatuses: string(statusesBody),
	})
	if err != nil {
		return "", err
	}

	return gocast.ToString(txID), nil
}

func (m *MockTXStore) TXUpdate(ctx context.Context, txID string, componentID string, accept bool) error {
	_txID := gocast.ToUint(txID)
	status := gotcc.TXFailure.String()
	if accept {
		status = gotcc.TXSuccessful.String()
	}
	return m.dao.UpdateComponentStatus(ctx, _txID, componentID, status)
}

func (m *MockTXStore) GetHangingTXs(ctx context.Context) ([]*gotcc.Transaction, error) {
	records, err := m.dao.GetTXRecords(ctx, expdao.WithStatus(gotcc.TryHanging))
	if err != nil {
		return nil, err
	}

	txs := make([]*gotcc.Transaction, 0, len(records))
	for _, record := range records {
		componentTryStatuses := make(map[string]*expdao.ComponentTryStatus)
		_ = json.Unmarshal([]byte(record.ComponentTryStatuses), &componentTryStatuses)
		components := make([]*gotcc.ComponentTryEntity, 0, len(componentTryStatuses))
		for _, component := range componentTryStatuses {
			components = append(components, &gotcc.ComponentTryEntity{
				ComponentID: component.ComponentID,
				TryStatus:   gotcc.ComponentTryStatus(component.TryStatus),
			})
		}

		txs = append(txs, &gotcc.Transaction{
			TXID:       gocast.ToString(record.ID),
			Status:     gotcc.TXHanging,
			CreatedAt:  record.CreatedAt,
			Components: components,
		})
	}

	return txs, nil
}

func (m *MockTXStore) Lock(ctx context.Context, expireDuration time.Duration) error {
	lock := redis_lock.NewRedisLock(pkg.BuildTXRecordLockKey(), m.client, redis_lock.WithExpireSeconds(int64(expireDuration.Seconds())))
	return lock.Lock(ctx)
}

func (m *MockTXStore) Unlock(ctx context.Context) error {
	lock := redis_lock.NewRedisLock(pkg.BuildTXRecordLockKey(), m.client)
	return lock.Unlock(ctx)
}

// 提交事务的最终状态
func (m *MockTXStore) TXSubmit(ctx context.Context, txID string, success bool) error {
	do := func(ctx context.Context, dao *expdao.TXRecordDAO, record *expdao.TXRecordPO) error {
		if success {
			if record.Status == gotcc.TXFailure.String() {
				return fmt.Errorf("invalid tx status: %s, txid: %s", record.Status, txID)
			}
			record.Status = gotcc.TXSuccessful.String()
		} else {
			if record.Status == gotcc.TXSuccessful.String() {
				return fmt.Errorf("invalid tx status: %s, txid: %s", record.Status, txID)
			}
			record.Status = gotcc.TXFailure.String()
		}
		return dao.UpdateTXRecord(ctx, record)
	}
	return m.dao.LockAndDo(ctx, gocast.ToUint(txID), do)
}

// 获取指定的一笔事务
func (m *MockTXStore) GetTX(ctx context.Context, txID string) (*gotcc.Transaction, error) {
	records, err := m.dao.GetTXRecords(ctx, expdao.WithID(gocast.ToUint(txID)))
	if err != nil {
		return nil, err
	}
	if len(records) != 1 {
		return nil, errors.New("get tx failed")
	}

	componentTryStatuses := make(map[string]*expdao.ComponentTryStatus)
	_ = json.Unmarshal([]byte(records[0].ComponentTryStatuses), &componentTryStatuses)

	components := make([]*gotcc.ComponentTryEntity, 0, len(componentTryStatuses))
	for _, tryItem := range componentTryStatuses {
		components = append(components, &gotcc.ComponentTryEntity{
			ComponentID: tryItem.ComponentID,
			TryStatus:   gotcc.ComponentTryStatus(tryItem.TryStatus),
		})
	}
	return &gotcc.Transaction{
		TXID:       txID,
		Status:     gotcc.TXStatus(records[0].Status),
		Components: components,
		CreatedAt:  records[0].CreatedAt,
	}, nil
}

type TXRecordDAO interface {
	GetTXRecords(ctx context.Context, opts ...expdao.QueryOption) ([]*expdao.TXRecordPO, error)
	CreateTXRecord(ctx context.Context, record *expdao.TXRecordPO) (uint, error)
	UpdateComponentStatus(ctx context.Context, id uint, componentID string, status string) error
	UpdateTXRecord(ctx context.Context, record *expdao.TXRecordPO) error
	LockAndDo(ctx context.Context, id uint, do func(ctx context.Context, dao *expdao.TXRecordDAO, record *expdao.TXRecordPO) error) error
}
