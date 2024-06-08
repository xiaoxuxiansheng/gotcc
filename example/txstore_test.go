package example

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"github.com/xiaoxuxiansheng/gotcc"
	expdao "github.com/xiaoxuxiansheng/gotcc/example/dao"
	"github.com/xiaoxuxiansheng/redis_lock"
)

type mockTXRecordDAO struct {
}

func newMockTXRecordDAO() TXRecordDAO {
	return &mockTXRecordDAO{}
}

func (m *mockTXRecordDAO) GetTXRecords(ctx context.Context, opts ...expdao.QueryOption) ([]*expdao.TXRecordPO, error) {
	componentTryStatuses := map[string]*expdao.ComponentTryStatus{
		"component": {
			ComponentID: "component",
			TryStatus:   gotcc.TryHanging.String(),
		},
	}
	body, _ := json.Marshal(componentTryStatuses)

	tx := expdao.TXRecordPO{
		Status:               gotcc.TXHanging.String(),
		ComponentTryStatuses: string(body),
	}

	return []*expdao.TXRecordPO{
		&tx,
	}, nil
}

func (m *mockTXRecordDAO) CreateTXRecord(ctx context.Context, record *expdao.TXRecordPO) (uint, error) {
	if record.ComponentTryStatuses == "{}" {
		return 0, errors.New("invalid component try statuses")
	}
	return 1, nil
}

func (m *mockTXRecordDAO) UpdateComponentStatus(ctx context.Context, id uint, componentID string, status string) error {
	return nil
}

func (m *mockTXRecordDAO) UpdateTXRecord(ctx context.Context, record *expdao.TXRecordPO) error {
	return nil
}

func (m *mockTXRecordDAO) LockAndDo(ctx context.Context, id uint, do func(ctx context.Context, dao *expdao.TXRecordDAO, record *expdao.TXRecordPO) error) error {
	switch id {
	case 1:
		record := expdao.TXRecordPO{
			Status: gotcc.TXSuccessful.String(),
		}
		return do(ctx, &expdao.TXRecordDAO{}, &record)
	case 2:
		record := expdao.TXRecordPO{
			Status: gotcc.TXFailure.String(),
		}
		return do(ctx, &expdao.TXRecordDAO{}, &record)
	default:
		record := expdao.TXRecordPO{
			Status: gotcc.TXHanging.String(),
		}
		return do(ctx, &expdao.TXRecordDAO{}, &record)
	}
}

func Test_MockTXStore_Lock(t *testing.T) {
	lockErr := "lockErr"
	lockErrCtxKey := &lockErr
	patch := gomonkey.ApplyMethod(reflect.TypeOf(&redis_lock.RedisLock{}), "Lock", func(_ *redis_lock.RedisLock, ctx context.Context) error {
		lockErr, _ := ctx.Value(lockErrCtxKey).(bool)
		if lockErr {
			return errors.New("lock err")
		}
		return nil
	})
	patch = patch.ApplyMethod(reflect.TypeOf(&redis_lock.RedisLock{}), "Unlock", func(_ *redis_lock.RedisLock, ctx context.Context) error {
		return nil
	})
	defer patch.Reset()

	ctx := context.Background()
	mockTXStore := NewMockTXStore(newMockTXRecordDAO(), &redis_lock.Client{})
	err := mockTXStore.Lock(ctx, time.Second)
	assert.Equal(t, nil, err)
	err = mockTXStore.Unlock(ctx)
	assert.Equal(t, nil, err)
}

func Test_MockTXStore_CreateTX(t *testing.T) {
	mockTXStore := NewMockTXStore(newMockTXRecordDAO(), &redis_lock.Client{})

	ctx := context.Background()
	_, err := mockTXStore.CreateTX(ctx)
	assert.Equal(t, true, err != nil)
	_, err = mockTXStore.CreateTX(ctx, NewMockComponent("id", nil))
	assert.Equal(t, nil, err)
}

func Test_MockTXStore_TXUpdate(t *testing.T) {
	mockTXStore := NewMockTXStore(newMockTXRecordDAO(), &redis_lock.Client{})
	err := mockTXStore.TXUpdate(context.Background(), "tx_id", "component_id", true)
	assert.Equal(t, nil, err)
}

func Test_MockTXStore_GetHangingTXs(t *testing.T) {
	mockTXStore := NewMockTXStore(newMockTXRecordDAO(), &redis_lock.Client{})
	_, err := mockTXStore.GetHangingTXs(context.Background())
	assert.Equal(t, nil, err)
}

func Test_MockTXStore_TXSubmit(t *testing.T) {
	patch := gomonkey.ApplyMethod(&expdao.TXRecordDAO{}, "UpdateTXRecord", func(_ *expdao.TXRecordDAO, ctx context.Context, record *expdao.TXRecordPO) error {
		return nil
	})
	defer patch.Reset()

	mockTXStore := NewMockTXStore(newMockTXRecordDAO(), &redis_lock.Client{})
	ctx := context.Background()
	err := mockTXStore.TXSubmit(ctx, "1", false)
	assert.Equal(t, true, err != nil)
	err = mockTXStore.TXSubmit(ctx, "2", true)
	assert.Equal(t, true, err != nil)
	err = mockTXStore.TXSubmit(ctx, "3", true)
	assert.Equal(t, nil, err)
	err = mockTXStore.TXSubmit(ctx, "3", false)
	assert.Equal(t, nil, err)
}

func Test_MockTXStore_GetTX(t *testing.T) {
	mockTXStore := NewMockTXStore(newMockTXRecordDAO(), &redis_lock.Client{})
	_, err := mockTXStore.GetTX(context.Background(), "1")
	assert.Equal(t, nil, err)
}
