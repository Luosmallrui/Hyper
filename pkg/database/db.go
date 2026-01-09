package database

import (
	"Hyper/config"
	"Hyper/pkg/log"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// NewDB 初始化数据库连接
func NewDB(conf *config.Config) *gorm.DB {
	dsn := conf.MySQL.Dsn()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.L.Error("failed to connect database", zap.Error(err))
	}
	log.L.Info("connect database success")
	return db
}
