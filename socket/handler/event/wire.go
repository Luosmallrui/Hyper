//go:build wireinject
// +build wireinject

// chatroom/socket/handler/event/wire/go
package event

import (
	"Hyper/socket/handler/event/chat"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	wire.Struct(new(ChatEvent), "*"),
	wire.Struct(new(chat.Handler), "*"),
)
