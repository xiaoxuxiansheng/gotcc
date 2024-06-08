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
	defer patch.Reset()

	ctx := context.Background()
	mockComponent := NewMockComponent("id", &redis_lock.Client{})
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
