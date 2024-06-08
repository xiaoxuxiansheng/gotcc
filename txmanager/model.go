package txmanager

import (
	"time"

	"github.com/xiaoxuxiansheng/gotcc/component"
)

type RequestEntity struct {
	// 组件名称
	ComponentID string `json:"componentName"`
	// 组件入参
	Request map[string]interface{} `json:"request"`
}

type ComponentEntities []*ComponentEntity

func (c ComponentEntities) ToComponents() []component.TCCComponent {
	components := make([]component.TCCComponent, 0, len(c))
	for _, entity := range c {
		components = append(components, entity.Component)
	}
	return components
}

type ComponentEntity struct {
	Request   map[string]interface{}
	Component component.TCCComponent
}

// 事务状态
type TXStatus string

const (
	// 事务执行中
	TXHanging TXStatus = "hanging"
	// 事务成功
	TXSuccessful TXStatus = "successful"
	// 事务失败
	TXFailure TXStatus = "failure"
)

func (t TXStatus) String() string {
	return string(t)
}

type ComponentTryStatus string

func (c ComponentTryStatus) String() string {
	return string(c)
}

const (
	TryHanging ComponentTryStatus = "hanging"
	// 事务成功
	TrySucceesful ComponentTryStatus = "successful"
	// 事务失败
	TryFailure ComponentTryStatus = "failure"
)

type ComponentTryEntity struct {
	ComponentID string
	TryStatus   ComponentTryStatus
}

// 事务
type Transaction struct {
	TXID       string `json:"txID"`
	Components []*ComponentTryEntity
	Status     TXStatus  `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (t *Transaction) getStatus(createdBefore time.Time) TXStatus {
	// 获取事务的状态
	// 1 如果事务超时了，都还未被置为成功，直接置为失败
	if t.CreatedAt.Before(createdBefore) {
		return TXFailure
	}

	// 如果当中出现失败的，直接置为失败
	var hangingExist bool
	for _, component := range t.Components {
		if component.TryStatus == TryFailure {
			return TXFailure
		}
		hangingExist = hangingExist || (component.TryStatus != TrySucceesful)
	}

	// 如果存在组件 try 操作没执行成功，则返回 hanging 状态
	if hangingExist {
		return TXHanging
	}
	return TXSuccessful
}
