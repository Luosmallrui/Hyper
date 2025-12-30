package handler

import (
	"Hyper/models"
	"Hyper/pkg/context"
	"Hyper/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	Service *service.MessageService
}

func (h *MessageHandler) RegisterRouter(r gin.IRouter) {
	message := r.Group("/api/message")
	message.POST("/send", context.Wrap(h.SendMessage))
	message.GET("/list", context.Wrap(h.GetRecentMessages))
}

func NewMessageHandler(s *service.MessageService) *MessageHandler {
	return &MessageHandler{Service: s}
}

// POST /api/message/send
func (h *MessageHandler) SendMessage(c *gin.Context) error {
	var msg models.Message
	if err := c.ShouldBindJSON(&msg); err != nil {
		return err
	}

	if err := h.Service.SendMessage(&msg); err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{"message": "发送成功"})
	return nil
}

func (h *MessageHandler) GetRecentMessages(c *gin.Context) error {
	targetID := c.Query("target_id")
	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)

	msgs, err := h.Service.GetRecentMessages(targetID, limit)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{"messages": msgs})
	return nil
}
