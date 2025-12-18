package database

import (
	"Hyper/config"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// NewDB 初始化数据库连接
func NewDB(conf *config.Config) *gorm.DB {
	dsn := conf.MySQL.Dsn()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Print("failed to connect database")
	}
	fmt.Println("connected to database")
	return db
}
