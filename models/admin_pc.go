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
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	AdminID   int64     `gorm:"column:admin_id;not null;index" json:"admin_id"`
	Action    string    `gorm:"column:action;size:100;not null" json:"action"`
	Resource  string    `gorm:"column:resource;size:100" json:"resource"`
	Method    string    `gorm:"column:method;size:20" json:"method"`
	Path      string    `gorm:"column:path;size:255" json:"path"`
	IP        string    `gorm:"column:ip;size:50" json:"ip"`
	Remark    string    `gorm:"column:remark;size:255" json:"remark"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
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
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Title     string    `gorm:"column:title;size:100;not null" json:"title"`
	Content   string    `gorm:"column:content;type:text" json:"content"`
	Type      string    `gorm:"column:type;size:50" json:"type"`
	Target    string    `gorm:"column:target;size:50" json:"target"`
	Status    int8      `gorm:"column:status;not null;default:1" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (PlatformMessage) TableName() string { return "platform_messages" }

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
