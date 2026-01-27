package types

import "time"

type PointRecord struct {
	ID          int       `json:"id"`
	Amount      int       `json:"amount"`
	Description string    `json:"description"`
	OrderType   string    `json:"order_type"`
	CreatedAt   time.Time `json:"created_at"`
	Status      int       `json:"status"`
}

type ListPointsRecord struct {
	Records  []PointRecord `json:"records"`
	Total    int           `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
	HasMore  bool          `json:"has_more"`
}

type PointsAccount struct {
	Balance       int `json:"balance"`        // 当前积分余额
	TotalEarned   int `json:"total_earned"`   // 累计获得
	TotalUsed     int `json:"total_used"`     // 累计使用
	PendingCount  int `json:"pending_count"`  // 待入账订单数
	PendingAmount int `json:"pending_amount"` // 待入账积分
}
