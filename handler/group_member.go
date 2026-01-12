package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
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
	group := r.Group("/groupmember")
	group.POST("/invit", authorize, context.Wrap(hm.InviteMember)) //邀请成员
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
	userId := c.GetInt("user_id")
	//req.GroupId,req.UserIds,请求参数前者是群ID，后者是被邀请的用户ID列表
	resp, err := h.GroupMemberService.InviteMembers(c, req.GroupId, req.UserIds, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "邀请成员失败: " + err.Error(),
		})
		return err
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "邀请成员成功",
		"data":    resp,
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

	err := h.GroupMemberService.KickMember(c, req.GroupId, userId, req.KickedUserId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "踢出成员失败: " + err.Error(),
		})
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "踢出成员成功",
	})
	return nil
}
