package handler

import (
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	Service service.IMessageService
}

func (h *MessageHandler) RegisterRouter(r gin.IRouter) {
	message := r.Group("/api/message")
	message.POST("/send", context.Wrap(h.SendMessage))
	message.GET("/list", context.Wrap(h.GetRecentMessages))
}

// POST /api/message/send
func (h *MessageHandler) SendMessage(c *gin.Context) error {
	var msg types.Message
	if err := c.ShouldBindJSON(&msg); err != nil {
		return response.NewError(500, err.Error())
	}

	if err := h.Service.SendMessage(&msg); err != nil {
		return response.NewError(500, err.Error())
	}
	response.Success(c, msg)
	return nil
}

func (h *MessageHandler) GetRecentMessages(c *gin.Context) error {
	targetID := c.Query("target_id")
	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)

	msgs, err := h.Service.GetRecentMessages(targetID, limit)
	if err != nil {
		return response.NewError(500, err.Error())
	}
	response.Success(c, msgs)
	return nil
}
