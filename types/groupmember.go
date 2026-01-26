package types

// 邀请成员请求
// InviteMemberRequest 邀请成员请求体
type InviteMemberRequest struct {
	GroupId        int   `json:"group_id" binding:"required"`
	InvitedUserIds []int `json:"invited_user_ids" binding:"required,min=1"`
}

// InviteMemberResponse 邀请成员响应体
type InviteMemberResponse struct {
	SuccessCount  int   `json:"success_count"`   // 成功入群的人数
	FailedCount   int   `json:"failed_count"`    // 处理失败的人数
	FailedUserIds []int `json:"failed_user_ids"` // 失败的用户ID列表（可选，方便前端展示）
}

// 踢出成员请求
type KickMemberRequest struct {
	GroupId      int `json:"group_id" binding:"required"`
	KickedUserId int `json:"kicked_user_id" binding:"required"`
}
type KickmemberResponse struct {
	Success bool `json:"success"`
}
