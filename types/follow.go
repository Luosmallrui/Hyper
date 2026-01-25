package types

import "Hyper/models"

type GetFollowingFeedResponse struct {
	Following  []*models.FollowingQueryResult `json:"following"`
	NextCursor int64                          `json:"next_cursor"` // 返回给前端，下次请求带上
	HasMore    bool                           `json:"has_more"`    // 告诉前端是否还有更多
}
