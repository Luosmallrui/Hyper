package models

import "time"

const (
	OrganizerTypeVenue    = "venue"
	OrganizerTypeMerchant = "merchant"

	ActivityTypeParty = "party"
	ActivityTypeVenue = "venue"

	OrganizerStatusPending  int8 = 0
	OrganizerStatusAuditing int8 = 1
	OrganizerStatusApproved int8 = 2
	OrganizerStatusRejected int8 = 3

	ActivityStatusDraft    int8 = 0
	ActivityStatusPending  int8 = 1
	ActivityStatusAuditing int8 = 2
	ActivityStatusOnline   int8 = 3
	ActivityStatusRejected int8 = 4

	TicketOrderStatusPending       int8 = 0
	TicketOrderStatusUsable        int8 = 1
	TicketOrderStatusUsed          int8 = 2
	TicketOrderStatusCancelled     int8 = 3
	TicketOrderStatusRefunding     int8 = 4
	TicketOrderStatusRefundSuccess int8 = 5
	TicketOrderStatusRefundReject  int8 = 6

	RefundStatusAuditing  int8 = 0
	RefundStatusRunning   int8 = 1
	RefundStatusSuccess   int8 = 2
	RefundStatusRejected  int8 = 3
	RefundStatusCancelled int8 = 4

	VerifierStatusInactive int8 = 0
	VerifierStatusActive   int8 = 1

	OrganizerWithdrawAllocationStatusPending  int8 = 0
	OrganizerWithdrawAllocationStatusSettled  int8 = 1
	OrganizerWithdrawAllocationStatusReleased int8 = 2
)

