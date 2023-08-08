package example

import (
	"context"
	"time"

	"github.com/xiaoxuxiansheng/gotcc/component"
	"github.com/xiaoxuxiansheng/gotcc/txmanager"
	"github.com/xiaoxuxiansheng/redis_lock"
)

type TXRecord struct {
	TXID                 string                         `json:"txID"`
	Status               string                         `json:"status"`
	ComponentTryStatuses map[string]*ComponentTryStatus `json:"componentTryStatuses"`
}

type ComponentTryStatus struct {
	ComponentID string `json:"componentID"`
	TryStatus   string `json:"tryStatus"`
}

type MockTXStore struct {
	client *redis_lock.Client
}

func NewMockTXStore(client *redis_lock.Client) *MockTXStore {
	return &MockTXStore{
		client: client,
	}
}

func (m *MockTXStore) CreateTX(ctx context.Context, components ...component.TCCComponent) (txID string, err error) {
	// incr 生成一个全局唯一的事务 id
	m.client.Incr(ctx, BuildTXIDKey())

	// 创建一项内容，里面以唯一事务 id 为 key，

	return "", nil
}

func (m *MockTXStore) TXUpdate(ctx context.Context, txID string, componentID string, accept bool) error {
	return nil
}

func (m *MockTXStore) GetHangingTXs(ctx context.Context) ([]*txmanager.Transaction, error) {
	return nil, nil
}

func (m *MockTXStore) Lock(ctx context.Context, expireDuration time.Duration) error {
	return nil
}

func (m *MockTXStore) Unlock(ctx context.Context) error {
	return nil
}
