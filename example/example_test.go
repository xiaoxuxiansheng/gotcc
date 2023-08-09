package example

import (
	"context"
	"testing"
	"time"

	"github.com/xiaoxuxiansheng/gotcc/example/dao"
	"github.com/xiaoxuxiansheng/gotcc/example/pkg"
	"github.com/xiaoxuxiansheng/gotcc/txmanager"
)

func Test_TCC(t *testing.T) {
	redisClient := pkg.GetRedisClient()
	componentAID := "componentA"
	componentBID := "componentB"
	componentCID := "componentC"

	// 构造出对应的 tcc component
	componentA := NewMockComponent(componentAID, redisClient)
	componentB := NewMockComponent(componentBID, redisClient)
	componentC := NewMockComponent(componentCID, redisClient)

	// 创建注册中心
	registryCenter := NewMockRegistryCenter()

	// 完成各组件的注册
	if err := registryCenter.Register(componentA); err != nil {
		t.Error(err)
		return
	}

	if err := registryCenter.Register(componentB); err != nil {
		t.Error(err)
		return
	}

	if err := registryCenter.Register(componentC); err != nil {
		t.Error(err)
		return
	}

	// 构造出事务日志存储模块
	txRecordDAO := dao.NewTXRecordDAO(pkg.GetDB())
	txStore := NewMockTXStore(txRecordDAO, redisClient)

	txManager := txmanager.NewTXManager(txStore, registryCenter)
	defer txManager.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	success, err := txManager.Transaction(ctx, []*txmanager.RequestEntity{
		{ComponentID: componentAID,
			Request: map[string]interface{}{
				"biz_id": componentAID + "_biz",
			},
		},
		{ComponentID: componentBID,
			Request: map[string]interface{}{
				"biz_id": componentBID + "_biz",
			},
		},
		{ComponentID: componentCID,
			Request: map[string]interface{}{
				"biz_id": componentCID + "_biz",
			},
		},
	}...)
	if err != nil {
		t.Errorf("tx failed, err: %v", err)
		return
	}
	if !success {
		t.Error("tx failed")
		return
	}

	t.Error("success")
}
