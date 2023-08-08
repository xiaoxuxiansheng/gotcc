package example

import (
	"context"
	"encoding/json"
	"time"

	"github.com/demdxx/gocast"
	"github.com/xiaoxuxiansheng/gotcc/component"
	"github.com/xiaoxuxiansheng/gotcc/txmanager"
	"github.com/xiaoxuxiansheng/redis_lock"
)

type MockTXStore struct {
	client *redis_lock.Client
	dao    *TXRecordDAO
}

func NewMockTXStore(dao *TXRecordDAO, client *redis_lock.Client) *MockTXStore {
	return &MockTXStore{
		dao:    dao,
		client: client,
	}
}

func (m *MockTXStore) CreateTX(ctx context.Context, components ...component.TCCComponent) (string, error) {
	// 创建一项内容，里面以唯一事务 id 为 key
	componentTryStatuses := make(map[string]*ComponentTryStatus, len(components))
	for _, component := range components {
		componentTryStatuses[component.ID()] = &ComponentTryStatus{
			ComponentID: component.ID(),
			TryStatus:   txmanager.TryHanging.String(),
		}
	}

	statusesBody, _ := json.Marshal(componentTryStatuses)
	txID, err := m.dao.CreateTXRecord(ctx, &TXRecordPO{
		Status:               HangingStatus,
		ComponentTryStatuses: string(statusesBody),
	})
	if err != nil {
		return "", err
	}

	return gocast.ToString(txID), nil
}

func (m *MockTXStore) TXUpdate(ctx context.Context, txID string, componentID string, accept bool) error {
	do := func(ctx context.Context, dao *TXRecordDAO, record *TXRecordPO) error {
		componentTryStatuses := make(map[string]*ComponentTryStatus)
		_ = json.Unmarshal([]byte(record.ComponentTryStatuses), &componentTryStatuses)
		if accept {
			componentTryStatuses[componentID].TryStatus = txmanager.TrySucceesful.String()
		} else {
			componentTryStatuses[componentID].TryStatus = txmanager.TryFailure.String()
		}
		newBody, _ := json.Marshal(componentTryStatuses)
		record.ComponentTryStatuses = string(newBody)
		return dao.UpdateTXRecord(ctx, record)
	}

	_txID := gocast.ToUint(txID)
	return m.dao.LockAndDo(ctx, _txID, do)
}

func (m *MockTXStore) GetHangingTXs(ctx context.Context) ([]*txmanager.Transaction, error) {
	records, err := m.dao.GetTXRecords(ctx, WithStatus(txmanager.TryHanging))
	if err != nil {
		return nil, err
	}

	txs := make([]*txmanager.Transaction, 0, len(records))
	for _, record := range records {
		componentTryStatuses := make(map[string]*ComponentTryStatus)
		_ = json.Unmarshal([]byte(record.ComponentTryStatuses), &componentTryStatuses)
		components := make([]*txmanager.ComponentTryEntity, 0, len(componentTryStatuses))
		for _, component := range componentTryStatuses {
			components = append(components, &txmanager.ComponentTryEntity{
				ComponentID: component.ComponentID,
				TryStatus:   txmanager.ComponentTryStatus(component.TryStatus),
			})
		}

		txs = append(txs, &txmanager.Transaction{
			TXID:       gocast.ToString(record.ID),
			Status:     txmanager.TXHanging,
			CreatedAt:  record.CreatedAt,
			Components: components,
		})
	}

	return txs, nil
}

func (m *MockTXStore) Lock(ctx context.Context, expireDuration time.Duration) error {
	lock := redis_lock.NewRedisLock(BuildTXRecordLockKey(), m.client, redis_lock.WithExpireSeconds(int64(expireDuration.Seconds())))
	return lock.Lock(ctx)
}

func (m *MockTXStore) Unlock(ctx context.Context) error {
	lock := redis_lock.NewRedisLock(BuildTXRecordLockKey(), m.client)
	return lock.Unlock(ctx)
}
