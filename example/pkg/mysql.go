package pkg

import (
	"fmt"
	"sync"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const dsn = ""

var (
	db     *gorm.DB
	dbonce sync.Once
)

func NewDB(dsn string, opts ...gorm.Option) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(dsn), opts...)
}

func GetDB() *gorm.DB {
	dbonce.Do(func() {
		var err error
		if db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{}); err != nil {
			panic(fmt.Errorf("failed to connect database, err: %w", err))
		}
	})
	return db
}
