package pkg

import (
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Test_NewDB(t *testing.T) {
	patch := gomonkey.ApplyMethod(reflect.TypeOf(&mysql.Dialector{}), "Initialize", func(_ *mysql.Dialector, db *gorm.DB) error {
		return nil
	})
	defer patch.Reset()

	db, err := NewDB("", &gorm.Config{
		DisableAutomaticPing: true,
	})

	assert.Equal(t, nil, err)
	patch = patch.ApplyFunc(gorm.Open, func(dialector gorm.Dialector, opts ...gorm.Option) (db *gorm.DB, err error) {
		return &gorm.DB{}, nil
	})
	defer patch.Reset()

	defaultDB := GetDB()
	assert.Equal(t, reflect.TypeOf(defaultDB), reflect.TypeOf(db))
}
