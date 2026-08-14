package types

import (
	"Hyper/models"
	"time"
)

type PageResponse[T any] struct {
	List  []T   `json:"list"`
	Total int64 `json:"total"`
}

type ActivityListFilter struct {
	Keyword       string
	Status        *int8
	PublishedFrom *time.Time
	PublishedTo   *time.Time
	ActivityFrom  *time.Time
	ActivityTo    *time.Time
}

type OrganizerApplyRequest struct {
	Name     string `json:"name" binding:"required"`
	Type     string `json:"type"`
	Logo     string `json:"logo"`
	Province string `json:"province"`
	City     string `json:"city"`
	District string `json:"district"`
}

type OrganizerApplyResponse struct {
	ApplicationID int64     `json:"application_id"`
	Status        int8      `json:"status"`
	SubmittedAt   time.Time `json:"submitted_at"`
}

type OrganizerAuditStatusResponse struct {
	OrganizerID  int64      `json:"organizer_id,omitempty"`
	Type         string     `json:"type,omitempty"`
	Status       int8       `json:"status"`
	Enabled      int8       `json:"enabled"`
	RejectReason string     `json:"reject_reason"`
	SubmittedAt  *time.Time `json:"submitted_at,omitempty"`
	ReviewedAt   *time.Time `json:"reviewed_at,omitempty"`
}

type OrganizerBasicRequest struct {
	Name     *string `json:"name"`
	Logo     *string `json:"logo"`
	Province *string `json:"province"`
	City     *string `json:"city"`
	District *string `json:"district"`
}

type OrganizerWithdrawRequest struct {
	BankAccountName string `json:"bank_account_name" binding:"required"`
	BankAccountNo   string `json:"bank_account_no" binding:"required"`
	BankName        string `json:"bank_name" binding:"required"`
	ContactName     string `json:"contact_name"`
	ContactPhone    string `json:"contact_phone"`
}

