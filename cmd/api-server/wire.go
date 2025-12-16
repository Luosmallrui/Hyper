//go:build wireinject
// +build wireinject

package main

import (
	"Hyper/internal/module/user"
	"Hyper/pkg/database"
	"Hyper/server"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitServer() *gin.Engine {
	wire.Build(
		database.NewDB,
		server.NewGinEngine,
		server.NewHandlers,
		user.ProviderSet,
	)
	return nil
}
