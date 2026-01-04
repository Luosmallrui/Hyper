package process

import (
	"Hyper/pkg/server"
	"Hyper/types"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type MessageSubscribe struct {
	Redis      *redis.Client
	MqConsumer rocketmq.PushConsumer
	DB         *gorm.DB
}

func (m *MessageSubscribe) Setup(ctx context.Context) error {
	log.Printf("[MQ] 正在启动消息消费者，ServerID: %d", server.GetServerId())
	err := m.MqConsumer.Subscribe(types.ImTopicChat, consumer.MessageSelector{}, m.handleMessage)
	if err != nil {
		return fmt.Errorf("subscribe topic error: %w", err)
	}

	if err := m.MqConsumer.Start(); err != nil {
		return fmt.Errorf("start consumer error: %w", err)
	}

	go func() {
		<-ctx.Done()
		log.Println("[MQ] 正在关闭消费者...")
		m.MqConsumer.Shutdown()
	}()

	return nil
}

func (m *MessageSubscribe) handleMessage(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	for _, msg := range msgs {
		// 1. 反序列化
		var imMsg types.Message
		if err := json.Unmarshal(msg.Body, &imMsg); err != nil {
			log.Printf("[MQ] 解析消息失败: %v, Body: %s", err, string(msg.Body))
			continue // 解析失败的消息通常直接跳过，避免阻塞队列
		}
		b, _ := json.Marshal(imMsg)
		log.Printf("[MQ] 解析成功: %s", string(b))

		// 2. 数据库幂等校验 (防止 Consumer 层面重复消费)
		// 即使 API 层有 Redis 去重，Consumer 也要通过数据库唯一索引或 Redis 再次确认
		// 如果数据库 im_single_messages 的 msg_id 是主键，Create 报错则说明已存在

		// 3. 持久化到 MySQL
		//if err := m.saveMessage(&imMsg); err != nil {
		//	log.Printf("[MQ] 消息入库失败: %v", err)
		//	// 重要：如果是数据库临时不可用，返回 Suspend，让 MQ 稍后重试，保证顺序性
		//	return consumer.SuspendCurrentQueueAMoment, err
		//}

		//// 4. 异步处理会话更新 (不阻塞主消费流程)
		//// 包括更新未读数、最后一条消息摘要等
		//go m.updateConversation(&imMsg)
		//
		//// 5. 实时推送分发
		//m.dispatchMessage(&imMsg)
	}

	return consumer.ConsumeSuccess, nil
}
