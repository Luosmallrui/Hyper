package types

// 用户通知类型（user_notifications.type）
const (
	NotifyTypeSystem      = "system"      // 系统消息
	NotifyTypeInteraction = "interaction" // 互动通知（点赞/评论/收藏/关注）
	NotifyTypePayment     = "payment"     // 支付消息（支付/退款/核销）
)

// NotificationItem 通知列表项
type NotificationItem struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Payload   string `json:"payload"` // JSON 字符串，跳转参数，如 {"note_id":123}
	IsRead    bool   `json:"is_read"`
	CreatedAt string `json:"created_at"`
}

// UnreadCountResponse 未读数（按类型分组）
type UnreadCountResponse struct {
	Total       int64 `json:"total"`
	System      int64 `json:"system"`
	Interaction int64 `json:"interaction"`
	Payment     int64 `json:"payment"`
}

// NotificationListQuery 通知列表查询参数
type NotificationListQuery struct {
	Type string `form:"type"` // 可选：system / interaction / payment，缺省全部
	Page int    `form:"page"`
	Size int    `form:"size"`
}

// MarkNotificationReadRequest 标记已读请求
type MarkNotificationReadRequest struct {
	IDs []int64 `json:"ids" binding:"required"`
}

// NotificationPayload 用户通知的 MQ 实时推送载荷
//（topic = SystemMessageTopic，SystemMessage.Type = "user_notification"）
type NotificationPayload struct {
	UserID    int64  `json:"user_id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Payload   string `json:"payload"`
	CreatedAt string `json:"created_at"`
}
