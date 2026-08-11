package types

import (
	"Hyper/models"
	"time"
)

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
	UserName   string           `json:"user_name"`
	UserAvatar string           `json:"user_avatar"`
	UserMobile string           `json:"user_mobile"`
	TagIDs     []int64          `json:"tag_ids"`
	Tags       []ContentTagItem `json:"tags"`
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

type AdminPlatformFlowFilter struct {
	Type        string     `json:"type"`
	Keyword     string     `json:"keyword"`
	OrganizerID int64      `json:"organizer_id"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
}

type AdminPlatformFlowItem struct {
	ID            int64     `json:"id"`
	FlowNo        string    `json:"flow_no"`
	Type          string    `json:"type"`
	Direction     string    `json:"direction"`
	Amount        int64     `json:"amount"`
	OrderNo       string    `json:"order_no"`
	RefundNo      string    `json:"refund_no"`
	WithdrawID    int64     `json:"withdraw_id"`
	OrganizerID   int64     `json:"organizer_id"`
	OrganizerName string    `json:"organizer_name"`
	PayMethod     string    `json:"pay_method"`
	Remark        string    `json:"remark"`
	OccurredAt    time.Time `json:"occurred_at"`
}

type AdminPlatformFlowListResponse struct {
	List     []AdminPlatformFlowItem `json:"list"`
	Total    int64                   `json:"total"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"pageSize"`
}

// AdminActivityListResponse 活动审核列表响应
type AdminActivityListResponse struct {
	List     []AdminActivityItem `json:"list"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"pageSize"`
}

// AdminActivityItem 管理后台活动列表项
type AdminActivityItem struct {
	ID               int64   `json:"id"`
	OrganizerID      int64   `json:"organizer_id"`
	OrganizerName    string  `json:"organizer_name"`
	OrganizerType    string  `json:"organizer_type"`
	Type             string  `json:"type"`
	Name             string  `json:"name"`
	ShareTitle       string  `json:"share_title"`
	StartTime        string  `json:"start_time"`
	EndTime          string  `json:"end_time"`
	RealNameMode     int8    `json:"real_name_mode"`
	MinorCheck       int8    `json:"minor_check"`
	Description      string  `json:"description"`
	Province         string  `json:"province"`
	City             string  `json:"city"`
	District         string  `json:"district"`
	Address          string  `json:"address"`
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
	PosterDetail     string  `json:"poster_detail"`
	PosterLong       string  `json:"poster_long"`
	PosterList       string  `json:"poster_list"`
	PosterWechat     string  `json:"poster_wechat"`
	QualificationDoc string  `json:"qualification_doc"`
	Status           int8    `json:"status"`
	RejectReason     string  `json:"reject_reason"`
	IsHidden         int8    `json:"is_hidden"`
	HiddenAt         string  `json:"hidden_at"`
	HiddenReason     string  `json:"hidden_reason"`
	TicketSpecCount  int     `json:"ticket_spec_count"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

// AdminActivityDetail 管理后台活动详情
type AdminActivityDetail struct {
	models.Activity
	Organizer   *models.Organizer   `json:"organizer,omitempty"`
	TicketSpecs []models.TicketSpec `json:"ticket_specs"`
	TagIDs      []int64             `json:"tag_ids"`
	Tags        []ContentTagItem    `json:"tags"`
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

type AdminTicketOrderListResponse struct {
	List     []AdminTicketOrderItem `json:"list"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"pageSize"`
}

type AdminTicketOrderItem struct {
	OrderNo        string `json:"order_no"`
	Status         int8   `json:"status"`
	TotalPrice     int64  `json:"total_price"`
	ActualPrice    int64  `json:"actual_price"`
	PointsAmount   int64  `json:"points_amount"`
	PointsDiscount int64  `json:"points_discount"`
	Quantity       int    `json:"quantity"`
	UserID         int64  `json:"user_id"`
	UserName       string `json:"user_name"`
	UserMobile     string `json:"user_mobile"`
	BuyerName      string `json:"buyer_name"`
	BuyerIDCard    string `json:"buyer_id_card"`
	ActivityID     int64  `json:"activity_id"`
	ActivityName   string `json:"activity_name"`
	TicketSpecID   int64  `json:"ticket_spec_id"`
	TicketSpecName string `json:"ticket_spec_name"`
	PayMethod      string `json:"pay_method"`
	SalesChannel   string `json:"sales_channel"`
	PayTime        string `json:"pay_time"`
	ExpireTime     string `json:"expire_time"`
	CreatedAt      string `json:"created_at"`
}

type AdminTicketOrderDetail struct {
	Order               AdminTicketOrderItem `json:"order"`
	Viewers             []OrderViewerItem    `json:"viewers"`
	Refunds             []models.Refund      `json:"refunds"`
	RefundLogs          []models.RefundLog   `json:"refund_logs"`
	VerificationRecords []map[string]any     `json:"verification_records"`
	PayRecords          []models.PayRecord   `json:"pay_records"`
}

type AdminRefundDetail struct {
	Refund              models.Refund        `json:"refund"`
	Order               AdminTicketOrderItem `json:"order"`
	Viewers             []OrderViewerItem    `json:"viewers"`
	RefundLogs          []models.RefundLog   `json:"refund_logs"`
	VerificationRecords []map[string]any     `json:"verification_records"`
	PayRecords          []models.PayRecord   `json:"pay_records"`
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
	Enabled        int8    `json:"enabled"`
	RejectReason   string  `json:"reject_reason"`
	Level          string  `json:"level"`
	ServiceFeeRate float64 `json:"service_fee_rate"`
	Province       string  `json:"province"`
	City           string  `json:"city"`
	District       string  `json:"district"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

type AdminActivityFilter struct {
	Status        *int8
	IsHidden      *int8
	Keyword       string
	OrganizerID   int64
	PublishedFrom *time.Time
	PublishedTo   *time.Time
	ActivityFrom  *time.Time
	ActivityTo    *time.Time
}

// AdminOrganizerDetail 管理后台入驻申请详情
type AdminOrganizerDetail struct {
	models.Organizer
	UserName   string           `json:"user_name"`
	UserAvatar string           `json:"user_avatar"`
	UserMobile string           `json:"user_mobile"`
	TagIDs     []int64          `json:"tag_ids"`
	Tags       []ContentTagItem `json:"tags"`
}

// AdminAuditOrganizerRequest 审核入驻申请
type AdminAuditOrganizerRequest struct {
	Status       int8   `json:"status" binding:"required,oneof=2 3"`
	RejectReason string `json:"reject_reason"`
}

// AdminAuditActivityRequest 审核活动
type AdminAuditActivityRequest struct {
	Status       int8   `json:"status" binding:"required,oneof=3 4"`
	RejectReason string `json:"reject_reason"`
}

// AdminActivityVisibilityRequest controls whether an approved activity is
// publicly visible. Hidden activities remain available to operations and
// historical order, refund and verification queries.
type AdminActivityVisibilityRequest struct {
	Visible *bool  `json:"visible" binding:"required"`
	Reason  string `json:"reason"`
}

// AdminWechatSubscribeRequest 管理员绑定微信订阅通知
type AdminWechatSubscribeRequest struct {
	Code string `json:"code" binding:"required"`
}

// AdminDashboardStats 后台统计
type AdminDashboardStats struct {
	TotalMerchants int64 `json:"total_merchants"`
	TotalParties   int64 `json:"total_parties"`
	TotalEvents    int64 `json:"total_events"`
	TotalTickets   int64 `json:"total_tickets"`
	TotalOrders    int64 `json:"total_orders"`
	TotalUsers     int64 `json:"total_users"`
	TotalRevenue   int64 `json:"total_revenue"` // 单位：分
}

type AdminFinanceSummary struct {
	TotalRevenue  int64 `json:"total_revenue"`
	PendingSettle int64 `json:"pending_settle"`
	SettledAmount int64 `json:"settled_amount"`
	RefundAmount  int64 `json:"refund_amount"`
	OrderCount    int64 `json:"order_count"`
}

type AdminSettlementItem struct {
	OrganizerID           int64  `json:"organizer_id"`
	OrganizerName         string `json:"organizer_name"`
	GrossAmount           int64  `json:"gross_amount"`
	RefundAmount          int64  `json:"refund_amount"`
	WithdrawAmount        int64  `json:"withdraw_amount"`
	PendingWithdrawAmount int64  `json:"pending_withdraw_amount"`
	SettleAmount          int64  `json:"settle_amount"`
	PendingSettleAmount   int64  `json:"pending_settle_amount"`
	OrderCount            int64  `json:"order_count"`
	UpdatedAt             string `json:"updated_at"`
}

type AdminSettlementListResponse struct {
	List     []AdminSettlementItem `json:"list"`
	Total    int64                 `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"pageSize"`
}

type AdminUserItem struct {
	ID          int    `json:"id"`
	Nickname    string `json:"nickname"`
	Avatar      string `json:"avatar"`
	Mobile      string `json:"mobile"`
	Status      int8   `json:"status"`
	TotalAmount int64  `json:"total_amount"`
	OrderCount  int64  `json:"order_count"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type AdminUserListResponse struct {
	List     []AdminUserItem `json:"list"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
}

type AdminUpdateUserStatusRequest struct {
	Status int8 `json:"status" binding:"required,oneof=0 1"`
}

type AdminBannerRequest struct {
	Title    string `json:"title"`
	Image    string `json:"image" binding:"required"`
	Link     string `json:"link"`
	Position string `json:"position"`
	Sort     int    `json:"sort"`
	Status   int8   `json:"status"`
}

type AdminBannerSortItem struct {
	ID   int64 `json:"id" binding:"required"`
	Sort int   `json:"sort"`
}

type AdminBannerSortRequest struct {
	List []AdminBannerSortItem `json:"list" binding:"required"`
}

type AdminSettingItem struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Remark string `json:"remark"`
}

type AdminSettingsRequest struct {
	Settings []AdminSettingItem `json:"settings" binding:"required"`
}

// AdminSystemConfig is the structured contract for the PC system configuration page.
type AdminSystemConfig struct {
	SystemName            string `json:"system_name"`
	ICPRecordNo           string `json:"icp_record_no"`
	CustomerServicePhone  string `json:"customer_service_phone"`
	CustomerServiceWechat string `json:"customer_service_wechat"`
	CustomerServiceEmail  string `json:"customer_service_email"`
	CustomerServiceHours  string `json:"customer_service_hours"`
	WithdrawArrivalCycle  string `json:"withdraw_arrival_cycle"`
}

// PublicSystemConfig is safe to expose before login, for example on the admin login page.
type PublicSystemConfig struct {
	SystemName  string `json:"system_name"`
	ICPRecordNo string `json:"icp_record_no"`
}
