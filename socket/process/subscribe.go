package process

import (
	"Hyper/dao/cache"
	"Hyper/models"
	"Hyper/pkg/server"
	"Hyper/rpc/kitex_gen/im/push"
	"Hyper/rpc/kitex_gen/im/push/pushservice"

	"Hyper/pkg/log"
	"Hyper/service"
	"Hyper/types"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/cloudwego/kitex/client"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
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
		log.L.Error("new push client error", zap.Error(err))
		return nil, err
	}
	clientCache.Store(addr, newCli)
	return newCli, nil
}

func (m *MessageSubscribe) Init() error {
	// 所有的 Subscribe 都在这里，由外部 Start 方法同步调用
	err := m.MqConsumer.Subscribe(types.ImTopicChat, consumer.MessageSelector{}, m.handleMessage)
	if err != nil {
		return fmt.Errorf("subscribe topic error: %w", err)
	}
	return nil
}
func (m *MessageSubscribe) Setup(ctx context.Context) error {
	log.L.Info("[MQ]MessageSubscribe 正在启动消息消费者", zap.String("serverId", server.GetServerId()))
	err := m.MqConsumer.Subscribe(types.ImTopicChat, consumer.MessageSelector{}, m.handleMessage)
	if err != nil {
		return fmt.Errorf("subscribe topic error: %w", err)
	}

	if err := m.MqConsumer.Start(); err != nil {
		log.L.Error("[MQ] start message consumer error", zap.Error(err))
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					if err := m.MqConsumer.Start(); err == nil {
						log.L.Info("[MQ] start message consumer success")
						return
					}
				}
			}
		}()
	}
	fmt.Println(11)

	go func() {
		<-ctx.Done()
		log.L.Info("[MQ] 正在关闭消费者...")
		m.MqConsumer.Shutdown()
	}()

	return nil
}

