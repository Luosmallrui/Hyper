package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	Service service.IMessageService
	Config  *config.Config
}

func (h *MessageHandler) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(h.Config.Jwt.Secret))
	message := r.Group("/v1/message")
	message.Use(authorize)
	message.POST("/send", context.Wrap(h.SendMessage))
	message.GET("/list", context.Wrap(h.ListMessages))
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

func (h *MessageHandler) ListMessages(c *gin.Context) error {
	userId, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(401, "未登录")
	}

	peerId, _ := strconv.ParseUint(c.Query("peer_id"), 10, 64)
	cursor, _ := strconv.ParseInt(c.Query("cursor"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if peerId == 0 {
		return response.NewError(400, "peer_id 不能为空")
	}

	list, err := h.Service.ListMessages(
		c.Request.Context(),
		uint64(userId),
		peerId,
		cursor,
		limit,
	)
	if err != nil {
		return response.NewError(500, "拉取消息失败")
	}

	response.Success(c, gin.H{
		"list": list,
		"next_cursor": func() int64 {
			if len(list) > 0 {
				return list[0].Time // 最老一条
			}
			return 0
		}(),
	})

	return nil
}
