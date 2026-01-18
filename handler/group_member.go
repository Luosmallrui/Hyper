package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GroupMemberHandler struct {
	Config             *config.Config
	GroupMemberService service.IGroupMemberService
}

func (h *GroupMemberHandler) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(h.Config.Jwt.Secret))
	group := r.Group("/v1/groupmember")
	group.POST("/invite", authorize, context.Wrap(h.InviteMember)) //邀请成员
	group.POST("/kick", authorize, context.Wrap(h.KickMember))
	group.GET("/list", authorize, context.Wrap(h.ListMembers))

}

func (h *GroupMemberHandler) InviteMember(c *gin.Context) error {
	var req types.InviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}
	userId := c.GetInt("user_id")

	resp, err := h.GroupMemberService.InviteMembers(c, req.GroupId, req.InvitedUserIds, userId)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}

	response.Success(c, resp)
	return nil
}

func (h *GroupMemberHandler) KickMember(c *gin.Context) error {
	var req types.KickMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}

	userId := c.GetInt("user_id")

	if userId == req.KickedUserId {
		return response.NewError(http.StatusBadRequest, "不能踢出自己")
	}
	err := h.GroupMemberService.KickMember(c, req.GroupId, req.KickedUserId, userId)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (h *GroupMemberHandler) ListMembers(c *gin.Context) error {
	// 1) 解析 group_id
	gidStr := c.Query("group_id")
	if gidStr == "" {
		return response.NewError(http.StatusBadRequest, "group_id 不能为空")
	}
	gid, err := strconv.Atoi(gidStr)
	if err != nil || gid <= 0 {
		return response.NewError(400, "group_id 参数错误")
	}

	// 2) 获取当前登录用户
	uid64, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}
	uid := int(uid64)

	// 3) 调 service
	members, err := h.GroupMemberService.ListMembers(c, gid, uid)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}

	// 4) 返回
	response.Success(c, types.GroupMemberListResponse{
		Members: members,
	})
	response.Success(c, "踢出成功")
	return nil
}
