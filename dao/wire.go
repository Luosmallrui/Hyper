//go:build wireinject

package dao

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewAdmin,
	NewUsers,
	NewMapDao,
)
