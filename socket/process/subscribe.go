package process

import (
	"Hyper/models"
	"Hyper/pkg/server"
	"Hyper/rpc/kitex_gen/im/push"
	"Hyper/rpc/kitex_gen/im/push/pushservice"
	"Hyper/service"
	"Hyper/types"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/cloudwego/kitex/client"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var clientCache sync.Map // map[string]pushservice.Client

func (m *MessageSubscribe) getRpcClient(addr string) (pushservice.Client, error) {
	if cli, ok := clientCache.Load(addr); ok {
		return cli.(pushservice.Client), nil
	}
	// 创建新 Client 开启多路复用
	newCli, err := pushservice.NewClient("im_push_service", client.WithHostPorts(addr))
	if err != nil {
		return nil, err
	}
	clientCache.Store(addr, newCli)
	return newCli, nil
}

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

		if imMsg.SessionType == types.SingleChat {
			immsg := &models.ImSingleMessage{
				Id:          imMsg.Id,
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

	}

	return consumer.ConsumeSuccess, nil
}

type RouterKey struct{}

func (RouterKey) UserLocation(uid int) string {
	return fmt.Sprintf("ws:user:location:%d", uid)
}

func (RouterKey) UserClients(sid, channel string, uid int) string {
	return fmt.Sprintf("ws:%s:%s:user:%d", sid, channel, uid)
}

func (RouterKey) ClientMap(sid, channel string) string {
	return fmt.Sprintf("ws:%s:%s:client", sid, channel)
}

func (m *MessageSubscribe) dispatchMessage(msg *types.Message) {
	ctx := context.Background()
	key := RouterKey{}
	uid := int(msg.TargetID)

	sids, err := m.Redis.HKeys(ctx, key.UserLocation(uid)).Result()
	if err != nil || len(sids) == 0 {
		log.Printf("[MQ] 用户 %d 不在线", uid)
		return
	}

	payload, _ := json.Marshal(msg)

	for _, sid := range sids {
		userKey := key.UserClients(sid, msg.Channel, uid)
		cids, _ := m.Redis.SMembers(ctx, userKey).Result()

		if len(cids) == 0 {
			// 如果某台机器上已经没客户端了，顺手清理一下 Hash
			m.Redis.HDel(ctx, key.UserLocation(uid), sid)
			continue
		}

		// 获取 RPC Client 并推送
		cli, err := m.getRpcClient(sid)
		if err != nil {
			continue
		}

		for _, cidStr := range cids {
			cid, _ := strconv.ParseInt(cidStr, 10, 64)
			_, err := cli.PushToClient(ctx, &push.PushRequest{
				Cid:     cid,
				Uid:     int32(uid),
				Payload: string(payload),
				Event:   "chat",
			})
			if err != nil {
				log.Printf("[RPC] 目标服务不可达，正在清理脏路由: %s", sid)

				// 立即从该用户的 Location Hash 中移除该 sid
				m.Redis.HDel(ctx, key.UserLocation(uid), sid)

				// 同时清理 RPC 客户端缓存
				clientCache.Delete(sid)
			}
		}
	}
}