type OrganizerBankAuditInfo struct {
	ID              int64      `json:"id"`
	BankAccountName string     `json:"bank_account_name"`
	BankAccountNo   string     `json:"bank_account_no"`
	BankName        string     `json:"bank_name"`
	ContactName     string     `json:"contact_name"`
	ContactPhone    string     `json:"contact_phone"`
	Status          int8       `json:"status"`
	RejectReason    string     `json:"reject_reason"`
	ReviewedAt      *time.Time `json:"reviewed_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type OrganizerWithdrawInfoResponse struct {
	BankAccountName       string                  `json:"bank_account_name"`
	BankAccountNo         string                  `json:"bank_account_no"`
	BankName              string                  `json:"bank_name"`
	ContactName           string                  `json:"contact_name"`
	ContactPhone          string                  `json:"contact_phone"`
	CanWithdraw           bool                    `json:"can_withdraw"`
	GrossAmount           int64                   `json:"gross_amount"`
	RefundAmount          int64                   `json:"refund_amount"`
	WithdrawAmount        int64                   `json:"withdraw_amount"`
	PendingWithdrawAmount int64                   `json:"pending_withdraw_amount"`
	AvailableAmount       int64                   `json:"available_amount"`
	ArrivalCycle          string                  `json:"arrival_cycle"`
	PendingAudit          *OrganizerBankAuditInfo `json:"pending_audit,omitempty"`
	LatestAudit           *OrganizerBankAuditInfo `json:"latest_audit,omitempty"`
}

type CreateOrganizerWithdrawRequest struct {
	Amount int64  `json:"amount" binding:"required,min=1"`
	Remark string `json:"remark"`
}

type OrganizerCollectionRequest struct {
	Title       string  `json:"title" binding:"required"`
	ShareTitle  string  `json:"share_title"`
	Description string  `json:"description"`
	ShareImage  string  `json:"share_image"`
	Status      int8    `json:"status"`
	ActivityIDs []int64 `json:"activity_ids"`
}

type OrganizerCollectionDetail struct {
	ActivityCollectionItem
	ActivityIDs []int64            `json:"activity_ids"`
	Activities  []ActivityListItem `json:"activities"`
}

type OrganizerMessageItem struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	ContentType string     `json:"content_type"`
	CoverImage  string     `json:"cover_image"`
	Type        string     `json:"type"`
	Target      string     `json:"target"`
	IsRead      bool       `json:"is_read"`
	ReadAt      *time.Time `json:"read_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

type OrganizerMessageDetail struct {
	OrganizerMessageItem
	MediaData []string `json:"media_data"`
}

type OrganizerReadAllResponse struct {
	UpdatedCount int64 `json:"updated_count"`
}

type OrganizerSubscriptionSummary struct {
	TotalSubscriptions int64 `json:"total_subscriptions"`
	TodaySubscriptions int64 `json:"today_subscriptions"`
}

type VenueListItem struct {
	ID               int64            `json:"id"`
	ActivityID       int64            `json:"activity_id,omitempty"`
	ActivityName     string           `json:"activity_name,omitempty"`
	UserID           int64            `json:"user_id"`
	Name             string           `json:"name"`
	Logo             string           `json:"logo"`
	CoverImage       string           `json:"cover_image"`
	Description      string           `json:"description"`
	BusinessHours    string           `json:"business_hours"`
	ServicePhone     string           `json:"service_phone"`
	Province         string           `json:"province"`
	City             string           `json:"city"`
	District         string           `json:"district"`
	Address          string           `json:"address"`
	Latitude         float64          `json:"latitude"`
	Longitude        float64          `json:"longitude"`
	AverageSpend     int64            `json:"average_spend"`
	IsFollow         bool             `json:"is_follow"`
	IsSubscribe      bool             `json:"is_subscribe"`
	FollowCount      int64            `json:"follow_count"`
	FollowTargetType string           `json:"follow_target_type"`
	FollowTargetID   int64            `json:"follow_target_id"`
	SubscribeCount   int64            `json:"subscribe_count"`
	PostCount        int64            `json:"post_count"`
	TagIDs           []int64          `json:"tag_ids"`
	Tags             []ContentTagItem `json:"tags"`
	CreatedAt        time.Time        `json:"created_at"`
}

type VenueDetailResponse struct {
	VenueListItem
	Gallery []string                `json:"gallery"`
	Stores  []models.OrganizerStore `json:"stores"`
}

type VenueNoteItem struct {
	ID           int64       `json:"id,string"`
	UserID       int64       `json:"user_id"`
	Title        string      `json:"title"`
	Content      string      `json:"content"`
	Type         int         `json:"type"`
	MediaData    []NoteMedia `json:"media_data"`
	LikeCount    int64       `json:"like_count"`
	CollCount    int64       `json:"coll_count"`
	ShareCount   int64       `json:"share_count"`
	CommentCount int64       `json:"comment_count"`
	ViewCount    int64       `json:"view_count,omitempty"`
	ActivityID   int         `json:"activity_id"`
	StoreID      int64       `json:"store_id"`
	Avatar       string      `json:"avatar"`
	Nickname     string      `json:"nickname"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	TimeStamp    int64       `json:"time_stamp"`
}

type VenueNotesResponse struct {
	Notes      []VenueNoteItem `json:"notes"`
	NextCursor int64           `json:"next_cursor"`
	HasMore    bool            `json:"has_more"`
}

type SubscriptionListItem struct {
	ID           string         `json:"id"`
	Source       string         `json:"source"`
	SourceID     int64          `json:"source_id"`
	Title        string         `json:"title"`
	CoverImage   string         `json:"cover_image"`
	Description  string         `json:"description,omitempty"`
	StartTime    *time.Time     `json:"start_time,omitempty"`
	EndTime      *time.Time     `json:"end_time,omitempty"`
	Status       any            `json:"status,omitempty"`
	Address      string         `json:"address,omitempty"`
	Latitude     float64        `json:"latitude,omitempty"`
	Longitude    float64        `json:"longitude,omitempty"`
	Extra        map[string]any `json:"extra,omitempty"`
	SubscribedAt time.Time      `json:"subscribed_at"`
}

type OrganizerProfileRequest struct {
	Name          string   `json:"name"`
	Logo          string   `json:"logo"`
	CoverImage    string   `json:"cover_image"`
	Gallery       []string `json:"gallery"`
	Description   string   `json:"description"`
	BusinessHours string   `json:"business_hours"`
	ContactName   string   `json:"contact_name"`
	ServicePhone  string   `json:"service_phone"`
	Province      string   `json:"province"`
	City          string   `json:"city"`
	District      string   `json:"district"`
	Address       string   `json:"address"`
	Latitude      float64  `json:"latitude"`
	Longitude     float64  `json:"longitude"`
	AverageSpend  int64    `json:"average_spend"`
}

type OrganizerProfileResponse struct {
	ID            int64    `json:"id"`
	Name          string   `json:"name"`
	Logo          string   `json:"logo"`
	CoverImage    string   `json:"cover_image"`
	Gallery       []string `json:"gallery"`
	Description   string   `json:"description"`
	BusinessHours string   `json:"business_hours"`
	ContactName   string   `json:"contact_name"`
	ServicePhone  string   `json:"service_phone"`
	Province      string   `json:"province"`
	City          string   `json:"city"`
	District      string   `json:"district"`
	Address       string   `json:"address"`
	Latitude      float64  `json:"latitude"`
	Longitude     float64  `json:"longitude"`
	AverageSpend  int64    `json:"average_spend"`
}

// PublicOrganizerHomeResponse is the C-end public storefront for an approved
// organizer. The two content sections paginate independently for tab views.
type PublicOrganizerHomeResponse struct {
	ID               int64                          `json:"id"`
	UserID           int64                          `json:"user_id"`
	Type             string                         `json:"type"`
	Name             string                         `json:"name"`
	Logo             string                         `json:"logo"`
	OwnerNickname    string                         `json:"owner_nickname"`
	OwnerAvatar      string                         `json:"owner_avatar"`
	CoverImage       string                         `json:"cover_image"`
	Gallery          []string                       `json:"gallery"`
	Description      string                         `json:"description"`
	BusinessHours    string                         `json:"business_hours"`
	ServicePhone     string                         `json:"service_phone"`
	Province         string                         `json:"province"`
	City             string                         `json:"city"`
	District         string                         `json:"district"`
	Address          string                         `json:"address"`
	Latitude         float64                        `json:"latitude"`
	Longitude        float64                        `json:"longitude"`
	AverageSpend     int64                          `json:"average_spend"`
	FollowCount      int64                          `json:"follow_count"`
	IsFollow         bool                           `json:"is_follow"`
	FollowTargetType string                         `json:"follow_target_type"`
	FollowTargetID   int64                          `json:"follow_target_id"`
	ActivityCount    int64                          `json:"activity_count"`
	VenueCount       int64                          `json:"venue_count"`
	Activities       PageResponse[ActivityListItem] `json:"activities"`
	Venues           PageResponse[VenueListItem]    `json:"venues"`
}

type OrganizerUserLookupResponse struct {
	UserID   int64  `json:"user_id"`
	Phone    string `json:"phone"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Status   int8   `json:"status"`
}

type OrganizerPostRequest struct {
	Title       string      `json:"title" binding:"required"`
	Content     string      `json:"content"`
	Images      []string    `json:"images"`
	MediaData   []NoteMedia `json:"media_data"`
	Location    *Location   `json:"location"`
	Type        int         `json:"type"`
	Status      int         `json:"status"`
	VisibleConf int         `json:"visible_conf"`
	ActivityID  int         `json:"activity_id"`
	StoreID     int64       `json:"store_id"`
}

type OrganizerPostVisibilityRequest struct {
	Visible *bool `json:"visible"`
	Status  *int  `json:"status"`
}

type OrganizerPostItem struct {
	ID           uint64      `json:"id,string"`
	UserID       uint64      `json:"user_id"`
	Title        string      `json:"title"`
	Content      string      `json:"content"`
	MediaData    []NoteMedia `json:"media_data"`
	Location     Location    `json:"location"`
	Type         int         `json:"type"`
	Status       int         `json:"status"`
	VisibleConf  int         `json:"visible_conf"`
	ActivityID   int         `json:"activity_id"`
	ActivityName string      `json:"activity_name,omitempty"`
	StoreID      int64       `json:"store_id"`
	StoreName    string      `json:"store_name,omitempty"`
	LikeCount    int64       `json:"like_count"`
	CollCount    int64       `json:"coll_count"`
	ShareCount   int64       `json:"share_count"`
	CommentCount int64       `json:"comment_count"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

type OrganizerFinanceSummary struct {
	GrossAmount      int64 `json:"gross_amount"`
	RefundAmount     int64 `json:"refund_amount"`
	SettleAmount     int64 `json:"settle_amount"`
	WithdrawAmount   int64 `json:"withdraw_amount"`
	OrderCount       int64 `json:"order_count"`
	TodayOrderCount  int64 `json:"today_order_count"`
	TodayOrderAmount int64 `json:"today_order_amount"`
	TodayTicketCount int64 `json:"today_ticket_count"`
}

type OrganizerOrderSummary struct {
	TotalAmount        int64                        `json:"total_amount"`
	OrderCount         int64                        `json:"order_count"`
	TicketCount        int64                        `json:"ticket_count"`
	AverageOrderAmount int64                        `json:"average_order_amount"`
	ViewCount          int64                        `json:"view_count"`
	VisitorCount       int64                        `json:"visitor_count"`
	PaidOrderCount     int64                        `json:"paid_order_count"`
	ConversionRate     float64                      `json:"conversion_rate"`
	ActivityRanks      []OrganizerOrderActivityRank `json:"activity_ranks"`
}

type OrganizerOrderActivityRank struct {
	ActivityID              int64   `json:"activity_id"`
	ActivityName            string  `json:"activity_name"`
	OrderCount              int64   `json:"order_count"`
	TicketCount             int64   `json:"ticket_count"`
	TotalAmount             int64   `json:"total_amount"`
	ViewCount               int64   `json:"view_count"`
	VisitorCount            int64   `json:"visitor_count"`
	PaidOrderCount          int64   `json:"paid_order_count"`
	ConversionRate          float64 `json:"conversion_rate"`
	AvailableWithdrawAmount int64   `json:"available_withdraw_amount"`
	PendingWithdrawAmount   int64   `json:"pending_withdraw_amount"`
	WithdrawnAmount         int64   `json:"withdrawn_amount"`
}

type OrganizerFinanceFlowItem struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Amount      int64     `json:"amount"`
	OrderNo     string    `json:"order_no,omitempty"`
	RefundNo    string    `json:"refund_no,omitempty"`
	ActivityID  int64     `json:"activity_id,omitempty"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type OrganizerLevelRuleRequest struct {
	Level                 int     `json:"level" binding:"required"`
	Name                  string  `json:"name"`
	FeeRate               float64 `json:"fee_rate"`
	RequiredActivityCount int64   `json:"required_activity_count"`
	Description           string  `json:"description"`
	Benefits              string  `json:"benefits"`
	Status                int8    `json:"status"`
}

type OrganizerRoleRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	Status      int8     `json:"status"`
}

type OrganizerStaffRequest struct {
	UserID int64  `json:"user_id"`
	RoleID int64  `json:"role_id"`
	Name   string `json:"name" binding:"required"`
	Phone  string `json:"phone" binding:"required"`
	Status int8   `json:"status"`
}

type OrganizerInfoResponse struct {
	ID                     int64   `json:"id"`
	Type                   string  `json:"type"`
	Name                   string  `json:"name"`
	Logo                   string  `json:"logo"`
	Status                 int8    `json:"status"`
	RejectReason           string  `json:"reject_reason,omitempty"`
	Level                  string  `json:"level"`
	ServiceFeeRate         float64 `json:"service_fee_rate"`
	LevelValue             int     `json:"level_value"`
	FeeRate                float64 `json:"fee_rate"`
	CompletedActivityCount int64   `json:"completed_activity_count"`
	NextLevelRequiredCount int64   `json:"next_level_required_count"`
	JoinDays               int     `json:"join_days"`
	BasicInfo              struct {
		Province string `json:"province"`
		City     string `json:"city"`
		District string `json:"district"`
	} `json:"basic_info"`
	AccountInfo struct {
		BankAccountName string `json:"bank_account_name"`
		BankAccountNo   string `json:"bank_account_no"`
		BankName        string `json:"bank_name"`
	} `json:"account_info"`
}

type ActivityCreateRequest struct {
	ActivityID       int64                `json:"activity_id"`
	Step             int                  `json:"step" binding:"required,min=1,max=5"`
	Type             *string              `json:"type"`
	TagIDs           []int64              `json:"tag_ids"`
	Name             *string              `json:"name"`
	ShareTitle       *string              `json:"share_title"`
	StartTime        *string              `json:"start_time"`
	EndTime          *string              `json:"end_time"`
	BusinessHours    *string              `json:"business_hours"`
	RealNameMode     *int8                `json:"real_name_mode"`
	MinorCheck       *int8                `json:"minor_check"`
	Description      *string              `json:"description"`
	Province         *string              `json:"province"`
	City             *string              `json:"city"`
	District         *string              `json:"district"`
	Address          *string              `json:"address"`
	Latitude         *float64             `json:"latitude"`
	Longitude        *float64             `json:"longitude"`
	PosterDetail     *string              `json:"poster_detail"`
	PosterLong       *string              `json:"poster_long"`
	PosterList       *string              `json:"poster_list"`
	PosterWechat     *string              `json:"poster_wechat"`
	TicketSpecs      []TicketSpecSaveItem `json:"ticket_specs"`
	QualificationDoc *string              `json:"qualification_doc"`
}

type ActivityDetailResponse struct {
	models.Activity
	// BusinessHours belongs to the organizer profile and is returned here for
	// venue editing. It is empty for party activities.
	BusinessHours         string              `json:"business_hours"`
	UserID                int64               `json:"user_id"`
	TagIDs                []int64             `json:"tag_ids"`
	Tags                  []ContentTagItem    `json:"tags"`
	TicketSpecs           []models.TicketSpec `json:"ticket_specs"`
	Organizer             *models.Organizer   `json:"organizer,omitempty"`
	IsSubscribe           bool                `json:"is_subscribe"`
	IsFollow              bool                `json:"is_follow"`
	FollowCount           int64               `json:"follow_count"`
	FollowTargetType      string              `json:"follow_target_type"`
	FollowTargetID        int64               `json:"follow_target_id"`
	HasPendingRevision    bool                `json:"has_pending_revision"`
	PendingRevisionReason string              `json:"pending_revision_reason,omitempty"`
}

// ActivityRevisionPayload is an unpublished snapshot for an already-online
// activity. It becomes public only after an administrator approves it.
type ActivityRevisionPayload struct {
	Activity      models.Activity     `json:"activity"`
	TicketSpecs   []models.TicketSpec `json:"ticket_specs"`
	TagIDs        []int64             `json:"tag_ids"`
	BusinessHours string              `json:"business_hours"`
}

type ActivityListItem struct {
	ID                    int64            `json:"id"`
	Type                  string           `json:"type"`
	Name                  string           `json:"name"`
	PosterList            string           `json:"poster_list"`
	StartTime             time.Time        `json:"start_time"`
	EndTime               time.Time        `json:"end_time"`
	Status                int8             `json:"status"`
	AuditType             string           `json:"audit_type"`
	HasPendingRevision    bool             `json:"has_pending_revision"`
	PendingRevisionReason string           `json:"pending_revision_reason,omitempty"`
	TagIDs                []int64          `json:"tag_ids"`
	Tags                  []ContentTagItem `json:"tags"`
	IsSubscribe           bool             `json:"is_subscribe,omitempty"`
	IsFollow              bool             `json:"is_follow"`
	FollowCount           int64            `json:"follow_count"`
	FollowTargetType      string           `json:"follow_target_type"`
	FollowTargetID        int64            `json:"follow_target_id"`
}

type TicketSpecSaveItem struct {
	ID            int64  `json:"id,string"`
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description"`
	IsEnabled     int8   `json:"is_enabled"`
	SaleStart     string `json:"sale_start"`
	SaleEnd       string `json:"sale_end"`
	Price         int64  `json:"price" binding:"required"` // 分
	Stock         int    `json:"stock" binding:"min=0"`
	PurchaseLimit int    `json:"purchase_limit"`
	MaxAttendees  int    `json:"max_attendees"`
}

type SaveTicketSpecsRequest struct {
	Specs []TicketSpecSaveItem `json:"specs" binding:"required"`
}

type CreateTicketOrderRequest struct {
	ActivityID   int64              `json:"activity_id" binding:"required"`
	TicketSpecID int64              `json:"ticket_spec_id" binding:"required"`
	Quantity     int                `json:"quantity" binding:"required,min=1"`
	BuyerName    string             `json:"buyer_name"`
	BuyerIDCard  string             `json:"buyer_id_card"`
	ViewerIDs    []int64            `json:"viewer_ids"`
	Viewers      []OrderViewerInput `json:"viewers"`
	UsePoints    bool               `json:"use_points"`
	PointsAmount int64              `json:"points_amount"`
	SalesChannel string             `json:"sales_channel"`
}

type OrderViewerInput struct {
	ID       int64  `json:"id"`
	ViewerID int64  `json:"viewer_id"`
	RealName string `json:"real_name"`
	IDCard   string `json:"id_card"`
	Phone    string `json:"phone"`
	Type     int8   `json:"type"`
}

type OrderViewerItem struct {
	ViewerID     int64  `json:"viewer_id"`
	RealName     string `json:"real_name"`
	IDCard       string `json:"id_card,omitempty"`
	IDCardMasked string `json:"id_card_masked"`
	Phone        string `json:"phone,omitempty"`
	PhoneMasked  string `json:"phone_masked"`
	Type         int8   `json:"type"`
}

type CreateTicketOrderResponse struct {
	OrderNo        string            `json:"order_no"`
	Status         int8              `json:"status"`
	TotalPrice     int64             `json:"total_price"`
	PointsAmount   int64             `json:"points_amount"`
	PointsDiscount int64             `json:"points_discount"`
	ActualPrice    int64             `json:"actual_price"`
	SalesChannel   string            `json:"sales_channel"`
	QRCode         string            `json:"qr_code"`
	QRCodeURL      string            `json:"qr_code_url"`
	Viewers        []OrderViewerItem `json:"viewers,omitempty"`
}

type TicketOrderDetailResponse struct {
	OrderNo          string `json:"order_no"`
	Status           int8   `json:"status"`
	RefundStatus     string `json:"refund_status,omitempty"`
	RefundStatusText string `json:"refund_status_text,omitempty"`
	TotalPrice       int64  `json:"total_price"`
	ActualPrice      int64  `json:"actual_price"`
	PointsAmount     int64  `json:"points_amount"`
	PointsDiscount   int64  `json:"points_discount"`
	RefundNo         string `json:"refund_no,omitempty"`
	Quantity         int    `json:"quantity"`
	Activity         struct {
		ID           int64     `json:"id"`
		Name         string    `json:"name"`
		StartTime    time.Time `json:"start_time"`
		EndTime      time.Time `json:"end_time"`
		PosterList   string    `json:"poster_list"`
		IsHidden     bool      `json:"is_hidden"`
		HiddenReason string    `json:"hidden_reason,omitempty"`
	} `json:"activity"`
	TicketSpec struct {
		Name string `json:"name"`
	} `json:"ticket_spec"`
	BuyerName    string            `json:"buyer_name"`
	BuyerIDCard  string            `json:"buyer_id_card"`
	Viewers      []OrderViewerItem `json:"viewers,omitempty"`
	PayMethod    string            `json:"pay_method"`
	SalesChannel string            `json:"sales_channel"`
	PayTime      *time.Time        `json:"pay_time"`
	CreatedAt    time.Time         `json:"created_at"`
	QRCode       string            `json:"qr_code"`
	QRCodeURL    string            `json:"qr_code_url"`
	ExpireTime   time.Time         `json:"expire_time"`
	RefundInfo   *struct {
		RefundNo         string `json:"refund_no"`
		RefundAmount     int64  `json:"refund_amount"`
		Status           int8   `json:"status"`
		StatusText       string `json:"status_text"`
		ExpectArriveDate string `json:"expect_arrive_date"`
	} `json:"refund_info,omitempty"`
	Refund *struct {
		RefundNo   string `json:"refund_no"`
		Status     int8   `json:"status"`
		StatusText string `json:"status_text"`
	} `json:"refund,omitempty"`
}

type TicketOrderListItem struct {
	OrderNo          string `json:"order_no"`
	Status           int8   `json:"status"`
	RefundNo         string `json:"refund_no,omitempty"`
	RefundStatus     string `json:"refund_status,omitempty"`
	RefundStatusText string `json:"refund_status_text,omitempty"`
	TotalPrice       int64  `json:"total_price"`
	ActualPrice      int64  `json:"actual_price"`
	Quantity         int    `json:"quantity"`
	Activity         struct {
		ID           int64     `json:"id"`
		Name         string    `json:"name"`
		StartTime    time.Time `json:"start_time"`
		EndTime      time.Time `json:"end_time"`
		PosterList   string    `json:"poster_list"`
		IsHidden     bool      `json:"is_hidden"`
		HiddenReason string    `json:"hidden_reason,omitempty"`
	} `json:"activity"`
	TicketSpec struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"ticket_spec"`
	BuyerName   string            `json:"buyer_name"`
	BuyerIDCard string            `json:"buyer_id_card"`
	Viewers     []OrderViewerItem `json:"viewers,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	ExpireTime  time.Time         `json:"expire_time"`
	PayTime     *time.Time        `json:"pay_time"`
}

type CancelOrderRequest struct {
	ReasonID int64 `json:"reason_id" binding:"required"`
}

type OrganizerCancelOrderResponse struct {
	OrderNo        string    `json:"order_no"`
	Status         int8      `json:"status"`
	CancelReasonID int64     `json:"cancel_reason_id"`
	CancelledAt    time.Time `json:"cancelled_at"`
}

type ApplyRefundRequest struct {
	OrderNo  string `json:"order_no" binding:"required"`
	ReasonID int64  `json:"reason_id" binding:"required"`
}

type RejectRefundRequest struct {
	RejectReason string `json:"reject_reason" binding:"required"`
}

type OrganizerOrderListItem struct {
	OrderNo        string            `json:"order_no"`
	Status         int8              `json:"status"`
	TotalPrice     int64             `json:"total_price"`
	ActualPrice    int64             `json:"actual_price"`
	PointsAmount   int64             `json:"points_amount"`
	PointsDiscount int64             `json:"points_discount"`
	Quantity       int               `json:"quantity"`
	UserID         int64             `json:"user_id"`
	UserName       string            `json:"user_name"`
	UserMobile     string            `json:"user_mobile"`
	UserAvatar     string            `json:"user_avatar"`
	BuyerName      string            `json:"buyer_name"`
	BuyerIDCard    string            `json:"buyer_id_card"`
	Viewers        []OrderViewerItem `json:"viewers,omitempty"`
	ActivityID     int64             `json:"activity_id"`
	ActivityName   string            `json:"activity_name"`
	TicketSpecID   int64             `json:"ticket_spec_id"`
	TicketSpecName string            `json:"ticket_spec_name"`
	PayMethod      string            `json:"pay_method"`
	SalesChannel   string            `json:"sales_channel"`
	PayTime        *time.Time        `json:"pay_time"`
	CreatedAt      time.Time         `json:"created_at"`
	ExpireTime     time.Time         `json:"expire_time"`
	WithdrawStatus string            `json:"withdraw_status"`
	WithdrawAmount int64             `json:"withdraw_amount"`
}

type OrganizerOrderDetailResponse struct {
	TicketOrderDetailResponse
	UserID         int64  `json:"user_id"`
	UserName       string `json:"user_name"`
	UserMobile     string `json:"user_mobile"`
	UserAvatar     string `json:"user_avatar"`
	WithdrawStatus string `json:"withdraw_status"`
	WithdrawAmount int64  `json:"withdraw_amount"`
}

type OrganizerRefundListItem struct {
	RefundNo         string    `json:"refund_no"`
	Status           int8      `json:"status"`
	RefundAmount     int64     `json:"refund_amount"`
	DeductAmount     int64     `json:"deduct_amount"`
	Reason           string    `json:"reason"`
	RejectReason     string    `json:"reject_reason"`
	ExpectArriveDate string    `json:"expect_arrive_date"`
	WechatRefundID   string    `json:"wechat_refund_id"`
	WechatStatus     string    `json:"wechat_status"`
	OrderNo          string    `json:"order_no"`
	UserID           int64     `json:"user_id"`
	BuyerName        string    `json:"buyer_name"`
	BuyerIDCard      string    `json:"buyer_id_card"`
	ActivityID       int64     `json:"activity_id"`
	ActivityName     string    `json:"activity_name"`
	TicketSpecID     int64     `json:"ticket_spec_id"`
	TicketSpecName   string    `json:"ticket_spec_name"`
	Quantity         int       `json:"quantity"`
	CreatedAt        time.Time `json:"created_at"`
}

type OrganizerRefundDetailResponse struct {
	Refund struct {
		RefundNo         string    `json:"refund_no"`
		RefundAmount     int64     `json:"refund_amount"`
		DeductAmount     int64     `json:"deduct_amount"`
		Reason           string    `json:"reason"`
		Status           int8      `json:"status"`
		WechatRefundID   string    `json:"wechat_refund_id"`
		WechatStatus     string    `json:"wechat_status"`
		RejectReason     string    `json:"reject_reason"`
		ExpectArriveDate string    `json:"expect_arrive_date"`
		CreatedAt        time.Time `json:"created_at"`
		UpdatedAt        time.Time `json:"updated_at"`
	} `json:"refund"`
	Order struct {
		OrderNo        string     `json:"order_no"`
		Status         int8       `json:"status"`
		ActualPrice    int64      `json:"actual_price"`
		Quantity       int        `json:"quantity"`
		UserName       string     `json:"user_name"`
		UserMobile     string     `json:"user_mobile"`
		ActivityName   string     `json:"activity_name"`
		TicketSpecName string     `json:"ticket_spec_name"`
		PayMethod      string     `json:"pay_method"`
		PayTime        *time.Time `json:"pay_time"`
	} `json:"order"`
	Viewers             []OrderViewerItem                 `json:"viewers"`
	RefundLogs          []models.RefundLog                `json:"refund_logs"`
	PayRecords          []OrganizerRefundPayRecord        `json:"pay_records"`
	VerificationRecords []OrganizerRefundVerificationItem `json:"verification_records"`
}

type OrganizerRefundPayRecord struct {
	ID            uint64     `json:"id"`
	PayPlatform   int8       `json:"pay_platform"`
	PayMethod     string     `json:"pay_method"`
	TransactionID string     `json:"transaction_id"`
	AmountTotal   uint64     `json:"amount_total"`
	Currency      string     `json:"currency"`
	PayStatus     int8       `json:"pay_status"`
	TradeState    string     `json:"trade_state"`
	FinishedAt    *time.Time `json:"finished_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

type OrganizerRefundVerificationItem struct {
	ID            int64     `json:"id"`
	VerifierID    int64     `json:"verifier_id"`
	VerifierName  string    `json:"verifier_name"`
	VerifierPhone string    `json:"verifier_phone"`
	ActivityID    int64     `json:"activity_id"`
	ActivityName  string    `json:"activity_name"`
	VerifiedAt    time.Time `json:"verified_at"`
}

type UserRefundListItem struct {
	RefundNo     string    `json:"refund_no"`
	OrderNo      string    `json:"order_no"`
	Status       int8      `json:"status"`
	StatusText   string    `json:"status_text"`
	RefundAmount int64     `json:"refund_amount"`
	Reason       string    `json:"reason"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Activity     struct {
		ID         int64  `json:"id"`
		Name       string `json:"name"`
		PosterList string `json:"poster_list"`
	} `json:"activity"`
}

type RefundDetailResponse struct {
	RefundNo         string `json:"refund_no"`
	Status           int8   `json:"status"`
	RefundAmount     int64  `json:"refund_amount"`
	DeductAmount     int64  `json:"deduct_amount"`
	ExpectArriveDate string `json:"expect_arrive_date"`
	Order            struct {
		OrderNo     string `json:"order_no"`
		TotalPrice  int64  `json:"total_price"`
		ActualPrice int64  `json:"actual_price"`
		Quantity    int    `json:"quantity"`
	} `json:"order"`
	Activity struct {
		Name      string    `json:"name"`
		StartTime time.Time `json:"start_time"`
		EndTime   time.Time `json:"end_time"`
	} `json:"activity"`
	Progress []models.RefundLog `json:"progress"`
}

type VerifierRequest struct {
	Name  string `json:"name" binding:"required"`
	Phone string `json:"phone" binding:"required"`
}

type ActivateVerifierRequest struct {
	VerifierID int64  `json:"verifier_id"`
	Phone      string `json:"phone" binding:"required"`
	Channel    string `json:"channel"`
}

type ActivateVerifierResponse struct {
	Success    bool  `json:"success"`
	VerifierID int64 `json:"verifier_id"`
	UserID     int64 `json:"user_id"`
	Status     int8  `json:"status"`
}

type VerifierActivationInfoResponse struct {
	VerifierID    int64  `json:"verifier_id"`
	Name          string `json:"name"`
	Phone         string `json:"phone"`
	Status        int8   `json:"status"`
	IsBound       bool   `json:"is_bound"`
	OrganizerID   int64  `json:"organizer_id"`
	OrganizerName string `json:"organizer_name"`
}

type ScanOrderRequest struct {
	QRCode     string `json:"qr_code" binding:"required"`
	ActivityID int64  `json:"activity_id"`
}

type ScanOrderResponse struct {
	Success bool `json:"success"`
	Order   *struct {
		ActivityName      string            `json:"activity_name"`
		TicketSpecName    string            `json:"ticket_spec_name"`
		Quantity          int               `json:"quantity"`
		BuyerNameMasked   string            `json:"buyer_name_masked"`
		BuyerIDCardMasked string            `json:"buyer_id_card_masked"`
		Viewers           []OrderViewerItem `json:"viewers,omitempty"`
	} `json:"order,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
}

type ConfirmVerifyRequest struct {
	OrderNo string `json:"order_no" binding:"required"`
}

type VerifiedListItem struct {
	ActivityName      string    `json:"activity_name"`
	TicketSpecName    string    `json:"ticket_spec_name"`
	Quantity          int       `json:"quantity"`
	BuyerNameMasked   string    `json:"buyer_name_masked"`
	BuyerIDCardMasked string    `json:"buyer_id_card_masked"`
	VerifiedAt        time.Time `json:"verified_at"`
}

type ActivityStatisticsResponse struct {
	VerifyRate              float64 `json:"verify_rate"`
	TicketCount             int64   `json:"ticket_count"`
	BuyerCount              int64   `json:"buyer_count"`
	GrossAmount             int64   `json:"gross_amount"`
	RefundAmount            int64   `json:"refund_amount"`
	NetAmount               int64   `json:"net_amount"`
	AverageTicketPrice      int64   `json:"average_ticket_price"`
	VerifiedCount           int64   `json:"verified_count"`
	ViewCount               int64   `json:"view_count"`
	VisitorCount            int64   `json:"visitor_count"`
	PaidOrderCount          int64   `json:"paid_order_count"`
	ConversionRate          float64 `json:"conversion_rate"`
	AvailableWithdrawAmount int64   `json:"available_withdraw_amount"`
	PendingWithdrawAmount   int64   `json:"pending_withdraw_amount"`
	WithdrawnAmount         int64   `json:"withdrawn_amount"`
}

type ActivityDailyStatisticsItem struct {
	Date         string `json:"date"`
	Amount       int64  `json:"amount"`
	TicketCount  int64  `json:"ticket_count"`
	OrderCount   int64  `json:"order_count"`
	ViewCount    int64  `json:"view_count"`
	VisitorCount int64  `json:"visitor_count"`
}

type ViewerItem struct {
	ID        int64     `json:"id"`
	RealName  string    `json:"real_name"`
	IDCard    string    `json:"id_card"`
	Phone     string    `json:"phone"`
	Type      int8      `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type StoreRequest struct {
	Name      string  `json:"name" binding:"required"`
	Logo      string  `json:"logo"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Phone     string  `json:"phone"`
}
