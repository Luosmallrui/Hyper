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

type NotificationHandler struct {
	Config              *config.Config
	NotificationService service.INotificationService
}

func (h *NotificationHandler) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(h.Config.Jwt.Secret))
	notificationGroup := r.Group("/v1/notifications")
	notificationGroup.GET("", authorize, context.Wrap(h.List))
	notificationGroup.GET("/unread-count", authorize, context.Wrap(h.UnreadCount))
	notificationGroup.POST("/read", authorize, context.Wrap(h.MarkRead))
	notificationGroup.POST("/read-all", authorize, context.Wrap(h.MarkAllRead))
}

func (h *NotificationHandler) currentUserID(c *gin.Context) int {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "Bearer debug-mode" {
		return 6 // Debug 模式下使用固定用户ID
	}
	return c.GetInt("user_id")
}

// List GET /v1/notifications?type=&page=&size=
func (h *NotificationHandler) List(c *gin.Context) error {
	var req types.NotificationListQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "请求参数格式错误")
	}
	data, err := h.NotificationService.List(c.Request.Context(), int64(h.currentUserID(c)), req.Type, req.Page, req.Size)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, data)
	return nil
}

// UnreadCount GET /v1/notifications/unread-count
func (h *NotificationHandler) UnreadCount(c *gin.Context) error {
	data, err := h.NotificationService.UnreadCount(c.Request.Context(), int64(h.currentUserID(c)))
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, data)
	return nil
}

// MarkRead POST /v1/notifications/read {ids:[...]}
func (h *NotificationHandler) MarkRead(c *gin.Context) error {
	var req types.MarkNotificationReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "请求参数格式错误")
	}
	if err := h.NotificationService.MarkRead(c.Request.Context(), int64(h.currentUserID(c)), req.IDs); err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, nil)
	return nil
}

// MarkAllRead POST /v1/notifications/read-all {type?}
func (h *NotificationHandler) MarkAllRead(c *gin.Context) error {
	// body 可选：{"type":"system"} 限定类型，空 body 表示全部已读
	var req struct {
		Type string `json:"type"`
	}
	_ = c.ShouldBindJSON(&req)
	if err := h.NotificationService.MarkAllRead(c.Request.Context(), int64(h.currentUserID(c)), req.Type); err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, nil)
	return nil
}
