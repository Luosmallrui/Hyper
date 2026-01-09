package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/models"
	"Hyper/pkg/context"
	"Hyper/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GroupHandler struct {
	Config       *config.Config
	GroupService service.IGroupService
}

func NewGroupHandler(config *config.Config, groupService *service.GroupService) *GroupHandler {
	return &GroupHandler{
		GroupService: groupService,
		Config:       config,
	}
}

func (h *GroupHandler) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(h.Config.Jwt.Secret))
	group := r.Group("/api/group")
	group.POST("/create", authorize, context.Wrap(h.CreateGroup)) //创建群
}

// 创建群
func (h *GroupHandler) CreateGroup(c *gin.Context) error {
	var req models.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: ",
		})
		return err
	}

	//从上下文中获取用户ID
	UserId := c.GetInt("user_id")

	//调用服务层创建群
	group, err := h.GroupService.CreateGroup(c, &req, UserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "创建群失败: " + err.Error(),
		})
		return err
	}

	//构建响应
	resp := models.CreateGroupResponse{
		Id:          group.Id,
		Name:        group.Name,
		Avatar:      group.Avatar,
		OwnerId:     group.OwnerId,
		MemberCount: group.MemberCount,
		CreatedAt:   group.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建群成功",
		"data":    resp,
	})
	return nil
}
