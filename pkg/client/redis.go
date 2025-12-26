package client

import (
	"Hyper/config"
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(conf *config.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:        conf.Redis.Host,
		Password:    conf.Redis.Auth,
		DB:          conf.Redis.Database,
		ReadTimeout: 0,
	})
	if _, err := client.Ping(context.TODO()).Result(); err != nil {
		panic(fmt.Errorf("client client error: %s", err))
	}
	return client
}
