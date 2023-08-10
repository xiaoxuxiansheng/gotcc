package example

import (
	"context"
	"testing"
	"time"

	"github.com/xiaoxuxiansheng/gotcc/example/dao"
	"github.com/xiaoxuxiansheng/gotcc/example/pkg"
	"github.com/xiaoxuxiansheng/gotcc/txmanager"
)

const (
	dsn      = "请输入你的 mysql dsn"
	network  = "tcp"
	address  = "请输入你的 redis ip"
	password = "请输入你的 redis 密码"
)

func Test_TCC(t *testing.T) {
	redisClient := pkg.NewRedisClient(network, address, password)
	mysqlDB, err := pkg.NewDB(dsn)
	if err != nil {
		t.Error(err)
		return
	}

	componentAID := "componentA"
	componentBID := "componentB"
	componentCID := "componentC"

	// 构造出对应的 tcc component
	componentA := NewMockComponent(componentAID, redisClient)
	componentB := NewMockComponent(componentBID, redisClient)
	componentC := NewMockComponent(componentCID, redisClient)

	// 构造出事务日志存储模块
	txRecordDAO := dao.NewTXRecordDAO(mysqlDB)
	txStore := NewMockTXStore(txRecordDAO, redisClient)

	txManager := txmanager.NewTXManager(txStore, txmanager.WithMonitorTick(time.Second))
	defer txManager.Stop()

	// 完成各组件的注册
	if err := txManager.Register(componentA); err != nil {
		t.Error(err)
		return
	}

	if err := txManager.Register(componentB); err != nil {
		t.Error(err)
		return
	}

	if err := txManager.Register(componentC); err != nil {
		t.Error(err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
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

	t.Log("success")
}
