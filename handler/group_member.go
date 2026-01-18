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

func NewGroupMemberHandler(config *config.Config, groupMemberService service.IGroupMemberService) *GroupMemberHandler {
	return &GroupMemberHandler{
		GroupMemberService: groupMemberService,
		Config:             config,
	}
}

func (hm *GroupMemberHandler) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(hm.Config.Jwt.Secret))
	group := r.Group("/groupmember")
	group.POST("/invite", authorize, context.Wrap(hm.InviteMember)) //邀请成员
	group.POST("/kick", authorize, context.Wrap(hm.KickMember))
	group.GET("/list", authorize, context.Wrap(hm.ListMembers))

}

func (h *GroupMemberHandler) InviteMember(c *gin.Context) error {
	var req types.InviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "请求参数错误")
	}

	uid, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}
	UserId := int(uid)

	resp, err := h.GroupMemberService.InviteMembers(c, req.GroupId, req.UserIds, UserId)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, "邀请成员失败: "+err.Error())
	}
	response.Success(c, gin.H{
		"invited_members": resp,
	})

	return nil
}

func (h *GroupMemberHandler) KickMember(c *gin.Context) error {
	var req types.KickMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "请求参数错误")
	}

	uid, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}
	UserId := int(uid)

	if err := h.GroupMemberService.KickMember(c, req.GroupId, req.KickedUserId, UserId); err != nil {
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
	return nil
}
