package models

import "time"

type AdminRole struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"column:name;size:50;not null" json:"name"`
	Description string    `gorm:"column:description;size:255" json:"description"`
	Permissions string    `gorm:"column:permissions;type:text" json:"permissions"`
	Status      int8      `gorm:"column:status;not null;default:1" json:"status"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (AdminRole) TableName() string { return "admin_roles" }

type AdminOperationLog struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	AdminID      int64     `gorm:"column:admin_id;not null;index" json:"admin_id"`
	Action       string    `gorm:"column:action;size:100;not null;index" json:"action"`
	Resource     string    `gorm:"column:resource;size:100" json:"resource"`
	ResourceType string    `gorm:"column:resource_type;size:50;index" json:"resource_type"`
	ResourceID   string    `gorm:"column:resource_id;size:100;index" json:"resource_id"`
	ResourceName string    `gorm:"column:resource_name;size:255" json:"resource_name"`
	Method       string    `gorm:"column:method;size:20" json:"method"`
	Path         string    `gorm:"column:path;size:255" json:"path"`
	Result       string    `gorm:"column:result;size:20;not null;default:success;index" json:"result"`
	ErrorCode    string    `gorm:"column:error_code;size:100" json:"error_code"`
	ErrorMessage string    `gorm:"column:error_message;size:255" json:"error_message"`
	IP           string    `gorm:"column:ip;size:50" json:"ip"`
	Remark       string    `gorm:"column:remark;size:255" json:"remark"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime;index" json:"created_at"`
}

func (AdminOperationLog) TableName() string { return "admin_operation_logs" }

type AdminCategory struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Type      string    `gorm:"column:type;size:50;not null;index" json:"type"`
	Name      string    `gorm:"column:name;size:100;not null" json:"name"`
	Image     string    `gorm:"column:image;size:255" json:"image"`
	Value     string    `gorm:"column:value;size:100" json:"value"`
	Sort      int       `gorm:"column:sort;not null;default:0" json:"sort"`
	Status    int8      `gorm:"column:status;not null;default:1" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (AdminCategory) TableName() string { return "admin_categories" }

type ActivityCollection struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	OrganizerID int64     `gorm:"column:organizer_id;not null;index" json:"organizer_id"`
	Title       string    `gorm:"column:title;size:100;not null" json:"title"`
	ShareTitle  string    `gorm:"column:share_title;size:100" json:"share_title"`
	Description string    `gorm:"column:description;type:text" json:"description"`
	ShareImage  string    `gorm:"column:share_image;size:255" json:"share_image"`
	Status      int8      `gorm:"column:status;not null;default:1" json:"status"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (ActivityCollection) TableName() string { return "activity_collections" }

type ActivityCollectionItem struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	CollectionID int64     `gorm:"column:collection_id;not null;index" json:"collection_id"`
	ActivityID   int64     `gorm:"column:activity_id;not null;index" json:"activity_id"`
	Sort         int       `gorm:"column:sort;not null;default:0" json:"sort"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (ActivityCollectionItem) TableName() string { return "activity_collection_items" }

type PlatformMessage struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Title       string    `gorm:"column:title;size:100;not null" json:"title"`
	Content     string    `gorm:"column:content;type:text" json:"content"`
	ContentType string    `gorm:"column:content_type;size:20;not null;default:text" json:"content_type"`
	CoverImage  string    `gorm:"column:cover_image;size:255" json:"cover_image"`
	MediaData   string    `gorm:"column:media_data;type:text" json:"media_data"`
	Type        string    `gorm:"column:type;size:50" json:"type"`
	Target      string    `gorm:"column:target;size:50" json:"target"`
	Channel     string    `gorm:"column:channel;size:30;not null;default:in_app" json:"channel"`
	CreatorID   int64     `gorm:"column:creator_id;not null;default:0;index" json:"creator_id"`
	Status      int8      `gorm:"column:status;not null;default:1" json:"status"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (PlatformMessage) TableName() string { return "platform_messages" }

