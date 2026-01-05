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

<<<<<<< HEAD
	wire.Struct(new(MessageService), "*"),
	wire.Bind(new(IMessageService), new(*MessageService)),

	wire.Struct(new(ClientConnectService), "*"),
	wire.Bind(new(IClientConnectService), new(*ClientConnectService)),

	wire.Struct(new(GroupMemberService), "*"),
	wire.Bind(new(IGroupMemberService), new(*GroupMemberService)),
=======
	wire.Struct(new(LikeService), "*"),
	wire.Bind(new(ILikeService), new(*LikeService)),
>>>>>>> 3bc8e5d (实现点赞、取消点赞、查询点赞次数、查询用户自己是否已经点赞接口)

	NewOssService,
)
