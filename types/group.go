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

// 解散群请求
type DismissGroupRequest struct {
	GroupId int `json:"group_id" binding:"required"` // 群ID，必填
}

// 开启全员禁言请求
type MuteAllRequest struct {
	GroupId int `json:"group_id" binding:"required"` // 群ID，必填
}
type UnMuteAllRequest struct {
	GroupId int `json:"group_id" binding:"required"` // 群ID，必填
}

type UpdateGroupNameRequest struct {
	GroupId int    `json:"group_id" binding:"required"` // 群ID，必填
	Name    string `json:"name" binding:"required,min=1,max=20"`
}
type UpdateGroupAvatarRequest struct {
	GroupId int    `json:"group_id" binding:"required"` // 群ID，必填
	Avatar  string `json:"avatar" binding:"required"`
}
type UpdateGroupDescriptionRequest struct {
	GroupId     int    `json:"group_id" binding:"required"` // 群ID，必填
	Description string `json:"description" binding:"required,max=100"`
}
