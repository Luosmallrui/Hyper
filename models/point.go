package models

import "time"

type UserPoint struct {
	ID          uint64    `gorm:"primaryKey;column:id"` // 注意大小写匹配
	UserID      uint64    `gorm:"column:user_id;uniqueIndex"`
	Balance     int64     `gorm:"column:balance;default:0"`
	TotalEarned uint64    `gorm:"column:total_earned;default:0"`
	TotalUsed   uint64    `gorm:"column:total_used;default:0"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (UserPoint) TableName() string {
	return "user_points"
}

// 积分变动类型常量定义
const (
	// 收入类
	TypeActivityReward   = 1 // 活动奖励
	TypeSignReward       = 2 // 签到奖励
	TypeOrderRefund      = 3 // 订单退款返还
	TypeSystemCompensate = 4 // 系统/人工补偿

	// 支出类
	TypeExchange    = 10 // 积分兑换
	TypeOrderDeduct = 11 // 购物抵扣
)

type PointsLog struct {
	ID         uint64    `gorm:"primaryKey;column:id"`
	UserID     uint64    `gorm:"column:user_id;index:idx_user_id"`
	Amount     int64     `gorm:"column:amount"`      // 变动数额（正负）
	Balance    int64     `gorm:"column:balance"`     // 变动后余额
	ChangeType int8      `gorm:"column:change_type"` // 1-签到, 2-消费返利, 3-积分抵扣, 4-后台调整
	Status     int8      `gorm:"column:status"`      // 0-待入账, 1-正式入账
	SourceID   string    `gorm:"column:source_id;index:idx_source_id;size:64"`
	Remark     string    `gorm:"column:remark;size:255"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
}

// TableName 指定表名，匹配你建表语句中的 point_logs
func (PointsLog) TableName() string {
	return "point_logs"
}
