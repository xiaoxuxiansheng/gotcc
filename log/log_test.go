package log

import (
	"context"
	"testing"
	"time"
)

func Test_customer_logger(t *testing.T) {
	logger := NewSugarLogger(NewOptions(
		WithFileName("gotcc.log"),
		WithLogLevel("info"),
	))
	logger.Info("test customer logger running...")
}

func Test_default_logger(t *testing.T) {
	now := time.Now()
	Debugf("debug... now: %v", now)
	Infof("info... now: %v", now)
	Warnf("warn... now: %v", now)
	Errorf("error... now: %v", now)
	Fatalf("fatal... now: %v", now)

	ctx := context.Background()
	DebugContext(ctx, "debug...")
	DebugContextf(ctx, "debug... now: %v", now)
	InfoContext(ctx, "info...")
	InfoContextf(ctx, "info... now: %v", now)
	WarnContext(ctx, "warn...")
	WarnContextf(ctx, "warn... now: %v", now)
	ErrorContext(ctx, "error...")
	ErrorContextf(ctx, "error... now: %v", now)
}
