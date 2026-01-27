package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Party 派对主表
type Party struct {
	ID           int64   `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       int     `gorm:"not null;index:idx_user_id" json:"user_id"`
	Title        string  `gorm:"size:255;not null" json:"title"`
	Type         string  `gorm:"size:50;not null" json:"type"`
	Description  string  `gorm:"type:text" json:"description"`
	LocationName string  `gorm:"size:255;not null" json:"location_name"`
	Address      string  `gorm:"size:500" json:"address"`
	Latitude     float64 `gorm:"type:decimal(10,7);not null" json:"lat"`
	Longitude    float64 `gorm:"type:decimal(10,7);not null" json:"lng"`

	// geo_point 是数据库生成的虚拟列
	// gorm:"->" 表示只读（只在查询时加载，不参与写入/更新）
	GeoPoint string `gorm:"->;column:geo_point" json:"-"`

	Price        int64      `gorm:"default:0;not null" json:"price"` // 分
	MaxAttendees int64      `gorm:"default:0;not null" json:"max_attendees"`
	StartTime    time.Time  `gorm:"not null;index:idx_search_composite" json:"start_time"`
	EndTime      *time.Time `json:"end_time"`

	CoverImage string `gorm:"size:512" json:"cover_image"`
	ImagesJSON string `gorm:"column:images_json;type:json" json:"images"` // 对应数据库 JSON 类型

	Status      int8 `gorm:"default:1;not null;index:idx_search_composite" json:"status"`
	AuditStatus int8 `gorm:"default:0;not null;index:idx_search_composite" json:"audit_status"`

	AttendeeCount int64 `gorm:"default:0;not null" json:"attendee_count"`
	ViewCount     int64 `gorm:"default:0;not null" json:"view_count"`
	LikeCount     int64 `gorm:"default:0;not null" json:"like_count"`
	CommentCount  int64 `gorm:"default:0;not null" json:"comment_count"`

	HotScore  int64  `gorm:"default:0;not null;index:idx_hot_score" json:"hot_score"`
	RankLabel string `gorm:"size:100" json:"rank_label"`

	CreatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index:idx_deleted_at" json:"-"`
	User      Users          `gorm:"foreignKey:UserID" json:"user"`

	// 关联对象 (可选)
	Tags []PartyTag `gorm:"foreignKey:PartyID" json:"tags"`
}

// PartyAttendee 报名记录表
type PartyAttendee struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	PartyID   int64     `gorm:"not null;uniqueIndex:uk_party_user" json:"party_id"`
	UserID    int       `gorm:"not null;uniqueIndex:uk_party_user" json:"user_id"`
	Status    int8      `gorm:"default:1;not null" json:"status"`
	PayStatus int8      `gorm:"default:0;not null" json:"pay_status"`
	OrderNo   string    `gorm:"size:64" json:"order_no"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}

// PartyTag 派对标签表
type PartyTag struct {
	ID      int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	PartyID int64  `gorm:"not null;uniqueIndex:uk_party_tag" json:"party_id"`
	TagName string `gorm:"size:50;not null;uniqueIndex:uk_party_tag" json:"tag_name"`
}

// StringArray 字符串数组类型（用于存储图片列表）
type StringArray []string

func (s StringArray) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, s)
}

// PartyLike 派对点赞
type PartyLike struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	PartyID   int64     `json:"party_id" gorm:"not null;index"`
	UserID    int       `json:"user_id" gorm:"not null;index"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (Party) TableName() string {
	return "parties"
}

func (PartyTag) TableName() string {
	return "party_tags"
}

func (PartyAttendee) TableName() string {
	return "party_attendees"
}

func (PartyLike) TableName() string {
	return "party_likes"
}

// PartyListItem 派对列表项（前端展示格式）
type PartyListItem struct {
	ID       int64    `json:"id,string"` // 派对ID (int64转string，防止前端精度丢失)
	Title    string   `json:"title"`     // 标题
	Type     string   `json:"type"`      // 类型：派对/夜店/复古
	Location string   `json:"location"`  // 对应数据库 location_name
	Distance string   `json:"distance"`  // 计算后的距离描述，如 "1.2km" 或 "500m"
	Price    string   `json:"price"`     // 格式化后的价格，如 "99.0" 或 "免费"
	Lat      float64  `json:"lat"`       // 纬度 (用于地图定位)
	Lng      float64  `json:"lng"`       // 经度
	Tags     []string `json:"tags"`      // 标签名称数组
	Tag      string   `json:"tag"`

	// 主办方信息
	User       string `json:"user"`       // 主办方昵称
	UserAvatar string `json:"userAvatar"` // 主办方头像
	Fans       string `json:"fans"`       // 主办方粉丝数 (通常需要关联查询)
	IsVerified bool   `json:"isVerified"` // 是否为认证主办方

	Time         string `json:"time"`         // 格式化后的开始时间 (2026.01.27)
	DynamicCount int64  `json:"dynamicCount"` // 对应数据库 comment_count
	Attendees    int64  `json:"attendees"`    // 已报名人数
	IsFull       bool   `json:"isFull"`       // 是否已满员 (attendee_count >= max_attendees)

	Rank  string `json:"rank"`  // 榜单标签，如 "热门榜No.1"
	Image string `json:"image"` // 封面图 cover_image

	// 针对当前登录用户的状态
	IsLiked     bool `json:"isLiked"`     // 是否已点赞
	IsAttending bool `json:"isAttending"` // 是否已报名
}
