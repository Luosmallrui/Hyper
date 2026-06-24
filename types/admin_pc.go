package types

import "time"

type AdminPageResponse[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

type AdminProfileResponse struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Avatar    string    `json:"avatar"`
	Mobile    string    `json:"mobile"`
	Email     string    `json:"email"`
	Motto     string    `json:"motto"`
	RoleID    int64     `json:"role_id"`
	Status    int8      `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AdminProfileRequest struct {
	Avatar string `json:"avatar"`
	Mobile string `json:"mobile"`
	Email  string `json:"email"`
	Motto  string `json:"motto"`
}

type AdminPasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type AdminAccountRequest struct {
	Username string `json:"username" binding:"required"`
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
	Status int8   `json:"status"`
}

type AdminSimpleStatusRequest struct {
	Status int8 `json:"status" binding:"required"`
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
	Title         string  `json:"title" binding:"required"`
	Content       string  `json:"content"`
	Type          string  `json:"type"`
	Target        string  `json:"target"`
	TargetUserIDs []int64 `json:"target_user_ids"`
	OrganizerIDs  []int64 `json:"organizer_ids"`
	Status        int8    `json:"status"`
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
	Status int8 `json:"status" binding:"oneof=-1 0 1"`
}
