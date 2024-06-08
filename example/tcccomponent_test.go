package example

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	"github.com/xiaoxuxiansheng/gotcc"
	"github.com/xiaoxuxiansheng/gotcc/example/pkg"
	"github.com/xiaoxuxiansheng/redis_lock"
)

func Test_MockComponent_Try(t *testing.T) {
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
	patch = patch.ApplyMethod(reflect.TypeOf(&redis_lock.Client{}), "Get", func(_ *redis_lock.Client, ctx context.Context, key string) (string, error) {
		switch key {
		case pkg.BuildTXKey("id", "err"):
			return "", errors.New("getErr")
		case pkg.BuildTXKey("id", "repeat"):
			return TXConfirmed.String(), nil
		case pkg.BuildTXKey("id", "cancel"):
			return TXCanceled.String(), nil
		default:
			return "", nil
		}
	})
	patch = patch.ApplyMethod(reflect.TypeOf(&redis_lock.Client{}), "Set", func(_ *redis_lock.Client, ctx context.Context, key string, value string) (int64, error) {
		if value == "setTXToBizErr" {
			return -1, errors.New("setTXToBizErr")
		}
		if key == pkg.BuildTXKey("id", "setTxStatusErr") {
			return -1, errors.New("setTxStatusErr")
		}

		return 1, nil
	})
	patch = patch.ApplyMethod(reflect.TypeOf(&redis_lock.Client{}), "SetNX", func(_ *redis_lock.Client, ctx context.Context, key string, value string) (int64, error) {
		switch key {
		case pkg.BuildDataKey("id", "tx", "frozeBizErr"):
			return -1, errors.New("frozeBizErr")
		case pkg.BuildDataKey("id", "tx", "frozeBizFail"):
			return 0, nil
		default:
			return 1, nil
		}
	})
	defer patch.Reset()

	ctx := context.Background()
	mockComponent := NewMockComponent("id", &redis_lock.Client{})
	assert.Equal(t, "id", mockComponent.ID())

	tests := []struct {
		name      string
		ctx       context.Context
		req       *gotcc.TCCReq
		expectErr bool
		ack       bool
	}{
		{
			name:      "lockErr",
			ctx:       context.WithValue(ctx, lockErrCtxKey, true),
			req:       &gotcc.TCCReq{},
			expectErr: true,
		},
		{
			name: "getTXKeyErr",
			ctx:  ctx,
			req: &gotcc.TCCReq{
				TXID: "err",
			},
			expectErr: true,
		},
		{
			name: "getTXKeyRepeat",
			ctx:  ctx,
			req: &gotcc.TCCReq{
				TXID: "repeat",
			},
			ack: true,
		},
		{
			name: "getTXKeyCancel",
			ctx:  ctx,
			req: &gotcc.TCCReq{
				TXID: "cancel",
			},
		},
		{
			name: "setTXToBizErr",
			ctx:  ctx,
			req: &gotcc.TCCReq{
				TXID: "tx",
				Data: map[string]interface{}{
					"biz_id": "setTXToBizErr",
				},
			},
			expectErr: true,
		},
		{
			name: "frozeBizErr",
			ctx:  ctx,
			req: &gotcc.TCCReq{
				TXID: "tx",
				Data: map[string]interface{}{
					"biz_id": "frozeBizErr",
				},
			},
			expectErr: true,
		},
		{
			name: "frozeBizFail",
			ctx:  ctx,
			req: &gotcc.TCCReq{
				TXID: "tx",
				Data: map[string]interface{}{
					"biz_id": "frozeBizFail",
				},
			},
		},
		{
			name: "setTxStatusErr",
			ctx:  ctx,
			req: &gotcc.TCCReq{
				TXID: "setTxStatusErr",
			},
			expectErr: true,
		},
		{
			name: "success",
			ctx:  ctx,
			req: &gotcc.TCCReq{
				TXID: "success",
			},
			ack: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := mockComponent.Try(tt.ctx, tt.req)
			assert.Equal(t, tt.expectErr, err != nil)
			if err != nil {
				return
			}
			assert.Equal(t, tt.ack, resp.ACK)
		})
	}
}

