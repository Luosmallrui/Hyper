//go:build wireinject

package dao

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewAdmin,
	NewUsers,
	NewMapDao,
	NewNoteDAO,
<<<<<<< HEAD
	NewMessageDAO,
	NewGroupDAO,
	NewMessageReadDAO,
	NewGroupMember,
=======
	NewNoteLikeDAO,
	NewNoteStatsDAO,
>>>>>>> 3bc8e5d (实现点赞、取消点赞、查询点赞次数、查询用户自己是否已经点赞接口)
)
