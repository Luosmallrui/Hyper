package types

import "Hyper/models"

type GetFollowingFeedResponse struct {
	Following  []*models.FollowingQueryResult `json:"following"`
	NextCursor int64                          `json:"next_cursor"` // 返回给前端，下次请求带上
	HasMore    bool                           `json:"has_more"`    // 告诉前端是否还有更多
}
type GetFollowingListRequest struct {
	Type     string `form:"type" binding:"required"` // following | follower
	Cursor   int64  `form:"cursor"`                  // 游标（时间戳）
	PageSize int    `form:"pageSize"`                // 每页数量
}

type FollowerRequest struct {
	UserId string `json:"user_id"`
}
