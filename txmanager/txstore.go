package txmanager

import (
	"context"
	"time"

	"github.com/xiaoxuxiansheng/gotcc/component"
)

// 事务日志管理
type TXStore interface {
	// 创建一条事务
	CreateTX(ctx context.Context, components ...component.TCCComponent) (txID string, err error)
	// 更新事务进度：
	// 规则为：倘若有一个 component try 操作执行失败，则整个事务失败；倘若所有 component try 操作执行成功，则事务成功
	TXUpdate(ctx context.Context, txID string, componentID string, accept bool) error
	// 提交事务的最终状态
	TXSubmit(ctx context.Context, txID string, success bool) error
	// 获取到所有处于中间态的事务
	GetHangingTXs(ctx context.Context) ([]*Transaction, error)
	// 获取指定的一笔事务
	GetTX(ctx context.Context, txID string) (*Transaction, error)
	// 锁住事务日志表
	Lock(ctx context.Context, expireDuration time.Duration) error
	// 解锁事务日志表
	Unlock(ctx context.Context) error
}
