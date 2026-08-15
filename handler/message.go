package handler

import (
	"Hyper/config"
	"Hyper/dao/cache"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Message struct {
	MessageService service.IMessageService
	FollowService  service.IFollowService
	UnreadStorage  *cache.UnreadStorage
	UserService    service.IUserService
	Config         *config.Config
	SessionService service.ISessionService
}

func (m *Message) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(m.Config.Jwt.Secret))
	message := r.Group("/v1/message")
	message.Use(authorize)
	message.POST("/send", context.Wrap(m.SendMessage))
	message.POST("/clear", context.Wrap(m.ClearMessageSession))
	message.DELETE("/:message_id", context.Wrap(m.DeleteMessage))
	message.GET("/list", context.Wrap(m.ListMessages))
}

func (m *Message) SendMessage(c *gin.Context) error {
	userId, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(401, "未登录")
	}
	var msg types.Message
	if err := c.ShouldBindJSON(&msg); err != nil {
		return response.NewError(500, err.Error())
	}
	msg.SenderID = userId

	if err := m.MessageService.SendMessage(&msg); err != nil {
		return response.NewError(500, err.Error())
	}
	response.Success(c, msg)
	return nil
}

// DeleteMessage hides one message for the current user only.
func (m *Message) DeleteMessage(c *gin.Context) error {
	userID, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(401, "未登录")
	}
	messageID, err := strconv.ParseInt(c.Param("message_id"), 10, 64)
	if err != nil || messageID <= 0 {
		return response.NewError(400, "message_id 非法")
	}
	var req types.DeleteMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(400, "参数格式错误")
	}
	if err := m.MessageService.DeleteMessageForUser(c.Request.Context(), uint64(userID), req.SessionType, req.PeerID, messageID); err != nil {
		if errors.Is(err, service.ErrGroupNotFound) {
			return response.NewError(404, "群不存在或已解散")
		}
		if errors.Is(err, service.ErrNotGroupMember) {
			return response.NewError(403, "你不在群内或已退群")
		}
		return response.NewError(400, err.Error())
	}
	response.Success(c, gin.H{"message_id": strconv.FormatInt(messageID, 10), "deleted": true})
	return nil
}

// ClearMessageSession hides all existing messages in a session for the
// current user while retaining the peer's history.
func (m *Message) ClearMessageSession(c *gin.Context) error {
	userID, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(401, "未登录")
	}
	var req types.ClearMessageSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(400, "参数格式错误")
	}
	if err := m.SessionService.ClearSessionMessages(c.Request.Context(), uint64(userID), req.SessionType, req.PeerID); err != nil {
		return response.NewError(400, err.Error())
	}
	response.Success(c, gin.H{"cleared": true})
	return nil
}

func (m *Message) ListMessages(c *gin.Context) error {
	userId := c.GetInt("user_id")

	peerId, _ := strconv.ParseUint(c.Query("peer_id"), 10, 64)
	sessionType, _ := strconv.Atoi(c.DefaultQuery("session_type", "1"))
	if sessionType != types.SessionTypeSingle && sessionType != types.GroupChatSessionTypeGroup {
		return response.NewError(400, "session_type 只能是 1(私聊) 或 2(群聊)")
	}
	cursor, _ := strconv.ParseInt(c.Query("cursor"), 10, 64)
	since, _ := strconv.ParseInt(c.Query("since"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if peerId == 0 {
		return response.NewError(400, "peer_id 不能为空")
	}

	list, err := m.MessageService.ListMessages(c.Request.Context(), uint64(userId), peerId, sessionType, cursor, since, limit)
	if err != nil {
		if sessionType == types.GroupChatSessionTypeGroup {
			switch {
			case errors.Is(err, service.ErrGroupNotFound):
				return response.NewError(404, "群不存在或已解散")
			case errors.Is(err, service.ErrNotGroupMember):
				return response.NewError(403, "你不在群内或已退群")
			}
		}
		return response.NewError(500, "拉取消息失败")
	}
	selfInfo := m.UserService.BatchGetUserInfo(c.Request.Context(), []uint64{uint64(userId)})

	follow, _ := m.FollowService.CheckFollowStatus(c.Request.Context(), uint64(userId), peerId)
	unreadNum, _ := m.SessionService.GetUnreadNum(c.Request.Context(), userId)
	resp := gin.H{
		"avatar":      "",
		"nickname":    "",
		"self_avatar": selfInfo[uint64(userId)].Avatar,
		"list":        list,
		"next_cursor": func() int64 {
			if len(list) > 0 {
				return list[0].Time // 最老一条
			}
			return 0
		}(),
		"unread_total": unreadNum,
		"is_followed":  follow,
	}

	if sessionType == types.SessionTypeSingle {
		userInfo := m.UserService.BatchGetUserInfo(c.Request.Context(), []uint64{peerId})
		resp["avatar"] = userInfo[peerId].Avatar
		resp["nickname"] = userInfo[peerId].Nickname
	}

	response.Success(c, resp)

	return nil
}
