package pkg

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetRedisClient(t *testing.T) {
	assert.Equal(t, reflect.TypeOf(NewRedisClient("", "", "")), reflect.TypeOf(GetRedisClient()))
}

func Test_BuildKey(t *testing.T) {
	assert.Equal(t, "txKey:component:tx", BuildTXKey("component", "tx"))
	assert.Equal(t, "txDetailKey:component:tx", BuildTXDetailKey("component", "tx"))
	assert.Equal(t, "txKey:component:tx:biz", BuildDataKey("component", "tx", "biz"))
	assert.Equal(t, "txLockKey:component:tx", BuildTXLockKey("component", "tx"))
	assert.Equal(t, "gotcc:txRecord:lock", BuildTXRecordLockKey())
}
