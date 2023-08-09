package pkg

import (
	"fmt"
	"sync"

	"github.com/xiaoxuxiansheng/redis_lock"
)

const (
	network  = "tcp"
	address  = ""
	password = ""
)

var (
	redisClient *redis_lock.Client
	once        sync.Once
)

func NewRedisClient(network, address, password string) *redis_lock.Client {
	return redis_lock.NewClient(network, address, password)
}

func GetRedisClient() *redis_lock.Client {
	once.Do(func() {
		redisClient = redis_lock.NewClient(network, address, password)
	})
	return redisClient
}

// 构造事务 id key，用于幂等去重
func BuildTXKey(componentID, txID string) string {
	return fmt.Sprintf("txKey:%s:%s", componentID, txID)
}

func BuildTXDetailKey(componentID, txID string) string {
	return fmt.Sprintf("txDetailKey:%s:%s", componentID, txID)
}

// 构造请求 id，用于记录状态机
func BuildDataKey(componentID, txID, bizID string) string {
	return fmt.Sprintf("txKey:%s:%s:%s", componentID, txID, bizID)
}

// 构造事务锁 key
func BuildTXLockKey(componentID, txID string) string {
	return fmt.Sprintf("txLockKey:%s:%s", componentID, txID)
}

func BuildTXRecordLockKey() string {
	return "gotcc:txRecord:lock"
}
