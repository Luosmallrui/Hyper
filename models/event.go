package models

import "time"

// Event 活动主表
type Event struct {
	ID              int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          int       `gorm:"not null;index:idx_event_user_id" json:"user_id"`
	Title           string    `gorm:"size:80;not null" json:"title"`
	ShareTitle      string    `gorm:"size:20;not null" json:"share_title"`
	StartAt         time.Time `gorm:"not null" json:"start_at"`
	EndAt           time.Time `gorm:"not null" json:"end_at"`
	Summary         string    `gorm:"type:text" json:"summary"`
	SummaryHTML     string    `gorm:"type:text;column:summary_html" json:"summary_html"`
	StrongRealName  bool      `gorm:"default:false" json:"strong_real_name"`
	MinorProtection bool      `gorm:"default:false" json:"minor_protection"`
	// 地址信息
	AddressMode string  `gorm:"size:20;default:default" json:"address_mode"`
	Province    string  `gorm:"size:50" json:"province"`
	City        string  `gorm:"size:50" json:"city"`
	District    string  `gorm:"size:50" json:"district"`
	Location    string  `gorm:"size:255" json:"location"`
	// 海报
	DetailPoster     string `gorm:"type:text;column:detail_poster" json:"detail_poster"`
	DetailLongPoster string `gorm:"type:text;column:detail_long_poster" json:"detail_long_poster"`
	SharePoster      string `gorm:"type:text;column:share_poster" json:"share_poster"`
	GroupPoster      string `gorm:"type:text;column:group_poster" json:"group_poster"`
	// 资质
	PromoterConfirmed bool   `gorm:"default:false" json:"promoter_confirmed"`
	SafetyAgreement   bool   `gorm:"default:false" json:"safety_agreement"`
	TicketStatement   bool   `gorm:"default:false" json:"ticket_statement"`
	Notes             string `gorm:"type:text" json:"notes"`
	// 状态
	Status    string    `gorm:"size:20;default:draft" json:"status"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (Event) TableName() string {
	return "events"
}

// EventTicket 票券表
type EventTicket struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	EventID       int64     `gorm:"not null;index:idx_ticket_event_id" json:"event_id"`
	Name          string    `gorm:"size:50;not null" json:"name"`
	Price         int64     `gorm:"not null" json:"price"` // 单位：分
	Stock         int       `gorm:"not null" json:"stock"`
	PurchaseLimit int       `gorm:"not null;default:1" json:"purchase_limit"`
	StartAt       time.Time `gorm:"not null" json:"start_at"`
	EndAt         time.Time `gorm:"not null" json:"end_at"`
	Description   string    `gorm:"type:text" json:"description"`
	CreatedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (EventTicket) TableName() string {
	return "event_tickets"
}

// EventListItem 活动列表项（前端展示）
type EventListItem struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	ShareTitle  string    `json:"share_title"`
	SharePoster string    `json:"share_poster"`
	StartAt     time.Time `json:"start_at"`
	EndAt       time.Time `json:"end_at"`
	Status      string    `json:"status"`
	City        string    `json:"city"`
	Location    string    `json:"location"`
	MinPrice    int64     `json:"min_price"` // 最低票价
	CreatedAt   time.Time `json:"created_at"`
}
