//go:build wireinject
// +build wireinject

package main

import (
	"Hyper/config"
	"Hyper/dao"
	"Hyper/handler"
	"Hyper/pkg/database"
	"Hyper/pkg/server"
	"Hyper/service"

	"github.com/google/wire"
)

func InitServer(cfg *config.Config) *server.AppProvider {
	wire.Build(
		database.NewDB,
		config.ProvideOssConfig,
		server.NewGinEngine,
		wire.Struct(new(handler.Auth), "*"),
		wire.Struct(new(handler.Map), "*"),
		wire.Struct(new(server.AppProvider), "*"),
		wire.Struct(new(server.Handlers), "*"),
		dao.ProviderSet,
		service.ProviderSet,
	)
	return nil
}
