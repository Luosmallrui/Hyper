//go:build wireinject
// +build wireinject

package main

import (
	"Hyper/config"
	"Hyper/dao"
	"Hyper/dao/cache"
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

		client.NewRedisClient,
		config.ProvideOssConfig,
		config.ProvideRocketMQConfig,
		rocketmq.InitProducer,
		server.NewGinEngine,
		cache.ProviderSet,
		wire.Struct(new(handler.Auth), "*"),
		wire.Struct(new(handler.Map), "*"),
		wire.Struct(new(handler.Note), "*"),
		wire.Struct(new(handler.Follow), "*"),
		wire.Struct(new(handler.User), "*"),
		wire.Struct(new(handler.Session), "*"),
		wire.Struct(new(handler.Message), "*"),
		wire.Struct(new(handler.CommentsHandler), "*"),
		wire.Struct(new(handler.TopicHandler), "*"),
		wire.Struct(new(handler.GroupHandler), "*"),
		wire.Struct(new(handler.GroupMemberHandler), "*"),
		wire.Struct(new(handler.Party), "*"),
		wire.Struct(new(handler.PointHandler), "*"),
		wire.Struct(new(handler.Order), "*"),
		wire.Struct(new(handler.SearchHandler), "*"),
		wire.Struct(new(handler.ProductHandler), "*"),

		wire.Struct(new(server.AppProvider), "*"),
		wire.Struct(new(server.Handlers), "*"),

		dao.ProviderSet,

		service.ProviderSet,
		handler.NewPay,
		database.NewDB,
	)
	return nil
}
