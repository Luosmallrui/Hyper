package types

import "time"

type AdminPageResponse[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

type AdminProfileResponse struct {
	ID          int       `json:"id"`
	Username    string    `json:"username"`
	Nickname    string    `json:"nickname"`
	Avatar      string    `json:"avatar"`
	Mobile      string    `json:"mobile"`
	Email       string    `json:"email"`
	Motto       string    `json:"motto"`
	RoleID      int64     `json:"role_id"`
	RoleName    string    `json:"role_name"`
	Permissions []string  `json:"permissions"`
	Status      int8      `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AdminRoleSummary struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

type AdminAccountItem struct {
	ID        int              `json:"id"`
	Username  string           `json:"username"`
	Nickname  string           `json:"nickname"`
	Avatar    string           `json:"avatar"`
	Mobile    string           `json:"mobile"`
	Email     string           `json:"email"`
	RoleID    int64            `json:"role_id"`
	Role      AdminRoleSummary `json:"role"`
	Status    int8             `json:"status"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

type AdminRoleItem struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Permissions []string  `json:"permissions"`
	Status      int8      `json:"status"`
	MemberCount int64     `json:"member_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AdminOperationLogFilter struct {
	AdminID      int64
	Action       string
	ResourceType string
	Result       string
	Keyword      string
	StartDate    *time.Time
	EndDate      *time.Time
}

type AdminOperationLogItem struct {
	ID            int64     `json:"id"`
	AdminID       int64     `json:"admin_id"`
	AdminName     string    `json:"admin_name"`
	AdminUsername string    `json:"admin_username"`
	Action        string    `json:"action"`
	ActionName    string    `json:"action_name"`
	ResourceType  string    `json:"resource_type"`
	ResourceID    string    `json:"resource_id"`
	ResourceName  string    `json:"resource_name"`
	Method        string    `json:"method"`
	Path          string    `json:"path"`
	Remark        string    `json:"remark"`
	Result        string    `json:"result"`
	ErrorCode     string    `json:"error_code"`
	ErrorMessage  string    `json:"error_message"`
	IP            string    `json:"ip"`
	CreatedAt     time.Time `json:"created_at"`
}

type AdminProfileRequest struct {
	Avatar   string `json:"avatar"`
	Nickname string `json:"nickname"`
	Mobile   string `json:"mobile"`
	Email    string `json:"email"`
	Motto    string `json:"motto"`
}

type AdminVerifierRequest struct {
	OrganizerID     int64  `json:"organizer_id" binding:"required"`
	UserID          int64  `json:"user_id"`
	Name            string `json:"name" binding:"required"`
	Phone           string `json:"phone" binding:"required"`
	Status          int8   `json:"status"`
	PermissionScope string `json:"permission_scope"`
	Channel         string `json:"channel"`
}

type AdminPasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type AdminAccountRequest struct {
	Username string `json:"username" binding:"required"`
	Nickname string `json:"nickname"`
	Password string `json:"password"`
	Avatar   string `json:"avatar"`
	Mobile   string `json:"mobile"`
	Email    string `json:"email"`
	RoleID   int64  `json:"role_id"`
	Status   int8   `json:"status"`
}

type AdminRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Permissions string `json:"permissions"`
	Status      int8   `json:"status"`
}

type AdminCategoryRequest struct {
	Type   string `json:"type" binding:"required"`
	Name   string `json:"name" binding:"required"`
	Image  string `json:"image"`
	Value  string `json:"value"`
	Sort   int    `json:"sort"`
	Status *int8  `json:"status"`
}

type AdminSimpleStatusRequest struct {
	Status int8 `json:"status" binding:"required"`
}

type AdminNoteStatusRequest struct {
	Status int8 `json:"status" binding:"oneof=-1 0 1"`
}

type ActivityCollectionRequest struct {
	OrganizerID int64   `json:"organizer_id" binding:"required"`
	Title       string  `json:"title" binding:"required"`
	ShareTitle  string  `json:"share_title"`
	Description string  `json:"description"`
	ShareImage  string  `json:"share_image"`
	Status      int8    `json:"status"`
	ActivityIDs []int64 `json:"activity_ids"`
}

type ActivityCollectionItem struct {
	ID            int64     `json:"id"`
	OrganizerID   int64     `json:"organizer_id"`
	OrganizerName string    `json:"organizer_name"`
	Title         string    `json:"title"`
	ShareTitle    string    `json:"share_title"`
	Description   string    `json:"description"`
	ShareImage    string    `json:"share_image"`
	Status        int8      `json:"status"`
	ActivityCount int       `json:"activity_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type PlatformMessageRequest struct {
	Title         string   `json:"title" binding:"required"`
	Content       string   `json:"content" binding:"required"`
	ContentType   string   `json:"content_type"`
	CoverImage    string   `json:"cover_image"`
	MediaData     []string `json:"media_data"`
	Type          string   `json:"type"`
	Target        string   `json:"target"`
	Channel       string   `json:"channel"`
	TargetUserIDs []int64  `json:"target_user_ids"`
	OrganizerIDs  []int64  `json:"organizer_ids"`
	Status        int8     `json:"status"`
}

type WithdrawAuditRequest struct {
	Status int8   `json:"status" binding:"required"`
	Remark string `json:"remark"`
}

type BankAccountAuditRequest struct {
	Status       int8   `json:"status" binding:"required"`
	RejectReason string `json:"reject_reason"`
}

type AdminCommentStatusRequest struct {
	Status int8   `json:"status" binding:"oneof=-1 0 1"`
	Reason string `json:"reason" binding:"max=255"`
}

type AdminNoteInteractionFilter struct {
	Type      string
	NoteID    int64
	UserID    int64
	Keyword   string
	Channel   string
	StartDate *time.Time
	EndDate   *time.Time
}

type AdminNoteCommentFilter struct {
	Status    *int8
	NoteID    int64
	UserID    int64
	Keyword   string
	StartDate *time.Time
	EndDate   *time.Time
}

type AdminMessageDeliveryFilter struct {
	DeliveryStatus string
	ReadStatus     string
}

type AdminPointsAdjustRequest struct {
	UserID    int64  `json:"user_id" binding:"required"`
	Points    int64  `json:"points" binding:"required"`
	Reason    string `json:"reason" binding:"required,max=255"`
	RequestNo string `json:"request_no" binding:"required,max=64"`
}

type AdminPointsAdjustResponse struct {
	UserID       int64  `json:"user_id"`
	ChangePoints int64  `json:"change_points"`
	Balance      int64  `json:"balance"`
	RequestNo    string `json:"request_no"`
}
