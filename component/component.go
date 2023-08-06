package component

import "context"

// tcc 请求参数
type TCCReq struct {
	// 全局唯一的事务 id
	ComponentID string                 `json:"componentID"`
	TXID        string                 `json:"txID"`
	Data        map[string]interface{} `json:"data"`
}

// tcc 响应结果
type TCCResp struct {
	ComponentID string `json:"componentID"`
	ACK         bool   `json:"ack"`
	TXID        string `json:"txID"`
}

// tcc 组件
type TCCComponent interface {
	ID() string
	Try(ctx context.Context, req *TCCReq) (*TCCResp, error)
	Confirm(ctx context.Context, txID string) (*TCCResp, error)
	Cancel(ctx context.Context, txID string) (*TCCResp, error)
}
