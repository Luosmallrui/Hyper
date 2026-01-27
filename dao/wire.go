//go:build wireinject

package dao

import (
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	NewAdmin,
	NewUsers,
	NewMapDao,
	NewNoteDAO,
	NewMessageDAO,
	NewSessionDAO,
	NewGroup,
	NewMessageReadDAO,
	NewGroupMember,
	NewImage,
	NewNoteLikeDAO,
	NewNoteStatsDAO,
	NewNoteCollectionDAO,
	NewUserFollowDAO,
	NewUserStatsDAO,
	NewComment,
	NewCommentLike,
	NewTopic,
	NewProduct,
)
