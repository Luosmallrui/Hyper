package database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// NewDB 初始化数据库连接
// Wire 会调用这个函数来获取 *gorm.DB 对象
func NewDB() *gorm.DB {
	dsn := "root:123456@tcp(127.0.0.1:3306)/ai_demo?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	return db
}
