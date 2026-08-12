package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type GroupMemberHandler struct {
	Config             *config.Config
	GroupMemberService service.IGroupMemberService
}

func (h *GroupMemberHandler) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(h.Config.Jwt.Secret))
	group := r.Group("/v1/groupmember")
	group.POST("/invite-candidate", authorize, context.Wrap(h.FindInviteCandidate))
	group.POST("/invite", authorize, context.Wrap(h.InviteMember)) //邀请成员
	group.POST("/kick", authorize, context.Wrap(h.KickMember))
	group.GET("/list", authorize, context.Wrap(h.ListMembers))
	group.POST("/quit", authorize, context.Wrap(h.QuitGroup))
	group.POST("/mute", authorize, context.Wrap(h.MuteMember))
	group.POST("/mute-all", authorize, context.Wrap(h.MuteAll))
	group.POST("/admin", authorize, context.Wrap(h.SetAdmin))
	group.POST("/transfer-owner", authorize, context.Wrap(h.TransferOwner))
}

func (h *GroupMemberHandler) FindInviteCandidate(c *gin.Context) error {
	var req types.FindGroupInviteCandidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}
	mobile := strings.TrimSpace(req.Mobile)
	if !isValidGroupInviteMobile(mobile) {
		return response.NewError(http.StatusBadRequest, "请输入正确的 11 位手机号")
	}
	userID, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}
	result, err := h.GroupMemberService.FindInviteCandidate(c.Request.Context(), req.GroupID, int(userID), mobile)
	if err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}
	response.Success(c, result)
	return nil
}

func (h *GroupMemberHandler) InviteMember(c *gin.Context) error {
	var req types.InviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}
	uid64, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}
	resp, err := h.GroupMemberService.InviteMembers(c, req.GroupId, req.InvitedUserIds, int(uid64))
	if err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}

	response.Success(c, resp)
	return nil
}

func (h *GroupMemberHandler) KickMember(c *gin.Context) error {
	var req types.KickMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}

	uid64, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}
	userId := int(uid64)

	if userId == req.KickedUserId {
		return response.NewError(http.StatusBadRequest, "不能踢出自己")
	}

	err = h.GroupMemberService.KickMember(c, req.GroupId, req.KickedUserId, userId)
	if err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}

	response.Success(c, gin.H{"success": true})
	return nil
}

func (h *GroupMemberHandler) ListMembers(c *gin.Context) error {
	gid, err := groupMemberIDFromQuery(c)
	if err != nil {
		return err
	}

	// 2) 获取当前登录用户
	uid64, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}
	uid := int(uid64)

	// 3) 调 service
	result, err := h.GroupMemberService.ListMembers(c, gid, uid)
	if err != nil {
		// 保留 service 返回的 403/404 业务状态，前端可据此提示已退群或群已解散。
		return err
	}

	// 4) 返回
	response.Success(c, result)
	return nil
}

func groupMemberIDFromQuery(c *gin.Context) (int, error) {
	// group_id 是正式契约；兼容历史页面使用的 groupId，避免参数命名差异导致查询失败。
	gidStr := strings.TrimSpace(c.Query("group_id"))
	if gidStr == "" {
		gidStr = strings.TrimSpace(c.Query("groupId"))
	}
	if gidStr == "" {
		return 0, response.NewError(http.StatusBadRequest, "group_id 不能为空")
	}
	gid, err := strconv.Atoi(gidStr)
	if err != nil || gid <= 0 {
		return 0, response.NewError(http.StatusBadRequest, "group_id 参数错误")
	}
	return gid, nil
}

var groupInviteMobilePattern = regexp.MustCompile(`^1[3-9]\d{9}$`)

func isValidGroupInviteMobile(mobile string) bool {
	return groupInviteMobilePattern.MatchString(mobile)
}
func (h *GroupMemberHandler) QuitGroup(c *gin.Context) error {
	var req types.QuitGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "请求参数错误")
	}

	uid64, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}

	resp, err := h.GroupMemberService.QuitGroup(c.Request.Context(), req.GroupId, int(uid64))
	if err != nil {
		if err.Error() == "无权限操作" {
			return response.NewError(http.StatusForbidden, err.Error())
		}
		return response.NewError(http.StatusBadRequest, err.Error())
	}

	response.Success(c, resp)
	return nil
}

func (h *GroupMemberHandler) MuteMember(c *gin.Context) error {
	var req types.MuteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}
	uid64, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}
	if req.Mute == nil {
		return response.NewError(http.StatusBadRequest, "mute 不能为空")
	}

	if err := h.GroupMemberService.MuteMember(
		c.Request.Context(),
		req.GroupId,
		int(uid64),
		req.TargetUserId,
		*req.Mute,
	); err != nil {
		if err.Error() == "无权限操作" {
			return response.NewError(http.StatusForbidden, err.Error())
		}
		return response.NewError(http.StatusBadRequest, err.Error())
	}
	response.Success(c, "ok")
	return nil
}

func (h *GroupMemberHandler) MuteAll(c *gin.Context) error {
	var req types.MuteAllRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}
	uid64, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}
	if req.Mute == nil {
		return response.NewError(http.StatusBadRequest, "mute 不能为空")
	}

	resp, err := h.GroupMemberService.SetMuteAll(
		c.Request.Context(),
		req.GroupId,
		int(uid64),
		*req.Mute,
	)
	if err != nil {
		if err.Error() == "无权限操作" {
			return response.NewError(http.StatusForbidden, err.Error())
		}
		return response.NewError(http.StatusBadRequest, err.Error())
	}
	response.Success(c, resp)
	return nil
}
func (h *GroupMemberHandler) SetAdmin(c *gin.Context) error {
	var req types.SetAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}

	uid64, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}

	if req.Admin == nil {
		return response.NewError(http.StatusBadRequest, "admin 不能为空")
	}

	if err := h.GroupMemberService.SetAdmin(
		c.Request.Context(),
		req.GroupId,
		int(uid64),
		req.TargetUserId,
		*req.Admin,
	); err != nil {
		if err.Error() == "无权限操作" {
			return response.NewError(http.StatusForbidden, err.Error())
		}
		return response.NewError(http.StatusBadRequest, err.Error())
	}

	response.Success(c, "ok")
	return nil
}
func (h *GroupMemberHandler) TransferOwner(c *gin.Context) error {
	var req types.TransferOwnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}

	uid64, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}

	resp, err := h.GroupMemberService.TransferOwner(
		c.Request.Context(),
		req.GroupId,
		int(uid64),
		req.NewOwnerId,
	)
	if err != nil {
		if err.Error() == "无权限操作" {
			return response.NewError(http.StatusForbidden, err.Error())
		}
		return response.NewError(http.StatusBadRequest, err.Error())
	}

	response.Success(c, resp)
	return nil
}
