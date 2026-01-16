package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"fmt"
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
	group := r.Group("/groupmember")
	group.POST("/invite", authorize, context.Wrap(hm.InviteMember)) //邀请成员
	group.POST("/kick", authorize, context.Wrap(hm.KickMember))
}

func (h *GroupMemberHandler) InviteMember(c *gin.Context) error {
	var req types.InviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: ",
		})
		return err
	}
	//执行邀请的用户ID
	//从上下文中获取用户ID
	UserIdval, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	UserId := int(UserIdval)
	//req.GroupId,req.UserIds,请求参数前者是群ID，后者是被邀请的用户ID列表
	resp, err := h.GroupMemberService.InviteMembers(c, req.GroupId, req.UserIds, UserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "邀请成员失败: " + err.Error(),
		})
		return err
	}
	response.Success(c, gin.H{
		"invited_members": resp,
	})

	return nil
}

func (h *GroupMemberHandler) KickMember(c *gin.Context) error {
	var req types.KickMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: ",
		})
		return err
	}

	userId := c.GetInt("user_id")

	v, ok := c.Get("user_id")
	fmt.Printf("[DEBUG] ctx user_id raw=%v ok=%v type=%T GetInt=%d\n", v, ok, v, userId)

	err := h.GroupMemberService.KickMember(c, req.GroupId, req.KickedUserId, userId)

	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}

	response.Success(c, "踢出成员成功")
	return nil
}
