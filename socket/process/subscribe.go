package process

import (
	"Hyper/dao/cache"
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

type MessageSubscribe struct {
	Redis          *redis.Client
	MqConsumer     rocketmq.PushConsumer
	DB             *gorm.DB
	MessageService service.IMessageService
	MessageStorage *cache.MessageStorage
	SessionService service.ISessionService
	UnreadStorage  *cache.UnreadStorage
}

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

func (m *MessageSubscribe) Setup(ctx context.Context) error {
	log.Printf("[MQ] 正在启动消息消费者，ServerID: %s", server.GetServerId())
	err := m.MqConsumer.Subscribe(types.ImTopicChat, consumer.MessageSelector{}, m.handleMessage)
	if err != nil {
		return fmt.Errorf("subscribe topic error: %w", err)
	}

	if err := m.MqConsumer.Start(); err != nil {
		log.Printf("[MQ Warning] Start message consumer failed: %v. Will retry in background...", err)
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					if err := m.MqConsumer.Start(); err == nil {
						log.Printf("[MQ] Message consumer started successfully in background")
						return
					}
				}
			}
		}()
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
			continue
		}
		b, _ := json.Marshal(imMsg)
		log.Printf("[MQ] 解析成功: %s", string(b))
		extBytes, _ := json.Marshal(imMsg.Ext)

		if imMsg.SessionType == types.SessionTypeSingle {
			immsg := &models.ImSingleMessage{
				Id:          imMsg.Id,
				SessionHash: imMsg.SessionHash,
				SessionId:   imMsg.SessionID,
				SenderId:    imMsg.SenderID,
				TargetId:    imMsg.TargetID,
				MsgType:     imMsg.MsgType,
				Content:     imMsg.Content,
				ParentMsgId: imMsg.ParentMsgID,
				Status:      types.MsgStatusSuccess,
				CreatedAt:   time.Now().Unix(),
				Ext:         string(extBytes),
				UpdatedAt:   time.Now(),
			}
			if err := m.MessageService.SaveMessage(immsg); err != nil {
				log.Printf("[MQ] 消息入库失败: %v", err)
				// 重要：如果是数据库临时不可用，返回 Suspend，让 MQ 稍后重试，保证顺序性
				//return consumer.SuspendCurrentQueueAMoment, err
			}
			if err := m.SessionService.UpdateSingleSession(ctx, &imMsg); err != nil {
				log.Printf("update session error: %v", err)
				//return consumer.SuspendCurrentQueueAMoment, err
			}
			go func(imMsg types.Message) {
				m.updateUserCache(ctx, &imMsg)
				m.dispatchMessage(&imMsg)
			}(imMsg)
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
	msg.Status = types.MsgStatusSuccess

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

func (m *MessageSubscribe) updateUserCache(ctx context.Context, msg *types.Message) {
	// 1. 确定接收者 ID
	// 这里的逻辑要小心：如果是单聊，接收者是 TargetID；
	// 如果是自己发给自己的同步消息，则不需要增加未读数
	if msg.SenderID == msg.TargetID {
		return
	}

	receiverID := int(msg.TargetID)
	senderID := int(msg.SenderID)

	// 2. 开启管道
	pipe := m.Redis.Pipeline()

	// --- A. 更新接收方的未读数 ---
	// Key 设计: unread:{user_id} -> Hash { "1_single_1001": 5 }
	// 表示：用户 receiverID 收到来自 1001 的单聊消息，未读数为 5
	m.UnreadStorage.PipeIncr(ctx, pipe, receiverID, types.SessionTypeSingle, senderID)

	// --- B. 更新双方的会话列表摘要 (Last Message) ---
	// 无论是发送方还是接收方，会话列表展示的最后一条消息必须是最新的
	fmt.Println(msg.Content, 2131242141)
	summary := &cache.LastCacheMessage{
		Content:   truncateContent(msg.Content, 50), // 截断内容防止浪费内存
		Timestamp: msg.Timestamp,
	}

	m.MessageStorage.Set(ctx, types.SessionTypeSingle, receiverID, senderID, summary)
	m.MessageStorage.Set(ctx, types.SessionTypeSingle, senderID, receiverID, summary)

	// 3. 执行管道
	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Printf("[Cache] 更新用户 %d 的缓存失败: %v", receiverID, err)
	}
}

func truncateContent(content string, maxLen int) string {
	runes := []rune(content)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	return content
}
