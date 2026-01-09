package models

import "time"

type Group struct {
	Id          int       `gorm:"column:id;primary_key;AUTO_INCREMENT" json:"id"` // 自增ID
	Name        string    `gorm:"column:name;" json:"name"`                       // 群组名称
	Avatar      string    `gorm:"column:avatar;" json:"avatar"`                   // 群组头像
	OwnerId     int       `gorm:"column:owner_id;" json:"owner_id"`               // 群主用户ID
	Description string    `gorm:"column:description" json:"description"`          // 群组描述
	MemberCount int       `gorm:"column:member_count;" json:"member_count"`       // 群成员数量
	MaxMembers  int       `gorm:"column:max_members;" json:"max_members"`         // 最大成员数量
	CreatedAt   time.Time `gorm:"column:created_at;" json:"created_at"`           // 创建时间
	UpdatedAt   time.Time `gorm:"column:updated_at;" json:"updated_at"`           // 更新时间
}

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

// 邀请成员请求
type InviteMemberRequest struct {
	GroupId int   `json:"group_id" binding:"required"`
	UserIds []int `json:"user_ids" binding:"required"`
}

type InviteMemberResponse struct {
	SuccessCount  int   `json:"success_count"`
	FailedCount   int   `json:"failed_count"`
	FailedUserIds []int `json:"failed_user_ids_user_ids"`
}

// 踢出成员请求
type KickMemberRequest struct {
	GroupId int `json:"group_id" binding:"required"`
	UserId  int `json:"user_id" binding:"required"`
}

// 转移群主请求
type TransferOwnerRequest struct {
	GroupId    int `json:"group_id" binding:"required"`
	NewOwnerId int `json:"new_owner_id" binding:"required"`
}

// 更新群信息请求
type UpdateGroupRequest struct {
	GroupId     int    `json:"group_id" binding:"required"`
	Name        string `json:"name" binding:"omitempty,min=1,max=100"`
	Avatar      string `json:"avatar" binding:"omitempty"`
	Description string `json:"description" binding:"omitempty,max=500"`
}

// 删除群请求
type DeleteGroupRequest struct {
	GroupId int `json:"group_id" binding:"required"`
}

// 退群请求
type QuitGroupRequest struct {
	GroupId int `json:"group_id" binding:"required"`
}
