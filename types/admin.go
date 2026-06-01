package types

import "Hyper/models"

// AdminLoginRequest 管理员登录请求
type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AdminPartyListResponse 派对列表响应
type AdminPartyListResponse struct {
	List     []AdminPartyItem `json:"list"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
}

// AdminPartyItem 管理后台派对列表项
type AdminPartyItem struct {
	ID            int64  `json:"id"`
	UserID        int    `json:"user_id"`
	UserName      string `json:"user_name"`
	UserAvatar    string `json:"user_avatar"`
	UserMobile    string `json:"user_mobile"`
	Title         string `json:"title"`
	Type          string `json:"type"`
	Status        string `json:"status"`
	LocationName  string `json:"location_name"`
	Address       string `json:"address"`
	CoverImage    string `json:"cover_image"`
	Category      int    `json:"category"`
	AttendeeCount int    `json:"attendee_count"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// AdminPartyDetail 管理后台派对详情
type AdminPartyDetail struct {
	models.Merchant
	UserName   string `json:"user_name"`
	UserAvatar string `json:"user_avatar"`
	UserMobile string `json:"user_mobile"`
}

// AdminTicketItem 票券列表项
type AdminTicketItem struct {
	models.EventTicket
	EventTitle string `json:"event_title"`
}

// AdminTicketListResponse 票券列表响应
type AdminTicketListResponse struct {
	List     []AdminTicketItem `json:"list"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
}

// AdminOrderItem 订单列表项
type AdminOrderItem struct {
	models.Order
	UserName   string `json:"user_name"`
	UserMobile string `json:"user_mobile"`
}

// AdminOrderListResponse 订单列表响应
type AdminOrderListResponse struct {
	List     []AdminOrderItem `json:"list"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
}

// AdminUpdatePartyStatusRequest 更新派对状态
type AdminUpdatePartyStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// AdminOrganizerListResponse 入驻申请列表响应
type AdminOrganizerListResponse struct {
	List     []AdminOrganizerItem `json:"list"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"pageSize"`
}

// AdminOrganizerItem 管理后台入驻申请列表项
type AdminOrganizerItem struct {
	ID             int64   `json:"id"`
	UserID         int64   `json:"user_id"`
	UserName       string  `json:"user_name"`
	UserAvatar     string  `json:"user_avatar"`
	UserMobile     string  `json:"user_mobile"`
	Type           string  `json:"type"`
	Name           string  `json:"name"`
	Logo           string  `json:"logo"`
	Status         int8    `json:"status"`
	RejectReason   string  `json:"reject_reason"`
	Level          string  `json:"level"`
	ServiceFeeRate float64 `json:"service_fee_rate"`
	Province       string  `json:"province"`
	City           string  `json:"city"`
	District       string  `json:"district"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// AdminOrganizerDetail 管理后台入驻申请详情
type AdminOrganizerDetail struct {
	models.Organizer
	UserName   string `json:"user_name"`
	UserAvatar string `json:"user_avatar"`
	UserMobile string `json:"user_mobile"`
}

// AdminAuditOrganizerRequest 审核入驻申请
type AdminAuditOrganizerRequest struct {
	Status       int8   `json:"status" binding:"required,oneof=2 3"`
	RejectReason string `json:"reject_reason"`
}

// AdminWechatSubscribeRequest 管理员绑定微信订阅通知
type AdminWechatSubscribeRequest struct {
	Code string `json:"code" binding:"required"`
}

// AdminDashboardStats 后台统计
type AdminDashboardStats struct {
	TotalParties int64 `json:"total_parties"`
	TotalEvents  int64 `json:"total_events"`
	TotalTickets int64 `json:"total_tickets"`
	TotalOrders  int64 `json:"total_orders"`
	TotalUsers   int64 `json:"total_users"`
	TotalRevenue int64 `json:"total_revenue"` // 单位：分
}
