package handler

import (
	"Hyper/dao"
	"net/http"
	"time"

	"Hyper/models"
	"Hyper/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

/*
 * 1. WebSocket 升级器
 */
var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

/*
 * 2. 在线用户表
 * key   = user_id
 * value = websocket 连接
 */

type OnlineUser struct {
	Conn     *websocket.Conn
	LastPing int64 // Unix 秒
}

// 清理假在线
var onlineUsers = make(map[string]*OnlineUser)

func StartOnlineChecker() {
	go func() {
		for {
			time.Sleep(30 * time.Second)
			now := time.Now().Unix()

			for uid, u := range onlineUsers {
				if now-u.LastPing > 60 {
					u.Conn.Close()
					delete(onlineUsers, uid)
				}
			}
		}
	}()
}

/*
 * 3. WSHandler
 */

type WSHandler struct {
	MessageService     *service.MessageService
	MessageReadService *service.MessageReadService
	GroupDAO           *dao.GroupDAO
}

func NewWSHandler(
	ms *service.MessageService,
	mrs *service.MessageReadService,
	gd *dao.GroupDAO,
) *WSHandler {
	return &WSHandler{
		MessageService:     ms,
		MessageReadService: mrs,
		GroupDAO:           gd,
	}
}

/*
 * 4. 注册路由
 */
func (h *WSHandler) RegisterRouter(r gin.IRouter) {
	r.GET("/ws", h.HandleWS)
}

// 系统消息
// PushSystemMessage 只负责 WS 推送（不落库）
func (h *WSHandler) PushSystemMessage(userID string, content string) {
	if u, ok := onlineUsers[userID]; ok {
		_ = u.Conn.WriteJSON(gin.H{
			"from":         "system",
			"session_type": 3,
			"content":      content,
		})
	}
}

// SendSystemMessage 系统消息统一出口（落库 + 推送）
func (h *WSHandler) SendSystemMessage(userID string, content string) {
	// 1. 落库
	_ = h.MessageService.SendSystemMessage(userID, content)

	// 2. 在线则实时推送
	h.PushSystemMessage(userID, content)
}

/*
 * 5. 处理 WebSocket 连接
 */
func (h *WSHandler) HandleWS(c *gin.Context) {
	// ===== Step 1: 获取 user_id =====
	userID := c.Query("user_id")
	if userID == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	// ===== Step 2: 升级为 WebSocket =====
	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	// ===== Step 3: 用户上线 =====
	onlineUsers[userID] = &OnlineUser{
		Conn:     conn,
		LastPing: time.Now().Unix(),
	}

	// ===== Step 3.5: 离线消息补发 =====
	msgs, err := h.MessageService.PullOfflineMessages(userID)
	if err == nil {
		for _, msg := range msgs {
			_ = conn.WriteJSON(gin.H{
				"msg_id":       msg.MsgID,
				"from":         msg.SenderID,
				"session_type": msg.SessionType,
				"target_id":    msg.TargetID,
				"content":      msg.Content,
				"timestamp":    msg.Timestamp,
			})
		}
	}

	// ===== Step 4: 用户下线清理 =====
	defer func() {
		delete(onlineUsers, userID)
		conn.Close()
	}()

	// ===== Step 5: 循环接收消息（关键）=====
	for {
		var wsMsg struct {
			Type        string   `json:"type"`         // ping / message
			SessionType int      `json:"session_type"` // 1=私聊，2=群聊
			TargetID    string   `json:"target_id"`    // user_id 或 group_id
			Content     string   `json:"content"`
			MsgIDs      []string `json:"msg_ids"` //ack
		}

		// 1. 读取消息
		if err := conn.ReadJSON(&wsMsg); err != nil {
			// 客户端断开是正常行为
			return
		}
		// ===== 心跳处理 =====
		if wsMsg.Type == "ping" {
			if u, ok := onlineUsers[userID]; ok {
				u.LastPing = time.Now().Unix()
			}
			continue
		}
		// ===== ACK 处理 + 已读回执 =====
		if wsMsg.Type == "ack" {

			// ===== 私聊 ACK =====
			if wsMsg.SessionType == 1 {
				_ = h.MessageService.AckMessages(wsMsg.MsgIDs)

				// 已读回执（通知发送者）
				for _, msgID := range wsMsg.MsgIDs {
					msg, err := h.MessageService.GetMessageByID(msgID)
					if err != nil {
						continue
					}

					if u, ok := onlineUsers[msg.SenderID]; ok {
						_ = u.Conn.WriteJSON(gin.H{
							"type":   "read_receipt",
							"msg_id": msgID,
							"from":   userID,
						})
					}
				}
			}

			// ===== 群聊 ACK =====
			if wsMsg.SessionType == 2 {
				for _, msgID := range wsMsg.MsgIDs {
					_ = h.MessageReadService.MarkGroupRead(msgID, userID)
				}
			}

			continue
		}
		// 2. 构造数据库消息
		dbMsg := &models.Message{
			MsgID:       uuid.NewString(),
			SenderID:    userID,
			TargetID:    wsMsg.TargetID,
			SessionType: wsMsg.SessionType,
			MsgType:     1,
			Content:     wsMsg.Content,
			Timestamp:   time.Now().UnixMilli(),
			Status:      1,
			Ext:         "{}",
		}

		switch wsMsg.SessionType {

		case 1:
			// ===== 私聊 =====
			_ = h.MessageService.SendMessage(dbMsg)

			if u, ok := onlineUsers[wsMsg.TargetID]; ok {
				_ = u.Conn.WriteJSON(gin.H{
					"msg_id":  dbMsg.MsgID,
					"from":    userID,
					"content": wsMsg.Content,
				})
			}

		case 2:
			// ===== 群聊 =====
			_ = h.MessageService.SendMessage(dbMsg)

			members, err := h.GroupDAO.GetGroupMembers(wsMsg.TargetID)
			if err != nil {
				continue
			}

			for _, uid := range members {
				if uid == userID {
					continue
				}
				if u, ok := onlineUsers[uid]; ok {
					_ = u.Conn.WriteJSON(gin.H{
						"msg_id":  dbMsg.MsgID,
						"from":    userID,
						"group":   wsMsg.TargetID,
						"content": wsMsg.Content,
					})
				}
			}
		}
	}
}
