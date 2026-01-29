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
	IsMuteAll   int       `gorm:"column:is_mute_all" json:"is_mute_all"`          //是否全员禁言[0否;1是]
	IsDismiss   int       `gorm:"column:is_dismiss;default:0"`                    // 新增：0正常，1解散
	IsAllMute   int       `gorm:"column:is_all_mute;default:0" json:"is_all_mute"`
}

func (Group) TableName() string {
	return "groups"
}
