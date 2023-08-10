# gotcc

<p align="center">
<img src="https://github.com/xiaoxuxiansheng/gotcc/blob/main/img/sdk_frame.png" height="400px/"><br/><br/>
<b>gotcc: 纯 golang 实现的 tcc sdk 框架</b>
<br/><br/>
</p>

## 📚 前言
正所谓“理论先行，实践紧随”. 使用此框架实战前，建议先行梳理 tcc 相关理论知识，做到知行合一、收放自如<br/><br/>
<img src="https://github.com/xiaoxuxiansheng/gotcc/blob/main/img/tcc_theory_frame.png" height="550px"/>

## 📖 sdk 核心能力
实现了 txManager 事务协调器，完成 try-confirm/cancel 二阶段提交流程的组织串联<br/><br/>
<img src="https://github.com/xiaoxuxiansheng/gotcc/blob/main/img/2pc.png" height="400px"/>

## 💡 `tcc` 技术原理篇与开源实战篇技术博客
<a href="https://xxxx">理论篇 待补充</a> <br/><br/>
<a href="https://xxxx">实战篇 待补充</a>

## 🖥 接入 sop
- 用户需要自行实现事务日志存储模块 TXStore interface，并将其注入到事务协调器 TXManager <br/><br/>
```go
// 事务日志存储模块
type TXStore interface {
	// 创建一条事务明细记录
	CreateTX(ctx context.Context, components ...component.TCCComponent) (txID string, err error)
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
```
- 用户需要自行实现 TCC 组件 TCCComponent，并将其注册到事务协调器 TXManager <br/><br/>
```go
// tcc 组件
type TCCComponent interface {
	// 返回组件唯一 id
	ID() string
	// 执行第一阶段的 try 操作
	Try(ctx context.Context, req *TCCReq) (*TCCResp, error)
	// 执行第二阶段的 confirm 操作
	Confirm(ctx context.Context, txID string) (*TCCResp, error)
	// 执行第二阶段的 cancel 操作
	Cancel(ctx context.Context, txID string) (*TCCResp, error)
}
```

## 🐧 使用示例
使用单测示例代码如下. 其中有关于 txStore 模块的实现类示例，同样参见 package example<br/><br/>
```go
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
```



