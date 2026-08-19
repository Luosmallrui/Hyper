package handler

import (
	"Hyper/pkg/context"
	"Hyper/pkg/log"
	"Hyper/pkg/socket"
	"Hyper/pkg/socket/adapter"
	"Hyper/service"
	"Hyper/socket/handler/event"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ChatChannel struct {
	Storage service.IClientConnectService
	Event   *event.ChatEvent
}

// Conn 初始化连接
func (ch *ChatChannel) Conn(c *gin.Context) error {
	conn, err := adapter.NewWsAdapter(c.Writer, c.Request)
	if err != nil {
		log.L.Error("WebSocket connection error", zap.Error(err))
		return err
	}
	userID, err := context.GetUserID(c)
	if err != nil {
		log.L.Error("WebSocket get user id error", zap.Error(err))
		_ = conn.Close() // Upgrade 已成功，错误路径必须关闭连接，避免泄漏
		return err
	}
	log.L.Info("WebSocket connected", zap.Any("user_id", userID))

	if err := ch.NewClient(int(userID), conn); err != nil {
		log.L.Error("WebSocket init client error", zap.Error(err))
		_ = conn.Close()
		return err
	}
	return nil
}

func (ch *ChatChannel) NewClient(uid int, conn socket.IConn) error {
	return socket.NewClient(conn, &socket.ClientOption{
		Uid:     uid,
		Channel: socket.Session.Chat,
		Storage: ch.Storage,
		Buffer:  10,
	}, socket.NewEvent(
		// 连接成功回调
		socket.WithOpenEvent(ch.Event.OnOpen), //推送自己已经上线
		// 接收消息回调
		socket.WithMessageEvent(ch.Event.OnMessage), //发送消息的回调
		// 关闭连接回调
		socket.WithCloseEvent(ch.Event.OnClose), //下线的回调
	))
}
