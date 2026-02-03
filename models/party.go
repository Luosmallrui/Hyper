package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Merchant 派对主表
type Merchant struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      int    `gorm:"not null;index:idx_user_id" json:"user_id"`
	Title       string `gorm:"size:255;not null" json:"title"`
	Type        string `gorm:"size:50;not null" json:"type"`
	Description string `gorm:"type:text" json:"description"`
	CoverImage  string `gorm:"type:text" json:"cover_image"`

	LocationName string    `gorm:"size:255" json:"location_name"`
	Address      string    `gorm:"size:500" json:"address"`
	Latitude     float64   `gorm:"type:decimal(10,7);not null" json:"lat"`
	Longitude    float64   `gorm:"type:decimal(10,7);not null" json:"lng"`
	CreatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	ImagesJSON   string    `gorm:"type:json;column:images_json" json:"images"`
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
func (Merchant) TableName() string {
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

// MerchantListItem 派对列表项（前端展示格式）
type MerchantListItem struct {
	ID       int64   `json:"id"`       // 派对 ID
	Title    string  `json:"title"`    // 标题
	Type     string  `json:"type"`     // 类型：派对/场地
	Location string  `json:"location"` //位置
	Lat      float64 `json:"lat"`      // 纬度
	Lng      float64 `json:"lng"`      // 经度

	UserName     string    `json:"username"`
	UserAvatar   string    `json:"user_avatar"`
	CoverImage   string    `json:"cover_image"`   //封面图片
	CreatedAt    time.Time `json:"created_at"`    //创建时间
	AvgPrice     int64     `json:"avg_price"`     //人均价格
	CurrentCount int64     `json:"current_count"` // 参与人数
	PostCount    int64     `json:"post_count"`
}
