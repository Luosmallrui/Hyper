package process

import (
	"Hyper/models"
	"Hyper/pkg/server"
	"Hyper/rpc/kitex_gen/im/push"
	"Hyper/rpc/kitex_gen/im/push/pushservice"
	"Hyper/types"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"Hyper/service"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/cloudwego/kitex/client"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type MessageSubscribe struct {
	Redis          *redis.Client
	MqConsumer     rocketmq.PushConsumer
	DB             *gorm.DB
	MessageService service.IMessageService
}

func (m *MessageSubscribe) Setup(ctx context.Context) error {
	log.Printf("[MQ] 正在启动消息消费者，ServerID: %s", server.GetServerId())
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

		immsg := &models.ImSingleMessage{
			Id:          imMsg.Id,
			ClientMsgId: imMsg.ClientMsgID,
			SessionHash: imMsg.SessionHash,
			SessionId:   imMsg.SessionID,
			SenderId:    imMsg.SenderID,
			TargetId:    imMsg.TargetID,
			MsgType:     imMsg.MsgType,
			Content:     imMsg.Content,
			ParentMsgId: imMsg.ParentMsgID,
			Status:      imMsg.Status,
			CreatedAt:   time.Now().Unix(),
			UpdatedAt:   time.Now(),
		}
		// 3. 持久化到 MySQL
		if err := m.MessageService.SaveMessage(immsg); err != nil {
			log.Printf("[MQ] 消息入库失败: %v", err)
			// 重要：如果是数据库临时不可用，返回 Suspend，让 MQ 稍后重试，保证顺序性
			return consumer.SuspendCurrentQueueAMoment, err
		}

		//// 4. 异步处理会话更新 (不阻塞主消费流程)
		//// 包括更新未读数、最后一条消息摘要等
		//go m.updateConversation(&imMsg)
		//
		//// 5. 实时推送分发
		m.dispatchMessage(&imMsg)
	}

	return consumer.ConsumeSuccess, nil
}

func (m *MessageSubscribe) dispatchMessage(msg *types.Message) {
	ctx := context.Background()

	// 1. 从 Redis 获取目标用户所在的服务器地址
	// 假设你存的 key 是 online:server:uid
	serverAddr, err := m.Redis.Get(ctx, fmt.Sprintf("online:server:%d", msg.TargetID)).Result()
	if err != nil || serverAddr == "" {
		log.Printf("[MQ] 用户 %d 不在线，放弃推送", msg.TargetID)
		return
	}

	// 2. 获取该用户的所有客户端 ID (Cid)
	// 对应你登录时的 SAdd(c.userKey(...), clientId)
	userKey := fmt.Sprintf("user:clients:%d", msg.TargetID)
	cids, _ := m.Redis.SMembers(ctx, userKey).Result()

	// 3. 构造 Kitex Client 调用目标服务器
	// 生产环境建议给每个 serverAddr 缓存一个 client 对象，不要每次都 NewClient
	cli, err := pushservice.NewClient("im_push_service", client.WithHostPorts(serverAddr))
	if err != nil {
		log.Printf("[RPC] 创建客户端失败: %v", err)
		return
	}

	// 4. 给该用户的所有设备发送消息
	for _, cidStr := range cids {
		cid, _ := strconv.ParseInt(cidStr, 10, 64)

		payload, _ := json.Marshal(msg)

		resp, err := cli.PushToClient(ctx, &push.PushRequest{
			Cid:     cid,
			Uid:     int32(msg.TargetID),
			Payload: string(payload),
			Event:   "chat", // 或者是 "like"
		})

		if err != nil || !resp.Success {
			log.Printf("[RPC] 推送给设备 %d 失败: %v", cid, err)
		}
	}
}
