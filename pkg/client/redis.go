package client

import (
	"Hyper/config"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
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
		panic(fmt.Errorf("client client error: %s", err))
	}
	fmt.Println("redis client success")
	return client
}