type PlatformMessageDelivery struct {
	ID        int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	MessageID int64      `gorm:"column:message_id;not null;uniqueIndex:uk_message_user,priority:1;index" json:"message_id"`
	UserID    int64      `gorm:"column:user_id;not null;uniqueIndex:uk_message_user,priority:2;index" json:"user_id"`
	Status    int8       `gorm:"column:status;not null;default:0;index" json:"status"`
	SentAt    *time.Time `gorm:"column:sent_at" json:"sent_at"`
	ReadAt    *time.Time `gorm:"column:read_at" json:"read_at"`
	Error     string     `gorm:"column:error;size:255" json:"error"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (PlatformMessageDelivery) TableName() string { return "platform_message_deliveries" }

type OrganizerMessageRead struct {
	ID          int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	OrganizerID int64      `gorm:"column:organizer_id;not null;uniqueIndex:uk_org_msg,priority:1;index" json:"organizer_id"`
	MessageID   int64      `gorm:"column:message_id;not null;uniqueIndex:uk_org_msg,priority:2;index" json:"message_id"`
	IsRead      int8       `gorm:"column:is_read;not null;default:0" json:"is_read"`
	ReadAt      *time.Time `gorm:"column:read_at" json:"read_at"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (OrganizerMessageRead) TableName() string { return "organizer_message_reads" }

type OrganizerProfile struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	OrganizerID   int64     `gorm:"column:organizer_id;not null;uniqueIndex" json:"organizer_id"`
	CoverImage    string    `gorm:"column:cover_image;size:255" json:"cover_image"`
	Gallery       string    `gorm:"column:gallery;type:text" json:"gallery"`
	Description   string    `gorm:"column:description;type:text" json:"description"`
	BusinessHours string    `gorm:"column:business_hours;size:100" json:"business_hours"`
	ContactName   string    `gorm:"column:contact_name;size:50" json:"contact_name"`
	ServicePhone  string    `gorm:"column:service_phone;size:20" json:"service_phone"`
	Address       string    `gorm:"column:address;size:255" json:"address"`
	Latitude      float64   `gorm:"column:latitude;type:decimal(10,6)" json:"latitude"`
	Longitude     float64   `gorm:"column:longitude;type:decimal(10,6)" json:"longitude"`
	AverageSpend  int64     `gorm:"column:average_spend;not null;default:0" json:"average_spend"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (OrganizerProfile) TableName() string { return "organizer_profiles" }

type OrganizerWithdraw struct {
	ID              int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	OrganizerID     int64     `gorm:"column:organizer_id;not null;index" json:"organizer_id"`
	Amount          int64     `gorm:"column:amount;not null" json:"amount"`
	BankAccountName string    `gorm:"column:bank_account_name;size:50;not null;default:''" json:"bank_account_name"`
	BankAccountNo   string    `gorm:"column:bank_account_no;size:50;not null;default:''" json:"bank_account_no"`
	BankName        string    `gorm:"column:bank_name;size:50;not null;default:''" json:"bank_name"`
	Status          int8      `gorm:"column:status;not null;default:0" json:"status"`
	Remark          string    `gorm:"column:remark;size:255" json:"remark"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (OrganizerWithdraw) TableName() string { return "organizer_withdraws" }

// PlatformFinanceFlow is the immutable platform-side accounting ledger.
// Rows are only inserted; correcting a business event creates a reversal row.
type PlatformFinanceFlow struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	FlowNo        string    `gorm:"column:flow_no;size:40;not null;uniqueIndex" json:"flow_no"`
	BusinessKey   string    `gorm:"column:business_key;size:100;not null;uniqueIndex" json:"-"`
	Type          string    `gorm:"column:type;size:32;not null;index" json:"type"`
	Direction     string    `gorm:"column:direction;size:16;not null;index" json:"direction"`
	Amount        int64     `gorm:"column:amount;not null" json:"amount"`
	OrderNo       string    `gorm:"column:order_no;size:30;not null;default:'';index" json:"order_no"`
	RefundNo      string    `gorm:"column:refund_no;size:30;not null;default:'';index" json:"refund_no"`
	WithdrawID    int64     `gorm:"column:withdraw_id;not null;default:0;index" json:"withdraw_id"`
	OrganizerID   int64     `gorm:"column:organizer_id;not null;default:0;index" json:"organizer_id"`
	OrganizerName string    `gorm:"column:organizer_name;size:100;not null;default:''" json:"organizer_name"`
	PayMethod     string    `gorm:"column:pay_method;size:20;not null;default:''" json:"pay_method"`
	Remark        string    `gorm:"column:remark;size:255;not null;default:''" json:"remark"`
	OccurredAt    time.Time `gorm:"column:occurred_at;not null;index" json:"occurred_at"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (PlatformFinanceFlow) TableName() string { return "platform_finance_flows" }

const (
	OrganizerBankAuditStatusPending  int8 = 0
	OrganizerBankAuditStatusApproved int8 = 1
	OrganizerBankAuditStatusRejected int8 = 2
)

type OrganizerBankAccountAudit struct {
	ID              int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	OrganizerID     int64      `gorm:"column:organizer_id;not null;index" json:"organizer_id"`
	UserID          int64      `gorm:"column:user_id;not null;index" json:"user_id"`
	BankAccountName string     `gorm:"column:bank_account_name;size:50;not null" json:"bank_account_name"`
	BankAccountNo   string     `gorm:"column:bank_account_no;size:50;not null" json:"bank_account_no"`
	BankName        string     `gorm:"column:bank_name;size:50;not null" json:"bank_name"`
	Status          int8       `gorm:"column:status;not null;default:0;index" json:"status"`
	RejectReason    string     `gorm:"column:reject_reason;size:255" json:"reject_reason"`
	ReviewedAt      *time.Time `gorm:"column:reviewed_at" json:"reviewed_at"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (OrganizerBankAccountAudit) TableName() string { return "organizer_bank_account_audits" }

type OrganizerLevelRule struct {
	ID                    int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Level                 int       `gorm:"column:level;not null;uniqueIndex" json:"level"`
	Name                  string    `gorm:"column:name;size:50;not null" json:"name"`
	FeeRate               float64   `gorm:"column:fee_rate;type:decimal(5,4);not null;default:0" json:"fee_rate"`
	RequiredActivityCount int64     `gorm:"column:required_activity_count;not null;default:0" json:"required_activity_count"`
	Description           string    `gorm:"column:description;type:text" json:"description"`
	Benefits              string    `gorm:"column:benefits;type:text" json:"benefits"`
	Status                int8      `gorm:"column:status;not null;default:1" json:"status"`
	CreatedAt             time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt             time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (OrganizerLevelRule) TableName() string { return "organizer_level_rules" }

type OrganizerRole struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	OrganizerID int64     `gorm:"column:organizer_id;not null;index" json:"organizer_id"`
	Name        string    `gorm:"column:name;size:50;not null" json:"name"`
	Description string    `gorm:"column:description;size:255" json:"description"`
	Permissions string    `gorm:"column:permissions;type:text" json:"permissions"`
	Status      int8      `gorm:"column:status;not null;default:1" json:"status"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (OrganizerRole) TableName() string { return "organizer_roles" }

type OrganizerStaff struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	OrganizerID int64     `gorm:"column:organizer_id;not null;index" json:"organizer_id"`
	UserID      int64     `gorm:"column:user_id;not null;default:0;index" json:"user_id"`
	RoleID      int64     `gorm:"column:role_id;not null;default:0;index" json:"role_id"`
	Name        string    `gorm:"column:name;size:50;not null" json:"name"`
	Phone       string    `gorm:"column:phone;size:20;not null;index" json:"phone"`
	Status      int8      `gorm:"column:status;not null;default:1" json:"status"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (OrganizerStaff) TableName() string { return "organizer_staff" }

type OrganizerOperationLog struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	OrganizerID int64     `gorm:"column:organizer_id;not null;index" json:"organizer_id"`
	OperatorID  int64     `gorm:"column:operator_id;not null;default:0;index" json:"operator_id"`
	Operator    string    `gorm:"column:operator;size:50" json:"operator"`
	Action      string    `gorm:"column:action;size:100;not null" json:"action"`
	Resource    string    `gorm:"column:resource;size:100" json:"resource"`
	Method      string    `gorm:"column:method;size:20" json:"method"`
	Path        string    `gorm:"column:path;size:255" json:"path"`
	IP          string    `gorm:"column:ip;size:50" json:"ip"`
	Remark      string    `gorm:"column:remark;size:255" json:"remark"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (OrganizerOperationLog) TableName() string { return "organizer_operation_logs" }
