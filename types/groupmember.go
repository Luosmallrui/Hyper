package types

// 邀请成员请求
type InviteMemberRequest struct {
	GroupId int   `json:"group_id" binding:"required"`
	UserIds []int `json:"user_ids" binding:"required"`
}

type InviteMemberResponse struct {
	SuccessCount  int   `json:"success_count"`
	FailedCount   int   `json:"failed_count"`
	FailedUserIds []int `json:"failed_user_ids"`
}

// 踢出成员请求
type KickMemberRequest struct {
	GroupId      int `json:"group_id" binding:"required"`
	KickedUserId int `json:"kicked_user_id" binding:"required"`
}
type KickmemberResponse struct {
	Success bool `json:"success"`
}

// 群成员列表元素（DTO）
type GroupMemberItemDTO struct {
	UserId   int    `json:"user_id"`
	Avatar   string `json:"avatar"`
	Nickname string `json:"nickname"`
	Gender   int    `json:"gender"`
	Motto    string `json:"motto"`
	Role     int    `json:"role"`
	IsMute   int    `json:"is_mute"`
	UserCard string `json:"user_card"`
}

// 群成员列表响应（DTO）
type GroupMemberListResponse struct {
	Members []GroupMemberItemDTO `json:"members"`
}