func (m *MessageSubscribe) handleMessage(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	for _, msg := range msgs {
		// 1. 反序列化
		var imMsg types.Message
		if err := json.Unmarshal(msg.Body, &imMsg); err != nil {
			log.L.Error("unmarshal msg error", zap.Error(err))
			continue
		}
		b, _ := json.Marshal(imMsg)
		log.L.Info("[MQ] 解析成功:", zap.Any("imMsg", imMsg), zap.String("body", string(b)))
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
				log.L.Error("[MQ] 消息入库失败", zap.Error(err))
				// 重要：如果是数据库临时不可用，返回 Suspend，让 MQ 稍后重试，保证顺序性
				//return consumer.SuspendCurrentQueueAMoment, err
			}
			if err := m.SessionService.UpdateSingleSession(ctx, &imMsg); err != nil {
				log.L.Error("update session error", zap.Error(err))
				//return consumer.SuspendCurrentQueueAMoment, err
			}
			go func(imMsg types.Message) {
				m.updateUserCache(ctx, &imMsg)
				m.dispatchMessage(&imMsg)
				m.pushToUser(ctx, int(immsg.SenderId), &imMsg)
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

func (m *MessageSubscribe) pushToUser(ctx context.Context, uid int, msg *types.Message) {
	key := RouterKey{}

	log.L.Info("push to self other device", zap.Any("msg", msg),
		zap.String("user_id", strconv.Itoa(uid)),
	)
	sids, err := m.Redis.HKeys(ctx, key.UserLocation(uid)).Result()
	if err != nil || len(sids) == 0 {
		return
	}
	log.L.Info("server_ids", zap.Any("sids", sids))

	payload, _ := json.Marshal(msg)

	for _, sid := range sids {
		userKey := key.UserClients(sid, msg.Channel, uid)
		cids, _ := m.Redis.SMembers(ctx, userKey).Result()
		log.L.Info("client_ids", zap.Any("cids", cids),
			zap.String("user_id", strconv.Itoa(uid)), zap.String("server_id", sid))

		if len(cids) == 0 {
			m.Redis.HDel(ctx, key.UserLocation(uid), sid)
			continue
		}

		cli, err := m.getRpcClient(sid)
		if err != nil {
			continue
		}

		for _, cidStr := range cids {
			cid, _ := strconv.ParseInt(cidStr, 10, 64)
			cli.PushToClient(ctx, &push.PushRequest{
				Cid:     cid,
				Uid:     int32(uid),
				Payload: string(payload),
				Event:   "chat",
			})
		}
	}
}

func (m *MessageSubscribe) dispatchMessage(msg *types.Message) {
	ctx := context.Background()
	key := RouterKey{}
	uid := int(msg.TargetID)
	msg.Status = types.MsgStatusSuccess

	trace := fmt.Sprintf(
		"[DISPATCH msg=%d from=%d to=%d]",
		msg.Id, msg.SenderID, msg.TargetID,
	)

	sids, err := m.Redis.HKeys(ctx, key.UserLocation(uid)).Result()
	if err != nil {
		log.L.Error("redis HKeys error:", zap.Error(err))
		return
	}
	if len(sids) == 0 {
		log.L.Error(" user offline", zap.Any("trace", trace))
		return
	}

	log.L.Info("user online status",
		zap.String("trace", trace),
		zap.Int("server_count", len(sids)),
		zap.Any("sids", sids), // %v 对应的 slice/map 使用 zap.Any
	)

	payload, _ := json.Marshal(msg)

	// 2️⃣ 遍历 server
	for _, sid := range sids {
		userKey := key.UserClients(sid, msg.Channel, uid)
		cids, err := m.Redis.SMembers(ctx, userKey).Result()
		if err != nil {
			log.L.Error("redis SMembers error:", zap.Error(err))
			continue
		}

		if len(cids) == 0 {
			log.L.Error(" user offline", zap.Any("trace", trace))
			m.Redis.HDel(ctx, key.UserLocation(uid), sid)
			continue
		}

		log.L.Info("user offline status", zap.String("trace", trace),
			zap.Int("server_count", len(cids)),
			zap.Any("cids", cids))

		cli, err := m.getRpcClient(sid)
		if err != nil {
			log.L.Error("get rpc client error:", zap.Error(err))
			continue
		}

		// 3️⃣ 遍历 client
		for _, cidStr := range cids {
			cid, _ := strconv.ParseInt(cidStr, 10, 64)
			log.L.Info("client offline", zap.Any("trace", trace),
				zap.Any("sid", sid), zap.Any("cid", cid))

			_, err := cli.PushToClient(ctx, &push.PushRequest{
				Cid:     cid,
				Uid:     int32(uid),
				Payload: string(payload),
				Event:   "chat",
			})
			if err != nil {
				log.L.Error("push to client error:", zap.Error(err))

				m.Redis.HDel(ctx, key.UserLocation(uid), sid)
				clientCache.Delete(sid)
			} else {
				log.L.Info("push success", zap.String("trace", trace))

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

	// Key 设计: unread:{user_id} -> Hash { "1_single_1001": 5 }
	// 表示：用户 receiverID 收到来自 1001 的单聊消息，未读数为 5
	m.UnreadStorage.PipeIncr(ctx, pipe, receiverID, types.SessionTypeSingle, senderID)

	// 无论是发送方还是接收方，会话列表展示的最后一条消息必须是最新的

	summary := &cache.LastCacheMessage{
		Content:   truncateContent(msg.Content, 50), // 截断内容防止浪费内存
		Timestamp: msg.Timestamp,
	}

	log.L.Info("update summary", zap.Any("summary", summary))

	err := m.MessageStorage.Set(ctx, types.SessionTypeSingle, receiverID, senderID, summary)
	if err != nil {
		log.L.Error("set message error", zap.Error(err))
	}
	log.L.Info("update user cache")
	err = m.MessageStorage.Set(ctx, types.SessionTypeSingle, senderID, receiverID, summary)
	if err != nil {
		log.L.Error("set message error", zap.Error(err))
	}
	// 3. 执行管道
	_, err = pipe.Exec(ctx)
	if err != nil {
		log.L.Error("pipe exec error", zap.Error(err))
	}
}

func truncateContent(content string, maxLen int) string {
	runes := []rune(content)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	return content
}
