package types

// 创建群请求
type CreateGroupRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Avatar      string `json:"avatar" binding:"omitempty"`
	Description string `json:"description" binding:"omitempty,max=500"`
}

// 创建群响应
type CreateGroupResponse struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	OwnerId     int    `json:"owner_id"`
	MemberCount int    `json:"member_count"`
	CreatedAt   string `json:"created_at"`
}
