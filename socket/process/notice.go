package process

import (
	"Hyper/pkg/log"
	"Hyper/pkg/server"
	"Hyper/pkg/socket"
	"Hyper/service"
	"Hyper/types"
	"context"
	"encoding/json"

	"github.com/apache/rocketmq-client-go/v2/consumer"
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

func (m *NoticeSubscribe) handleSystem(ctx context.Context, msgs *rmq_client.MessageView) (consumer.ConsumeResult, error) {
	var event types.SystemMessage
	if err := json.Unmarshal(msgs.GetBody(), &event); err != nil {
		log.L.Error("unmarshal msg error", zap.Error(err))
	}

	switch event.Type {
	case "follow":
		var data types.FollowPayload
		if err := json.Unmarshal(event.Data, &data); err != nil {
			log.L.Error("unmarshal msg error", zap.Error(err))
		}
		m.handleFollowNotice(ctx, &data)
	}

	return consumer.ConsumeSuccess, nil
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
