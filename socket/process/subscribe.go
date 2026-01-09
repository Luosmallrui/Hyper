package process

import (
	"Hyper/dao"
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
	GroupDAO       *dao.GroupDAO
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
			continue
		}

		b, _ := json.Marshal(imMsg)
		log.Printf("[MQ] 解析成功: %s", string(b))
		extBytes, _ := json.Marshal(imMsg.Ext)

		switch imMsg.SessionType {
		case types.SessionTypeSingle:
			imdb := &models.ImSingleMessage{
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

			if err := m.MessageService.SaveSingleMessage(imdb); err != nil {
				log.Printf("[MQ] 单聊消息入库失败: %v", err)
			}

			if err := m.SessionService.UpdateSingleSession(ctx, &imMsg); err != nil {
				log.Printf("[MQ] update single session error: %v", err)
			}
			go func(copyMsg types.Message) {
				m.updateCacheSingle(ctx, &copyMsg)
				m.dispatchToUser(ctx, int(copyMsg.TargetID), &copyMsg)

			}(imMsg)
		case types.GroupChatSessionTypeGroup:
			// 2) 群聊落库：只存 1 条到 im_message（models.Message）
			gdb := &models.ImGroupMessage{
				Id:          imMsg.Id,
				SessionHash: imMsg.SessionHash,
				SessionId:   imMsg.SessionID,
				SenderId:    imMsg.SenderID,
				TargetId:    imMsg.TargetID, // groupId
				MsgType:     imMsg.MsgType,
				Content:     imMsg.Content,
				ParentMsgId: imMsg.ParentMsgID,
				Status:      types.MsgStatusSuccess,
				CreatedAt:   time.Now().Unix(),
				Ext:         string(extBytes),
				UpdatedAt:   time.Now(),
			}
			if err := m.MessageService.SaveGroupMessage(gdb); err != nil {
				log.Printf("[MQ] 群聊消息入库失败: %v", err)
			}
			// 3) 查群成员（批量推送的“成员列表”来源）
			memberUIDs, err := m.GroupDAO.GetGroupMembers(fmt.Sprintf("%d", imMsg.TargetID))
			if err != nil {
				log.Printf("[MQ] 获取群成员失败 group=%d err=%v", imMsg.TargetID, err)
				continue
			}
			go func(copyMsg types.Message, members []string) {
				m.updateCacheGroup(ctx, &copyMsg, members)
				m.dispatchToGroup(ctx, &copyMsg, members)
			}(imMsg, memberUIDs)
		default:
			log.Printf("[MQ] 未知 SessionType=%d, msg_id=%d", imMsg.SessionType, imMsg.Id)
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

// 私聊推送
//
//	func (m *MessageSubscribe) dispatchToUser(ctx context.Context, uid int, msg *types.Message) {
//		ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
//		defer cancel()
//
//		key := RouterKey{}
//		msg.Status = types.MsgStatusSuccess
//
//		sids, err := m.Redis.HKeys(ctx, key.UserLocation(uid)).Result()
//		if err != nil || len(sids) == 0 {
//			log.Printf("[MQ] 用户 %d 不在线", uid)
//			return
//		}
//
//		payload, _ := json.Marshal(msg)
//
//		for _, sid := range sids {
//			userKey := key.UserClients(sid, msg.Channel, uid)
//			cids, _ := m.Redis.SMembers(ctx, userKey).Result()
//
//			if len(cids) == 0 {
//				// 如果某台机器上已经没客户端了，顺手清理一下 Hash
//				m.Redis.HDel(ctx, key.UserLocation(uid), sid)
//				continue
//			}
//
//			// 获取 RPC Client 并推送
//			cli, err := m.getRpcClient(sid)
//			if err != nil {
//				continue
//			}
//
//			for _, cidStr := range cids {
//				cid, _ := strconv.ParseInt(cidStr, 10, 64)
//				_, err := cli.PushToClient(ctx, &push.PushRequest{
//					Cid:     cid,
//					Uid:     int32(uid),
//					Payload: string(payload),
//					Event:   "chat",
//				})
//				if err != nil {
//					log.Printf("[RPC] 目标服务不可达，正在清理脏路由: %s", sid)
//
//					// 立即从该用户的 Location Hash 中移除该 sid
//					m.Redis.HDel(ctx, key.UserLocation(uid), sid)
//
//					// 同时清理 RPC 客户端缓存
//					clientCache.Delete(sid)
//				}
//			}
//		}
//	}
//
// 私聊推送（也被群聊复用）
func (m *MessageSubscribe) dispatchToUser(ctx context.Context, uid int, msg *types.Message) {
	key := RouterKey{}
	msg.Status = types.MsgStatusSuccess

	// 1) 取出用户在哪些 sid 上在线
	sids, err := m.Redis.HKeys(ctx, key.UserLocation(uid)).Result()
	if err != nil || len(sids) == 0 {
		log.Printf("[Router] uid=%d offline or no route (no sids), err=%v", uid, err)
		return
	}

	payload, _ := json.Marshal(msg)

	for _, sid := range sids {

		userKey := key.UserClients(sid, msg.Channel, uid)

		// 2) 这里做“短暂重试”，避免刚连接时 set 还没写完
		var cids []string
		for attempt := 0; attempt < 3; attempt++ {
			cids, err = m.Redis.SMembers(ctx, userKey).Result()
			if err == nil && len(cids) > 0 {
				break
			}
			// 10ms, 30ms, 50ms（总共 < 100ms）
			time.Sleep(time.Duration(10+attempt*20) * time.Millisecond)
		}

		if err != nil {
			log.Printf("[Router] uid=%d sid=%s read cids err=%v", uid, sid, err)
			continue
		}

		// 关键：不要因为 cids 空就删除 location
		if len(cids) == 0 {
			log.Printf("[Router] uid=%d sid=%s has route but empty cids (skip, no delete)", uid, sid)
			continue
		}

		// 3) RPC 推送
		cli, err := m.getRpcClient(sid)
		if err != nil {
			log.Printf("[RPC] get client fail sid=%s err=%v", sid, err)
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
				// 只有 RPC 明确不可达，才清理 sid（这个清理是合理的）
				log.Printf("[RPC] push fail uid=%d sid=%s cid=%d err=%v, cleanup route", uid, sid, cid, err)

				m.Redis.HDel(ctx, key.UserLocation(uid), sid)
				clientCache.Delete(sid)
				break
			}
		}
	}
}

// 群聊推送
func (m *MessageSubscribe) dispatchToGroup(ctx context.Context, msg *types.Message, members []string) {
	const maxFanout = 32
	sem := make(chan struct{}, maxFanout)
	var wg sync.WaitGroup

	for _, uidStr := range members {
		uid, err := strconv.Atoi(uidStr)
		if err != nil {
			continue
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(receiver int) {
			defer wg.Done()
			defer func() { <-sem }()

			m.dispatchToUser(ctx, receiver, msg)
		}(uid)
	}

	wg.Wait()
}

func (m *MessageSubscribe) updateCacheSingle(ctx context.Context, msg *types.Message) {
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

	summary := &cache.LastCacheMessage{
		Content:   truncateContent(msg.Content, 50), // 截断内容防止浪费内存
		Timestamp: msg.Timestamp,
	}

	m.MessageStorage.Set(ctx, types.SessionTypeSingle, receiverID, senderID, summary)
	m.MessageStorage.Set(ctx, types.SessionTypeSingle, senderID, receiverID, summary)

	// 3. 执行管道
	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Printf("[Cache]single 更新用户 %d 的缓存失败: %v", receiverID, err)
	}
}
func (m *MessageSubscribe) updateCacheGroup(ctx context.Context, msg *types.Message, members []string) {
	groupID := int(msg.TargetID)
	senderID := int(msg.SenderID)
	pipe := m.Redis.Pipeline()
	summary := &cache.LastCacheMessage{
		Content:   truncateContent(msg.Content, 50),
		Timestamp: msg.Timestamp,
	}
	for _, uidStr := range members {
		uid, err := strconv.Atoi(uidStr)
		if err != nil {
			continue
		}
		// sender 不加未读，其他人加未读
		if uid != senderID {
			m.UnreadStorage.PipeIncr(ctx, pipe, uid, types.GroupChatSessionTypeGroup, groupID)
		}
		// 所有人（含 sender）更新会话列表摘要：peerId 用 groupID
		m.MessageStorage.Set(ctx, types.GroupChatSessionTypeGroup, uid, groupID, summary)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		log.Printf("[Cache] group 更新失败 group=%d err=%v", groupID, err)
	}
}
func truncateContent(content string, maxLen int) string {
	runes := []rune(content)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	return content
}
