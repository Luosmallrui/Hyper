package models

import "time"

// Topic 话题表
type Topic struct {
	// 1. 基础信息
	// 显式关闭自增，配合你手动生成的唯一 ID
	ID          uint64 `gorm:"primaryKey;autoIncrement:false" json:"id"`
	Name        string `gorm:"type:varchar(64);uniqueIndex:idx_topics_name;not null" json:"name"`
	Description string `gorm:"type:varchar(255);default:''" json:"description"`
	CoverURL    string `gorm:"type:varchar(255);default:''" json:"cover_url"`

	// 2. 归属与溯源
	CreatorID  uint64 `gorm:"index" json:"creator_id"`  // 谁创建了这个话题
	CategoryID uint32 `gorm:"index" json:"category_id"` // 所属大类

	// 3. 统计数据 (建议使用 uint32 或 uint64)
	PostCount   uint32 `gorm:"default:0" json:"post_count"`
	ViewCount   uint32 `gorm:"default:0" json:"view_count"`
	FollowCount uint32 `gorm:"default:0" json:"follow_count"` // 新增：关注该话题的人数

	// 4. 运营权重
	IsHot      bool  `gorm:"default:false;index" json:"is_hot"`
	SortWeight int32 `gorm:"default:0" json:"sort_weight"`

	// 5. 状态与审计
	Status int8 `gorm:"type:tinyint;default:1" json:"status"` // 1正常, 0隐藏, -1封禁

	// 6. 时间戳
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	LastPostAt time.Time `gorm:"index" json:"last_post_at"` // 新增：该话题下最后一次发帖时间

}

func (Topic) TableName() string {
	return "topics"
}

// NoteTopic 笔记与话题的中间表
type NoteTopic struct {
	ID uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	// 联合唯一索引：确保 (note_id, topic_id) 组合唯一
	NoteID  uint64 `gorm:"uniqueIndex:uk_note_topic;not null" json:"note_id"`
	TopicID uint64 `gorm:"uniqueIndex:uk_note_topic;not null;index:idx_topic_id" json:"topic_id"`

	CreatedAt time.Time `json:"created_at"`
}

func (NoteTopic) TableName() string {
	return "note_topics"
}
