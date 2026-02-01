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

type GroupHandler struct {
	Config       *config.Config
	GroupService service.IGroupService
}

func (h *GroupHandler) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(h.Config.Jwt.Secret))
	group := r.Group("/v1/group")
	group.POST("/create", authorize, context.Wrap(h.CreateGroup))   //创建群
	group.POST("/dismiss", authorize, context.Wrap(h.DismissGroup)) //解散群
	//group.POST("/muteall", authorize, context.Wrap(h.MuteAllMembers))                    //开启全员禁言
	//group.POST("/unmuteall", authorize, context.Wrap(h.UnMuteAllMembers))                //关闭全员禁言
	group.POST("/update-name", authorize, context.Wrap(h.UpdateGroupName))               //修改群名称
	group.POST("/update-avatar", authorize, context.Wrap(h.UpdateGroupAvatar))           //修改群头像
	group.POST("/update-description", authorize, context.Wrap(h.UpdateGroupDescription)) //修改群描述

}

// 创建群
func (h *GroupHandler) CreateGroup(c *gin.Context) error {
	var req types.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}
	userId := c.GetInt("user_id")
	//调用服务层创建群
	group, err := h.GroupService.CreateGroup(c.Request.Context(), &req, userId)

	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}

	//构建响应
	resp := types.CreateGroupResponse{
		Id:          group.Id,
		Name:        group.Name,
		Avatar:      group.Avatar,
		OwnerId:     group.OwnerId,
		MemberCount: group.MemberCount,
		CreatedAt:   group.CreatedAt.Format("2006-01-02 15:04:05"),
		SessionId:   group.SessionId,
	}
	response.Success(c, &resp)
	return nil
}

// 解散群
func (h *GroupHandler) DismissGroup(c *gin.Context) error {
	var req types.DismissGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}
	userId := c.GetInt("user_id")

	//调用服务层解散群
	err := h.GroupService.DismissGroup(c, req.GroupId, userId)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, "群解散成功")
	return nil
}

////开启全员禁言
//
//func (h *GroupHandler) MuteAllMembers(c *gin.Context) error {
//	var req types.MuteAllRequest
//	if err := c.ShouldBindJSON(&req); err != nil {
//		return response.NewError(http.StatusBadRequest, err.Error())
//	}
//	userId := c.GetInt("user_id")
//
//	//调用服务层开启全员禁言
//	err := h.GroupService.MuteAllMembers(c, userId, req.GroupId)
//	if err != nil {
//		return response.NewError(http.StatusInternalServerError, err.Error())
//	}
//	response.Success(c, "全员禁言已开启")
//	return nil
//}
//
//// 关闭全员禁言
//func (h *GroupHandler) UnMuteAllMembers(c *gin.Context) error {
//	var req types.UnMuteAllRequest
//	if err := c.ShouldBindJSON(&req); err != nil {
//		return response.NewError(http.StatusBadRequest, err.Error())
//	}
//	userId := c.GetInt("user_id")
//
//	//调用服务层关闭全员禁言
//	err := h.GroupService.UnMuteAllMembers(c, req.GroupId, userId)
//	if err != nil {
//		return response.NewError(http.StatusInternalServerError, err.Error())
//	}
//	response.Success(c, "全员禁言已关闭")
//	return nil
//}

// 修改群信息（参考小红书，群聊的消息修改是一个一个的，默认群主才能修改）
func (h *GroupHandler) UpdateGroupName(c *gin.Context) error {
	var req types.UpdateGroupNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}
	userId := c.GetInt("user_id")

	//调用服务层修改群名称
	err := h.GroupService.UpdateGroupName(c, req.GroupId, userId, &req)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, "群名称已更新")
	return nil
}
func (h *GroupHandler) UpdateGroupAvatar(c *gin.Context) error {
	var req types.UpdateGroupAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}
	userId := c.GetInt("user_id")

	//调用服务层修改群头像
	err := h.GroupService.UpdateGroupAvatar(c, req.GroupId, userId, &req)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, "群头像已更新")
	return nil
}

func (h *GroupHandler) UpdateGroupDescription(c *gin.Context) error {
	var req types.UpdateGroupDescriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}
	userId := c.GetInt("user_id")

	//调用服务层修改群描述
	err := h.GroupService.UpdateGroupDescription(c, req.GroupId, userId, &req)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, "群描述已更新")
	return nil
}

//发布群公告
