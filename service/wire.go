package service

import (
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	wire.Struct(new(UserService), "*"),
	wire.Bind(new(IUserService), new(*UserService)),
)
