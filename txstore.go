package gotcc

import (
	"context"
	"time"
)

// 事务日志存储模块
type TXStore interface {
	// 创建一条事务明细记录
	CreateTX(ctx context.Context, components ...TCCComponent) (txID string, err error)
	// 更新事务进度：实际更新的是每个组件的 try 请求响应结果
	TXUpdate(ctx context.Context, txID string, componentID string, accept bool) error
	// 提交事务的最终状态, 标识事务执行结果为成功或失败
	TXSubmit(ctx context.Context, txID string, success bool) error
	// 获取到所有未完成的事务
	GetHangingTXs(ctx context.Context) ([]*Transaction, error)
	// 获取指定的一笔事务
	GetTX(ctx context.Context, txID string) (*Transaction, error)
	// 锁住整个 TXStore 模块（要求为分布式锁）
	Lock(ctx context.Context, expireDuration time.Duration) error
	// 解锁TXStore 模块
	Unlock(ctx context.Context) error
}
