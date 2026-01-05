//go:build wireinject
// +build wireinject

// socket/handler/wire.go
package handler

import (
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	wire.Struct(new(ChatChannel), "*"),
)
