package client

import (
	"Hyper/config"
	"Hyper/pkg/log"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func NewRedisClient(conf *config.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:        fmt.Sprintf("%s:%d", conf.Redis.Address, conf.Redis.Port),
		Password:    conf.Redis.Password,
		Username:    conf.Redis.Username,
		DB:          conf.Redis.Database,
		ReadTimeout: 0,
	})
	if _, err := client.Ping(context.TODO()).Result(); err != nil {
		log.L.Fatal("connect redis error", zap.Error(err))
	}
	log.L.Info("redis client success")
	return client
}
