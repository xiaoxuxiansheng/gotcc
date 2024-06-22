package gotcc

import (
	"time"
)

type RequestEntity struct {
	// 组件名称
	ComponentID string `json:"componentName"`
	// 组件入参
	Request map[string]interface{} `json:"request"`
}

type ComponentEntities []*ComponentEntity

func (c ComponentEntities) ToComponents() []TCCComponent {
	components := make([]TCCComponent, 0, len(c))
	for _, entity := range c {
		components = append(components, entity.Component)
	}
	return components
}

type ComponentEntity struct {
	Request   map[string]interface{}
	Component TCCComponent
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
	// 1 如果当中出现失败的，直接置为失败
	var hangingExist bool
	for _, component := range t.Components {
		if component.TryStatus == TryFailure {
			return TXFailure
		}
		hangingExist = hangingExist || (component.TryStatus != TrySucceesful)
	}

	// 2 如果存在 hanging 状态，并且已经超时，也直接置为失败
	if hangingExist && t.CreatedAt.Before(createdBefore) {
		return TXFailure
	}

	// 3 如果存在组件 try 操作处于 hanging 状态，则返回 hanging 状态
	if hangingExist {
		return TXHanging
	}

	// 4 走到这个分支必然意味着所有组件的 try 操作都成功了
	return TXSuccessful
}