type Organizer struct {
	ID               int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID           int64     `gorm:"column:user_id;not null;uniqueIndex:uk_organizer_user" json:"user_id"`
	Type             string    `gorm:"column:type;size:20;not null;default:merchant" json:"type"`
	Name             string    `gorm:"column:name;size:100;not null" json:"name"`
	Logo             string    `gorm:"column:logo;size:255" json:"logo"`
	Status           int8      `gorm:"column:status;not null;default:0" json:"status"`
	Enabled          int8      `gorm:"column:enabled;not null;default:1" json:"enabled"`
	RejectReason     string    `gorm:"column:reject_reason;size:500" json:"reject_reason"`
	Level            string    `gorm:"column:level;size:10;default:LV1" json:"level"`
	ServiceFeeRate   float64   `gorm:"column:service_fee_rate;type:decimal(5,2);default:0" json:"service_fee_rate"`
	JoinDays         int       `gorm:"-" json:"join_days"`
	Province         string    `gorm:"column:province;size:50" json:"province"`
	City             string    `gorm:"column:city;size:50" json:"city"`
	District         string    `gorm:"column:district;size:50" json:"district"`
	BankAccountName  string    `gorm:"column:bank_account_name;size:50" json:"bank_account_name"`
	BankAccountNo    string    `gorm:"column:bank_account_no;size:50" json:"bank_account_no"`
	BankName         string    `gorm:"column:bank_name;size:50" json:"bank_name"`
	BankContactName  string    `gorm:"column:bank_contact_name;size:50" json:"bank_contact_name"`
	BankContactPhone string    `gorm:"column:bank_contact_phone;size:20" json:"bank_contact_phone"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Organizer) TableName() string {
	return "organizers"
}

type Activity struct {
	ID               int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	OrganizerID      int64      `gorm:"column:organizer_id;not null;index" json:"organizer_id"`
	Type             string     `gorm:"column:type;size:20;not null;default:party;index" json:"type"`
	Name             string     `gorm:"column:name;size:80;not null" json:"name"`
	ShareTitle       string     `gorm:"column:share_title;size:20" json:"share_title"`
	StartTime        time.Time  `gorm:"column:start_time" json:"start_time"`
	EndTime          time.Time  `gorm:"column:end_time" json:"end_time"`
	RealNameMode     int8       `gorm:"column:real_name_mode;default:0" json:"real_name_mode"`
	MinorCheck       int8       `gorm:"column:minor_check;default:0" json:"minor_check"`
	Description      string     `gorm:"column:description;type:text" json:"description"`
	Province         string     `gorm:"column:province;size:50" json:"province"`
	City             string     `gorm:"column:city;size:50" json:"city"`
	District         string     `gorm:"column:district;size:50" json:"district"`
	Address          string     `gorm:"column:address;size:200" json:"address"`
	Latitude         float64    `gorm:"column:latitude;type:decimal(10,6)" json:"latitude"`
	Longitude        float64    `gorm:"column:longitude;type:decimal(10,6)" json:"longitude"`
	PosterDetail     string     `gorm:"column:poster_detail;size:255" json:"poster_detail"`
	PosterLong       string     `gorm:"column:poster_long;size:255" json:"poster_long"`
	PosterList       string     `gorm:"column:poster_list;size:255" json:"poster_list"`
	PosterWechat     string     `gorm:"column:poster_wechat;size:255" json:"poster_wechat"`
	QualificationDoc string     `gorm:"column:qualification_doc;size:255" json:"qualification_doc"`
	Status           int8       `gorm:"column:status;not null;default:0;index" json:"status"`
	RejectReason     string     `gorm:"column:reject_reason;size:500" json:"reject_reason"`
	IsHidden         int8       `gorm:"column:is_hidden;not null;default:0;index" json:"is_hidden"`
	HiddenAt         *time.Time `gorm:"column:hidden_at" json:"hidden_at,omitempty"`
	HiddenReason     string     `gorm:"column:hidden_reason;size:500" json:"hidden_reason"`
	CreatedAt        time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Activity) TableName() string {
	return "activities"
}

// ActivityDailyStat stores lightweight per-day PV/UV totals. Detailed unique
// visitor keys live in ActivityDailyVisitor so arbitrary date-range UV can be
// calculated without retaining every page-view event.
type ActivityDailyStat struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ActivityID   int64     `gorm:"column:activity_id;not null;uniqueIndex:uk_activity_daily_stat,priority:1;index" json:"activity_id"`
	StatDate     time.Time `gorm:"column:stat_date;type:date;not null;uniqueIndex:uk_activity_daily_stat,priority:2;index" json:"stat_date"`
	ViewCount    int64     `gorm:"column:view_count;not null;default:0" json:"view_count"`
	VisitorCount int64     `gorm:"column:visitor_count;not null;default:0" json:"visitor_count"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (ActivityDailyStat) TableName() string { return "activity_daily_stats" }

type ActivityDailyVisitor struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ActivityID int64     `gorm:"column:activity_id;not null;uniqueIndex:uk_activity_daily_visitor,priority:1;index" json:"activity_id"`
	StatDate   time.Time `gorm:"column:stat_date;type:date;not null;uniqueIndex:uk_activity_daily_visitor,priority:2;index" json:"stat_date"`
	VisitorKey string    `gorm:"column:visitor_key;size:80;not null;uniqueIndex:uk_activity_daily_visitor,priority:3" json:"-"`
	UserID     int64     `gorm:"column:user_id;not null;default:0;index" json:"user_id"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (ActivityDailyVisitor) TableName() string { return "activity_daily_visitors" }

type TicketSpec struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ActivityID    int64     `gorm:"column:activity_id;not null;index" json:"activity_id"`
	Name          string    `gorm:"column:name;size:30;not null" json:"name"`
	Description   string    `gorm:"column:description;size:500" json:"description"`
	IsEnabled     int8      `gorm:"column:is_enabled;not null;default:1" json:"is_enabled"`
	SaleStart     time.Time `gorm:"column:sale_start" json:"sale_start"`
	SaleEnd       time.Time `gorm:"column:sale_end" json:"sale_end"`
	Price         int64     `gorm:"column:price;not null" json:"price"` // 分
	Stock         int       `gorm:"column:stock;not null;default:0" json:"stock"`
	SoldCount     int       `gorm:"column:sold_count;not null;default:0" json:"sold_count"`
	PurchaseLimit int       `gorm:"column:purchase_limit;not null;default:1" json:"purchase_limit"`
	MaxAttendees  int       `gorm:"column:max_attendees;not null;default:1" json:"max_attendees"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (TicketSpec) TableName() string {
	return "ticket_specs"
}

type ActivitySubscription struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ActivityID int64     `gorm:"column:activity_id;not null;uniqueIndex:uk_activity_user,priority:1;index" json:"activity_id"`
	UserID     int64     `gorm:"column:user_id;not null;uniqueIndex:uk_activity_user,priority:2;index" json:"user_id"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (ActivitySubscription) TableName() string {
	return "activity_subscriptions"
}

type VenueSubscription struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	OrganizerID int64     `gorm:"column:organizer_id;not null;uniqueIndex:uk_venue_user,priority:1;index" json:"organizer_id"`
	UserID      int64     `gorm:"column:user_id;not null;uniqueIndex:uk_venue_user,priority:2;index" json:"user_id"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (VenueSubscription) TableName() string {
	return "venue_subscriptions"
}

type TicketOrder struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderNo        string     `gorm:"column:order_no;size:30;not null;uniqueIndex" json:"order_no"`
	UserID         int64      `gorm:"column:user_id;not null;index" json:"user_id"`
	ActivityID     int64      `gorm:"column:activity_id;not null;index" json:"activity_id"`
	TicketSpecID   int64      `gorm:"column:ticket_spec_id;not null;index" json:"ticket_spec_id"`
	Quantity       int        `gorm:"column:quantity;not null" json:"quantity"`
	TotalPrice     int64      `gorm:"column:total_price;not null" json:"total_price"`   // 分
	ActualPrice    int64      `gorm:"column:actual_price;not null" json:"actual_price"` // 分
	PointsAmount   int64      `gorm:"column:points_amount;not null;default:0" json:"points_amount"`
	PointsDiscount int64      `gorm:"column:points_discount;not null;default:0" json:"points_discount"` // 分
	PayMethod      string     `gorm:"column:pay_method;size:20" json:"pay_method"`
	SalesChannel   string     `gorm:"column:sales_channel;size:20;not null;default:wechat;index" json:"sales_channel"`
	PayTime        *time.Time `gorm:"column:pay_time" json:"pay_time"`
	BuyerName      string     `gorm:"column:buyer_name;size:50" json:"buyer_name"`
	BuyerIDCard    string     `gorm:"column:buyer_id_card;size:20" json:"buyer_id_card"`
	Status         int8       `gorm:"column:status;not null;default:0;index" json:"status"`
	ExpireTime     time.Time  `gorm:"column:expire_time" json:"expire_time"`
	QRCode         string     `gorm:"column:qr_code;size:255" json:"qr_code"`
	QRCodeURL      string     `gorm:"column:qr_code_url;size:255" json:"qr_code_url"`
	CancelReason   string     `gorm:"column:cancel_reason;size:100" json:"cancel_reason"`
	UserDeletedAt  *time.Time `gorm:"column:user_deleted_at" json:"user_deleted_at"`
	CreatedAt      time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (TicketOrder) TableName() string {
	return "ticket_orders"
}

// OrganizerWithdrawAllocation makes a withdrawal traceable to the exact
// order revenue it reserves. Status follows the parent withdrawal: pending,
// settled after approval, or released after rejection.
type OrganizerWithdrawAllocation struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	WithdrawID  int64     `gorm:"column:withdraw_id;not null;uniqueIndex:uk_withdraw_order,priority:1;index" json:"withdraw_id"`
	OrganizerID int64     `gorm:"column:organizer_id;not null;index" json:"organizer_id"`
	OrderID     int64     `gorm:"column:order_id;not null;uniqueIndex:uk_withdraw_order,priority:2;index" json:"order_id"`
	OrderNo     string    `gorm:"column:order_no;size:30;not null;index" json:"order_no"`
	ActivityID  int64     `gorm:"column:activity_id;not null;index" json:"activity_id"`
	Amount      int64     `gorm:"column:amount;not null" json:"amount"`
	Status      int8      `gorm:"column:status;not null;default:0;index" json:"status"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (OrganizerWithdrawAllocation) TableName() string { return "organizer_withdraw_allocations" }

type TicketOrderViewer struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID   int64     `gorm:"column:order_id;not null;index" json:"order_id"`
	OrderNo   string    `gorm:"column:order_no;size:30;not null;index" json:"order_no"`
	ViewerID  int64     `gorm:"column:viewer_id;not null;default:0;index" json:"viewer_id"`
	RealName  string    `gorm:"column:real_name;size:50;not null" json:"real_name"`
	IDCard    string    `gorm:"column:id_card;size:20;not null" json:"id_card"`
	Phone     string    `gorm:"column:phone;size:20;not null;default:''" json:"phone"`
	Type      int8      `gorm:"column:type;not null;default:1" json:"type"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (TicketOrderViewer) TableName() string {
	return "ticket_order_viewers"
}

type Refund struct {
	ID               int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID          int64     `gorm:"column:order_id;not null;index" json:"order_id"`
	RefundNo         string    `gorm:"column:refund_no;size:30;not null;uniqueIndex" json:"refund_no"`
	RefundAmount     int64     `gorm:"column:refund_amount;not null" json:"refund_amount"`
	DeductAmount     int64     `gorm:"column:deduct_amount;not null;default:0" json:"deduct_amount"`
	Reason           string    `gorm:"column:reason;size:200" json:"reason"`
	Status           int8      `gorm:"column:status;not null;default:0;index" json:"status"`
	WechatRefundID   string    `gorm:"column:wechat_refund_id;size:64" json:"wechat_refund_id"`
	WechatStatus     string    `gorm:"column:wechat_status;size:32" json:"wechat_status"`
	RejectReason     string    `gorm:"column:reject_reason;size:500" json:"reject_reason"`
	ExpectArriveDate string    `gorm:"column:expect_arrive_date;type:date" json:"expect_arrive_date"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Refund) TableName() string {
	return "refunds"
}

type RefundLog struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	RefundID    int64     `gorm:"column:refund_id;not null;index" json:"refund_id"`
	Status      string    `gorm:"column:status;size:50;not null" json:"status"`
	Description string    `gorm:"column:description;size:200;not null" json:"description"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (RefundLog) TableName() string {
	return "refund_logs"
}

type Verifier struct {
	ID              int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	OrganizerID     int64      `gorm:"column:organizer_id;not null;index" json:"organizer_id"`
	UserID          int64      `gorm:"column:user_id;not null;default:0;index" json:"user_id"`
	Name            string     `gorm:"column:name;size:50;not null" json:"name"`
	Phone           string     `gorm:"column:phone;size:20;not null;index" json:"phone"`
	Status          int8       `gorm:"column:status;not null;default:0" json:"status"`
	PermissionScope string     `gorm:"column:permission_scope;size:20;default:活动" json:"permission_scope"`
	Channel         string     `gorm:"column:channel;size:20" json:"channel"`
	BoundAt         *time.Time `gorm:"column:bound_at" json:"bound_at"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Verifier) TableName() string {
	return "verifiers"
}

type VerificationRecord struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID    int64     `gorm:"column:order_id;not null;index" json:"order_id"`
	VerifierID int64     `gorm:"column:verifier_id;not null;index" json:"verifier_id"`
	ActivityID int64     `gorm:"column:activity_id;not null;index" json:"activity_id"`
	VerifiedAt time.Time `gorm:"column:verified_at;not null" json:"verified_at"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (VerificationRecord) TableName() string {
	return "verification_records"
}

type CancelReason struct {
	ID     int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Reason string `gorm:"column:reason;size:100;not null" json:"reason"`
	Sort   int    `gorm:"column:sort;not null;default:0" json:"sort"`
}

func (CancelReason) TableName() string {
	return "cancel_reasons"
}

type RefundReason struct {
	ID     int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Reason string `gorm:"column:reason;size:100;not null" json:"reason"`
	Sort   int    `gorm:"column:sort;not null;default:0" json:"sort"`
}

func (RefundReason) TableName() string {
	return "refund_reasons"
}

type OrganizerStore struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	OrganizerID int64     `gorm:"column:organizer_id;not null;index" json:"organizer_id"`
	Name        string    `gorm:"column:name;size:100;not null" json:"name"`
	Logo        string    `gorm:"column:logo;size:255" json:"logo"`
	Address     string    `gorm:"column:address;size:200" json:"address"`
	Latitude    float64   `gorm:"column:latitude;type:decimal(10,6)" json:"latitude"`
	Longitude   float64   `gorm:"column:longitude;type:decimal(10,6)" json:"longitude"`
	Phone       string    `gorm:"column:phone;size:20" json:"phone"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (OrganizerStore) TableName() string {
	return "organizer_stores"
}

type PlatformBanner struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Title     string    `gorm:"column:title;size:100" json:"title"`
	Image     string    `gorm:"column:image;size:255;not null" json:"image"`
	Link      string    `gorm:"column:link;size:255" json:"link"`
	Position  string    `gorm:"column:position;size:50;not null;default:home" json:"position"`
	Sort      int       `gorm:"column:sort;not null;default:0" json:"sort"`
	Status    int8      `gorm:"column:status;not null;default:1" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (PlatformBanner) TableName() string {
	return "platform_banners"
}

type PlatformSetting struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Key       string    `gorm:"column:setting_key;size:100;not null;uniqueIndex" json:"key"`
	Value     string    `gorm:"column:setting_value;type:text" json:"value"`
	Remark    string    `gorm:"column:remark;size:255" json:"remark"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (PlatformSetting) TableName() string {
	return "platform_settings"
}
