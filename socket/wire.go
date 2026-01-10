//go:build wireinject

package socket

import (
	"Hyper/pkg/rocketmq"
	"Hyper/pkg/socket"
	"Hyper/socket/handler"
	"Hyper/socket/handler/event"
	"Hyper/socket/process"

	//"Hyper/socket/process"
	"Hyper/socket/router"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	//business.ProviderSet,
	router.NewRouter,
	socket.NewRoomStorage,
	rocketmq.InitProducer,
	wire.Struct(new(handler.Handler), "*"),

	// process
	wire.Struct(new(process.SubServers), "*"),
	process.NewServer,
	process.NewHealthSubscribe,
	wire.Struct(new(process.NoticeSubscribe), "*"),
	wire.Struct(new(process.MessageSubscribe), "*"),
	//wire.Struct(new(process.QueueSubscribe), "*"),
	//wire.Struct(new(queue.GlobalMessage), "*"),
	//wire.Struct(new(queue.LocalMessage), "*"),
	//wire.Struct(new(queue.RoomControl), "*"),

	handler.ProviderSet,
	event.ProviderSet,
	//consume.ProviderSet,

	// AppProvider
	wire.Struct(new(AppProvider), "*"),
)
