package process

import (
	"Hyper/pkg/log"
	"Hyper/pkg/server"
	"Hyper/pkg/socket"
	"Hyper/service"
	"Hyper/types"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type NoticeSubscribe struct {
	Redis          *redis.Client
	MqConsumer     rocketmq.PushConsumer
	ConnectService service.IClientConnectService
}

func (m *NoticeSubscribe) Init() error {
	// 所有的 Subscribe 都在这里，由外部 Start 方法同步调用
	err := m.MqConsumer.Subscribe("hyper_system_messages", consumer.MessageSelector{}, m.handleMessage)
	if err != nil {
		return fmt.Errorf("subscribe topic error: %w", err)
	}
	return nil
}

func (m *NoticeSubscribe) Setup(ctx context.Context) error {
	log.L.Info(fmt.Sprintf("[MQ] 正在启动notice消费者，ServerID: %s", server.GetServerId()))
	err := m.MqConsumer.Subscribe("hyper_system_messages", consumer.MessageSelector{}, m.handleMessage)
	if err != nil {
		return fmt.Errorf("subscribe topic error: %w", err)
	}

	if err := m.MqConsumer.Start(); err != nil {
		log.L.Error("start mq noteice consumer error", zap.Error(err))
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					if err := m.MqConsumer.Start(); err == nil {
						log.L.Info("[MQ Warning] Start notice consumer successfully")
						return
					}
				}
			}
		}()
	}

	go func() {
		<-ctx.Done()
		log.L.Info("[MQ] 正在关闭Notice消费者...")
		m.MqConsumer.Shutdown()
	}()

	return nil
}

func (m *NoticeSubscribe) handleMessage(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	for _, msg := range msgs {
		fmt.Println("Received message:", string(msg.Body))

		var event types.SystemMessage
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.L.Error("unmarshal msg error", zap.Error(err))
			continue
		}

		switch event.Type {
		case "follow":
			var data types.FollowPayload
			if err := json.Unmarshal(event.Data, &data); err != nil {
				log.L.Error("unmarshal msg error", zap.Error(err))
				continue
			}
			m.handleFollowNotice(ctx, &data)
		}
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
