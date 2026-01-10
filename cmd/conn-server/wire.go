//go:build wireinject
// +build wireinject

package main

import (
	"Hyper/config"
	"Hyper/dao"
	"Hyper/dao/cache"
	"Hyper/pkg/client"
	"Hyper/pkg/database"
	"Hyper/pkg/rocketmq"
	"Hyper/service"
	"Hyper/socket"

	"github.com/google/wire"
)

func InitSocketServer(cfg *config.Config) *socket.AppProvider {
	wire.Build(
		database.NewDB,
		client.NewRedisClient,
		dao.ProviderSet,
		rocketmq.InitConsumer,
		cache.ProviderSet,
		socket.ProviderSet,
		service.ProviderSet,
	)
	return nil
}
