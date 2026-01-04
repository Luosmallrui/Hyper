//go:build wireinject
// +build wireinject

package main

import (
	"Hyper/config"
	"Hyper/dao"
	"Hyper/handler"
	"Hyper/pkg/client"
	"Hyper/pkg/database"
	"Hyper/pkg/rocketmq"
	"Hyper/pkg/server"
	"Hyper/service"

	"github.com/google/wire"
)

func InitServer(cfg *config.Config) *server.AppProvider {
	wire.Build(
		database.NewDB,
		client.NewRedisClient,
		config.ProvideOssConfig,
		rocketmq.InitRocketmqClient,
		server.NewGinEngine,
		wire.Struct(new(handler.Auth), "*"),
		wire.Struct(new(handler.Map), "*"),
		wire.Struct(new(handler.Note), "*"),

		wire.Struct(new(server.AppProvider), "*"),
		wire.Struct(new(server.Handlers), "*"),

		wire.Struct(new(handler.MessageHandler), "*"),
		//wire.Struct(new(handler.WSHandler), "*"),

		dao.ProviderSet,

		service.ProviderSet,
		//service.NewMessageReadService,
	)
	return nil
}