func Test_MockComponent_Confirm(t *testing.T) {
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
	patch = patch.ApplyMethod(reflect.TypeOf(&redis_lock.Client{}), "Get", func(_ *redis_lock.Client, ctx context.Context, key string) (string, error) {
		switch key {
		case pkg.BuildTXKey("id", "err"):
			return "", errors.New("getErr")
		case pkg.BuildTXKey("id", "repeat"):
			return TXConfirmed.String(), nil
		case pkg.BuildTXKey("id", "cancel"):
			return TXCanceled.String(), nil
		case pkg.BuildTXDetailKey("id", "txToBizErr"):
			return "", errors.New("txToBizErr")
		case pkg.BuildDataKey("id", "getBizErr", "tried"):
			return "", errors.New("getBizErr")
		case pkg.BuildDataKey("id", "getBizUnfrozen", "tried"):
			return "", nil
		case pkg.BuildDataKey("id", "setBizResErr", "tried"):
			return DataFrozen.String(), nil
		case pkg.BuildDataKey("id", "success", "tried"):
			return DataFrozen.String(), nil
		default:
			return TXTried.String(), nil
		}
	})

	patch = patch.ApplyMethod(reflect.TypeOf(&redis_lock.Client{}), "Set", func(_ *redis_lock.Client, ctx context.Context, key string, value string) (int64, error) {
		if key == pkg.BuildDataKey("id", "setBizResErr", "tried") {
			return -1, errors.New("setBizResErr")
		}
		return 1, nil
	})
	defer patch.Reset()

	ctx := context.Background()
	mockComponent := NewMockComponent("id", &redis_lock.Client{})
	assert.Equal(t, "id", mockComponent.ID())

	tests := []struct {
		name      string
		ctx       context.Context
		txid      string
		expectErr bool
		ack       bool
	}{
		{
			name:      "lockErr",
			ctx:       context.WithValue(ctx, lockErrCtxKey, true),
			expectErr: true,
		},
		{
			name:      "getTXKeyErr",
			ctx:       ctx,
			txid:      "err",
			expectErr: true,
		},
		{
			name: "getTXKeyRepeat",
			ctx:  ctx,
			txid: "repeat",
			ack:  true,
		},
		{
			name: "getTXKeyCancel",
			ctx:  ctx,
			txid: "cancel",
		},
		{
			name:      "txToBizErr",
			ctx:       ctx,
			txid:      "txToBizErr",
			expectErr: true,
		},
		{
			name:      "getBizErr",
			ctx:       ctx,
			txid:      "getBizErr",
			expectErr: true,
		},
		{
			name: "getBizUnfrozen",
			ctx:  ctx,
			txid: "getBizUnfrozen",
		},
		{
			name:      "setBizResErr",
			ctx:       ctx,
			txid:      "setBizResErr",
			expectErr: true,
		},
		{
			name: "success",
			ctx:  ctx,
			txid: "success",
			ack:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := mockComponent.Confirm(tt.ctx, tt.txid)
			assert.Equal(t, tt.expectErr, err != nil)
			if err != nil {
				return
			}
			assert.Equal(t, tt.ack, resp.ACK)
		})
	}
}

func Test_MockComponent_Cancel(t *testing.T) {
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
	patch = patch.ApplyMethod(reflect.TypeOf(&redis_lock.Client{}), "Get", func(_ *redis_lock.Client, ctx context.Context, key string) (string, error) {
		switch key {
		case pkg.BuildTXKey("id", "err"):
			return "", errors.New("getErr")
		case pkg.BuildTXKey("id", "invalidTXStatus"):
			return TXConfirmed.String(), nil
		case pkg.BuildTXDetailKey("id", "getBizErr"):
			return "", errors.New("getBizErr")
		case pkg.BuildTXDetailKey("id", "deleteBizFrozeErr"):
			return "deleteBizFrozeErr", nil
		default:
			return TXTried.String(), nil
		}
	})
	patch = patch.ApplyMethod(reflect.TypeOf(&redis_lock.Client{}), "Del", func(_ *redis_lock.Client, ctx context.Context, key string) error {
		if key == pkg.BuildDataKey("id", "deleteBizFrozeErr", "deleteBizFrozeErr") {
			return errors.New("deleteBizFrozeErr")
		}
		return nil
	})

	patch = patch.ApplyMethod(reflect.TypeOf(&redis_lock.Client{}), "Set", func(_ *redis_lock.Client, ctx context.Context, key string, value string) (int64, error) {
		return 1, nil
	})
	defer patch.Reset()

	ctx := context.Background()
	mockComponent := NewMockComponent("id", &redis_lock.Client{})
	assert.Equal(t, "id", mockComponent.ID())

	tests := []struct {
		name      string
		ctx       context.Context
		txid      string
		expectErr bool
		ack       bool
	}{
		{
			name:      "lockErr",
			ctx:       context.WithValue(ctx, lockErrCtxKey, true),
			expectErr: true,
		},
		{
			name:      "getTXKeyErr",
			ctx:       ctx,
			txid:      "err",
			expectErr: true,
		},
		{
			name:      "invalidTXStatus",
			ctx:       ctx,
			txid:      "invalidTXStatus",
			expectErr: true,
		},
		{
			name:      "getBizErr",
			ctx:       ctx,
			txid:      "getBizErr",
			expectErr: true,
		},
		{
			name:      "deleteBizFrozeErr",
			ctx:       ctx,
			txid:      "deleteBizFrozeErr",
			expectErr: true,
		},
		{
			name: "success",
			ctx:  ctx,
			txid: "success",
			ack:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := mockComponent.Cancel(tt.ctx, tt.txid)
			assert.Equal(t, tt.expectErr, err != nil)
			if err != nil {
				return
			}
			assert.Equal(t, tt.ack, resp.ACK)
		})
	}
}
