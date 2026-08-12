package types

// GroupMemberPermissions describes which management actions the current
// member can perform in a group. It is deliberately action-based so clients
// do not need to reimplement role comparison rules.
type GroupMemberPermissions struct {
	CanInvite          bool `json:"can_invite"`
	CanManageMembers   bool `json:"can_manage_members"`
	CanMuteMembers     bool `json:"can_mute_members"`
	CanMuteAll         bool `json:"can_mute_all"`
	CanSetAdmin        bool `json:"can_set_admin"`
	CanTransferOwner   bool `json:"can_transfer_owner"`
	CanUpdateGroupInfo bool `json:"can_update_group_info"`
	CanDismissGroup    bool `json:"can_dismiss_group"`
	CanQuit            bool `json:"can_quit"`
}

// GroupMemberGroupInfo contains the public group header information needed by
// a WeChat-style group detail page. MemberCount is calculated from active
// member records rather than trusting a potentially stale denormalized count.
type GroupMemberGroupInfo struct {
	ID               int      `json:"id"`
	Name             string   `json:"name"`
	Avatar           string   `json:"avatar"`
	Description      string   `json:"description"`
	OwnerID          int      `json:"owner_id"`
	MemberCount      int      `json:"member_count"`
	MaxMembers       int      `json:"max_members"`
	IsMuteAll        bool     `json:"is_mute_all"`
	CreatedAt        string   `json:"created_at"`
	MemberAvatarList []string `json:"member_avatar_list"`
}

// GroupCurrentMemberInfo is the current caller's relationship to the group.
type GroupCurrentMemberInfo struct {
	UserID         int                    `json:"user_id"`
	Role           int                    `json:"role"`
	RoleName       string                 `json:"role_name"`
	UserCard       string                 `json:"user_card"`
	DisplayName    string                 `json:"display_name"`
	IsOwner        bool                   `json:"is_owner"`
	IsAdmin        bool                   `json:"is_admin"`
	IsMuted        bool                   `json:"is_muted"`
	CanSendMessage bool                   `json:"can_send_message"`
	Permissions    GroupMemberPermissions `json:"permissions"`
}

// 邀请成员请求
// InviteMemberRequest 邀请成员请求体
type InviteMemberRequest struct {
	GroupId        int   `json:"group_id" binding:"required"`
	InvitedUserIds []int `json:"invited_user_ids" binding:"required,min=1"`
}

// InviteMemberResponse 邀请成员响应体
type InviteMemberResponse struct {
	SuccessCount  int   `json:"success_count"`   // 成功入群的人数
	FailedCount   int   `json:"failed_count"`    // 处理失败的人数
	FailedUserIds []int `json:"failed_user_ids"` // 失败的用户ID列表（可选，方便前端展示）
}

// GroupInviteCandidate is the minimum profile needed to confirm an invitation
// after a group manager searches an exact mobile number.
type GroupInviteCandidate struct {
	UserID               int    `json:"user_id"`
	MobileMasked         string `json:"mobile_masked"`
	Nickname             string `json:"nickname"`
	Avatar               string `json:"avatar"`
	Motto                string `json:"motto"`
	MembershipStatus     string `json:"membership_status"`
	CanInvite            bool   `json:"can_invite"`
	InviteDisabledReason string `json:"invite_disabled_reason"`
}

type FindGroupInviteCandidateRequest struct {
	GroupID int    `json:"group_id" binding:"required"`
	Mobile  string `json:"mobile" binding:"required"`
}

// GroupInviteCandidateResponse distinguishes a normal no-match result from a
// failed permission or service request.
type GroupInviteCandidateResponse struct {
	Found     bool                  `json:"found"`
	Candidate *GroupInviteCandidate `json:"candidate"`
}

// 踢出成员请求
type KickMemberRequest struct {
	GroupId      int `json:"group_id" binding:"required"`
	KickedUserId int `json:"kicked_user_id" binding:"required"`
}
type KickmemberResponse struct {
	Success bool `json:"success"`
}

// 群成员列表元素（DTO）
type GroupMemberItemDTO struct {
	UserId         int    `json:"user_id"`
	Avatar         string `json:"avatar"`
	Nickname       string `json:"nickname"`
	DisplayName    string `json:"display_name"`
	Gender         int    `json:"gender"`
	Motto          string `json:"motto"`
	Role           int    `json:"role"`
	RoleName       string `json:"role_name"`
	IsMute         int    `json:"is_mute"`
	IsMuted        bool   `json:"is_muted"`
	CanSendMessage bool   `json:"can_send_message"`
	UserCard       string `json:"user_card"`
	JoinTime       string `json:"join_time"`
	IsCurrentUser  bool   `json:"is_current_user"`
	CanKick        bool   `json:"can_kick"`
	CanMute        bool   `json:"can_mute"`
	CanSetAdmin    bool   `json:"can_set_admin"`
	CanTransfer    bool   `json:"can_transfer_owner"`
}

// 群成员列表响应（DTO）
type GroupMemberListResponse struct {
	Group       GroupMemberGroupInfo   `json:"group"`
	CurrentUser GroupCurrentMemberInfo `json:"current_user"`
	Members     []GroupMemberItemDTO   `json:"members"`
}

// 退群/解散群 请求
type QuitGroupRequest struct {
	GroupId int `json:"group_id" binding:"required"`
}

// 退群/解散群 响应
type QuitGroupResponse struct {
	Disbanded bool `json:"disbanded"` // true=群主触发解散；false=普通退群
}

// 个人禁言/解除
type MuteMemberRequest struct {
	GroupId      int   `json:"group_id" binding:"required"`
	TargetUserId int   `json:"target_user_id" binding:"required"`
	Mute         *bool `json:"mute" binding:"required"` // true=禁言 false=解除
}

// 群全员禁言开关
type MuteAllRequest struct {
	GroupId int   `json:"group_id" binding:"required"`
	Mute    *bool `json:"mute" binding:"required"` // true=开启 false=关闭
}

type MuteAllResponse struct {
	IsMuteAll bool `json:"is_mute_all"`
}

// 设置/撤销管理员 请求
type SetAdminRequest struct {
	GroupId      int   `json:"group_id" binding:"required"`
	TargetUserId int   `json:"target_user_id" binding:"required"`
	Admin        *bool `json:"admin" binding:"required"` // true=设为管理员 false=撤销
}

// 转让群主 请求
type TransferOwnerRequest struct {
	GroupId    int `json:"group_id" binding:"required"`
	NewOwnerId int `json:"new_owner_id" binding:"required"`
}

// 转让群主 响应
type TransferOwnerResponse struct {
	Success bool `json:"success"`
}
