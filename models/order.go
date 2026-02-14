package models

import (
	"time"

	"gorm.io/datatypes"
)

// Order 订单主表
type Order struct {
	ID          int        `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      int        `gorm:"column:user_id;not null;index:idx_user_id" json:"user_id"`
	OrderSn     string     `gorm:"column:order_sn;type:varchar(32);not null;uniqueIndex:idx_order_sn" json:"order_sn"`
	TotalAmount uint64     `gorm:"column:total_amount;not null" json:"total_amount"` // 单位：分
	Description string     `gorm:"column:description;type:varchar(255)" json:"description"`
	Status      int8       `gorm:"column:status;not null;default:10" json:"status"` // 10:待支付, 20:已支付...
	PaidAt      *time.Time `gorm:"column:paid_at" json:"paid_at"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	//DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"` // 如果需要软删除
}

func (Order) TableName() string {
	return "orders"
}

// PayRecord 支付流水记录表
type PayRecord struct {
	ID            uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderSn       string         `gorm:"column:order_sn;type:varchar(32);not null;uniqueIndex:idx_order_sn" json:"order_sn"`
	PayPlatform   int8           `gorm:"column:pay_platform;not null;default:1" json:"pay_platform"` // 1:微信, 2:支付宝
	PayMethod     string         `gorm:"column:pay_method;type:varchar(20)" json:"pay_method"`       // JSAPI, APP...
	TransactionId string         `gorm:"column:transaction_id;type:varchar(64);index:idx_transaction_id" json:"transaction_id"`
	OutRequestNo  string         `gorm:"column:out_request_no;type:varchar(64)" json:"out_request_no"` // prepay_id
	PayerId       string         `gorm:"column:payer_id;type:varchar(128)" json:"payer_id"`            // openid/buyer_id
	AmountTotal   uint64         `gorm:"column:amount_total;not null;default:0" json:"amount_total"`
	Currency      string         `gorm:"column:currency;type:varchar(10);default:'CNY'" json:"currency"`
	PayStatus     int8           `gorm:"column:pay_status;not null;default:0" json:"pay_status"`
	RawTradeState string         `gorm:"column:raw_trade_state;type:varchar(32)" json:"raw_trade_state"`
	NotifyRaw     datatypes.JSON `gorm:"column:notify_raw" json:"notify_raw"` // 使用 gorm.io/datatypes 处理 JSON
	FinishedAt    *time.Time     `gorm:"column:finished_at" json:"finished_at"`
	CreatedAt     time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (PayRecord) TableName() string {
	return "pay_records"
}

// Viewer 观影人表
type Viewer struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int       `gorm:"not null;index" json:"user_id"`                                // 所属用户ID（关联 users 表）
	RealName  string    `gorm:"size:50;not null" json:"real_name"`                            // 真实姓名
	IDCard    string    `gorm:"type:char(18);not null;uniqueIndex:uk_id_card" json:"id_card"` // 身份证号
	Phone     string    `gorm:"type:char(11);not null;uniqueIndex:uk_phone" json:"phone"`     // 手机号码
	Type      int8      `gorm:"default:1" json:"type"`                                        // 类型：1-成人，2-儿童，3-老人
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Viewer) TableName() string {
	return "viewers"
}
