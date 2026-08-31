package process

import (
	"Hyper/pkg/log"
	"Hyper/pkg/server"
	"Hyper/pkg/socket"
	"Hyper/service"
	"Hyper/types"
	"context"
	"encoding/json"

	rmq_client "github.com/apache/rocketmq-clients/golang/v5"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type NoticeSubscribe struct {
	Redis *redis.Client

	ConnectService service.IClientConnectService
}

func (m *NoticeSubscribe) Init() error {

	return nil
}

func (m *NoticeSubscribe) Setup(ctx context.Context) error {

	return nil
}

func (m *NoticeSubscribe) handleSystem(ctx context.Context, msgs *rmq_client.MessageView) error {
	var event types.SystemMessage
	if err := json.Unmarshal(msgs.GetBody(), &event); err != nil {
		// 消息体解析失败，重试无意义，直接返回 nil（上层 Ack）而不是用零值继续推送
		log.L.Error("unmarshal msg error", zap.Error(err))
		return nil
	}

	switch event.Type {
	case "follow":
		var data types.FollowPayload
		if err := json.Unmarshal(event.Data, &data); err != nil {
			log.L.Error("unmarshal msg error", zap.Error(err))
			return nil
		}
		m.handleFollowNotice(ctx, &data)
	case "platform_message":
		var data types.PlatformMessagePayload
		if err := json.Unmarshal(event.Data, &data); err != nil {
			log.L.Error("unmarshal platform message error", zap.Error(err))
			return nil
		}
		m.handlePlatformMessage(ctx, &data)
	case "user_notification":
		var data types.NotificationPayload
		if err := json.Unmarshal(event.Data, &data); err != nil {
			log.L.Error("unmarshal user notification error", zap.Error(err))
			return nil
		}
		m.handleUserNotification(ctx, &data)
	}

	return nil
}

func (m *NoticeSubscribe) handleFollowNotice(ctx context.Context, data *types.FollowPayload) {
	// 获取被关注者在当前节点的连接
	sid := server.GetServerId()
	channel := socket.Session.Chat.Name()

	// log.Printf("[排查调试] 准备推送关注消息. ServerID: %s, Channel: %s, 目标用户ID: %d", sid, channel, data.TargetId)

	// 查找被关注者是否在当前服务器在线
	cids, err := m.ConnectService.GetUidFromClientIds(ctx, sid, channel, data.TargetId)
	if err != nil {
		log.L.Error("GetUidFromClientIds error", zap.Error(err))
		return
	}

	// log.Printf("[排查调试] 在线连接查询结果: 找到 %d 个连接 (ClientIDs: %v)", len(cids), cids)

	if len(cids) == 0 {
		return
	}

	// 构造推送消息
	content := socket.NewSenderContent().
		SetReceive(cids...).
		SetMessage("notice.follow", data)

	// 推送消息
	socket.Session.Chat.Write(content)
	// log.Printf("[排查调试] 消息已推送到 Socket 写入队列")
}

func (m *NoticeSubscribe) handlePlatformMessage(ctx context.Context, data *types.PlatformMessagePayload) {
	sid := server.GetServerId()
	channel := socket.Session.Chat.Name()
	cids, err := m.ConnectService.GetUidFromClientIds(ctx, sid, channel, int(data.TargetID))
	if err != nil {
		log.L.Error("GetUidFromClientIds error", zap.Error(err))
		return
	}
	if len(cids) == 0 {
		return
	}
	content := socket.NewSenderContent().
		SetReceive(cids...).
		SetMessage("notice.platform_message", data)
	socket.Session.Chat.Write(content)
}

// handleUserNotification 用户收件箱通知的在线实时提醒（离线靠收件箱接口补拉）
func (m *NoticeSubscribe) handleUserNotification(ctx context.Context, data *types.NotificationPayload) {
	sid := server.GetServerId()
	channel := socket.Session.Chat.Name()
	cids, err := m.ConnectService.GetUidFromClientIds(ctx, sid, channel, int(data.UserID))
	if err != nil {
		log.L.Error("GetUidFromClientIds error", zap.Error(err))
		return
	}
	if len(cids) == 0 {
		return
	}
	content := socket.NewSenderContent().
		SetReceive(cids...).
		SetMessage("notice.new", data)
	socket.Session.Chat.Write(content)
}
