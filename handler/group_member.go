package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GroupMemberHandler struct {
	Config             *config.Config
	GroupMemberService service.IGroupMemberService
}

func NewGroupMemberHandler(config *config.Config, groupMemberService service.IGroupMemberService) *GroupMemberHandler {
	return &GroupMemberHandler{
		GroupMemberService: groupMemberService,
		Config:             config,
	}
}

func (hm *GroupMemberHandler) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(hm.Config.Jwt.Secret))
	group := r.Group("/v1/groupmember")
	group.POST("/invite", authorize, context.Wrap(hm.InviteMember)) //邀请成员
	group.POST("/kick", authorize, context.Wrap(hm.KickMember))
}

func (h *GroupMemberHandler) InviteMember(c *gin.Context) error {
	var req types.InviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}
	userId := c.GetInt("user_id")

	//调用服务层邀请成员
	resp, err := h.GroupMemberService.InviteMembers(c, req.GroupId, req.InvitedUserIds, int(userId))
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
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
	err := h.GroupMemberService.KickMember(c, req.GroupId, req.KickedUserId, int(userId))
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, "踢出成功")
	return nil
}
