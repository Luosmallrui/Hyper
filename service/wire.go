package service

import (
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	wire.Struct(new(UserService), "*"),
	wire.Bind(new(IUserService), new(*UserService)),

	wire.Struct(new(WeChatService), "*"),
	wire.Bind(new(IWeChatService), new(*WeChatService)),

	wire.Struct(new(MapService), "*"),
	wire.Bind(new(IMapService), new(*MapService)),

	wire.Struct(new(NoteService), "*"),
	wire.Bind(new(INoteService), new(*NoteService)),

	wire.Struct(new(LikeService), "*"),
	wire.Bind(new(ILikeService), new(*LikeService)),

	NewOssService,
)
