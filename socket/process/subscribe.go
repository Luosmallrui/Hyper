package process

import (
	"Hyper/dao"
	"Hyper/dao/cache"
	"Hyper/models"
	"Hyper/pkg/log"
	"Hyper/rpc/kitex_gen/im/push"
	"Hyper/rpc/kitex_gen/im/push/pushservice"
	"Hyper/service"
	"Hyper/types"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	rmq_client "github.com/apache/rocketmq-clients/golang/v5"
	"github.com/cloudwego/kitex/client"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type MessageSubscribe struct {
	Redis          *redis.Client
	DB             *gorm.DB
	MessageService service.IMessageService
	MessageStorage *cache.MessageStorage
	SessionService service.ISessionService
	UnreadStorage  *cache.UnreadStorage
	GroupMemberDAO *dao.GroupMember
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
	return nil
}
func (m *MessageSubscribe) Setup(ctx context.Context) error {
	return nil
}

func (m *MessageSubscribe) handleMessage(ctx context.Context, msgs *rmq_client.MessageView) error {
	var imMsg types.Message
	if err := json.Unmarshal(msgs.GetBody(), &imMsg); err != nil {
		log.L.Error("unmarshal msg error", zap.Error(err))
		return err //返回err防止Ack 掉
	}
	// 幂等去重：done + lock 两段式
	doneKey := fmt.Sprintf("im:dedup:done:%d", imMsg.Id)
	lockKey := fmt.Sprintf("im:dedup:lock:%d", imMsg.Id)

	// 1) 已经成功处理过：直接返回 nil（上层会 Ack），避免重复推送/重复入库
	done, err := m.Redis.Exists(ctx, doneKey).Result()
	if err != nil {
		return err
	}
	if done == 1 {
		return nil
	}

	// 2) 抢占处理锁：抢不到说明其他进程/协程正在处理 -> 返回 error（不要 Ack，让 MQ 重试）
	ok, err := m.Redis.SetNX(ctx, lockKey, 1, 2*time.Minute).Result()
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("msg %d is being processed", imMsg.Id)
	}

	// 3) 后续失败要释放锁，保证能重试
	defer func() {
		_ = m.Redis.Del(context.Background(), lockKey).Err()
	}()

	extBytes, _ := json.Marshal(imMsg.Ext)
	switch imMsg.SessionType {
	case types.SessionTypeSingle:
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
			//CreatedAt:   time.Now().Unix(),
			CreatedAt: func() int64 {
				if imMsg.Timestamp > 0 {
					return imMsg.Timestamp // 毫秒：发送时用的毫秒
				}
				return time.Now().UnixMilli() // 兜底
			}(),

			Ext:       string(extBytes),
			UpdatedAt: time.Now(),
		}

		// 1. 同步持久化
		if err := m.MessageService.SaveMessage(immsg); err != nil {
			log.L.Error("[MQ] 消息入库失败", zap.Error(err))
			return err // 让上层不 Ack，MQ 自动重试
			// 数据库挂了建议让 MQ 重试
		}
		// 入库成功：标记已完成（24h），并释放锁
		_ = m.Redis.Set(ctx, doneKey, 1, 24*time.Hour).Err()
		_ = m.Redis.Del(ctx, lockKey).Err()

		if err := m.SessionService.UpdateSingleSession(ctx, &imMsg); err != nil {
			log.L.Error("update session error", zap.Error(err))
		}

		// 2. 异步分发推送
		// 注意：闭包一定要传参，防止 imMsg 在循环/协程中被污染
		go func(msg types.Message) {
			// 推送前可以使用新的 background ctx，防止父 context 取消导致推送中断
			bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			m.updateUserCache(bgCtx, &msg)

			// 给接收者推（dispatchMessage 的逻辑）
			m.doBatchPush(bgCtx, int(msg.SenderID), &msg, int(msg.TargetID))

			// 给发送者其他端推（多端同步，pushToUser 的逻辑）
			if msg.SenderID != msg.TargetID {
				m.doBatchPush(bgCtx, int(msg.SenderID), &msg, int(msg.SenderID))
			}
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
			//CreatedAt:   time.Now().Unix(),
			CreatedAt: func() int64 {
				if imMsg.Timestamp > 0 {
					return imMsg.Timestamp // 毫秒：发送时用的毫秒
				}
				return time.Now().UnixMilli() // 兜底
			}(),

			Ext:       string(extBytes),
			UpdatedAt: time.Now(),
		}
		if err := m.MessageService.SaveGroupMessage(gdb); err != nil {
			log.L.Error("[MQ] 群聊消息入库失败: %v", zap.Error(err))
			return err
		}
		// 入库成功：标记已完成（24h），并释放锁
		_ = m.Redis.Set(ctx, doneKey, 1, 24*time.Hour).Err()
		_ = m.Redis.Del(ctx, lockKey).Err()

		// 3) 查群成员：使用 group_member DAO（真实成员表）
		memberIDs := m.GroupMemberDAO.GetMemberIds(ctx, int(imMsg.TargetID))
		if len(memberIDs) == 0 {
			// 没成员就不推了（也可以只推给自己，按产品定义）
			return nil
		}

		// 转成 []string（因为后面的函数签名是 []string）
		memberUIDs := make([]string, 0, len(memberIDs))
		for _, id := range memberIDs {
			memberUIDs = append(memberUIDs, strconv.Itoa(id))
		}

		go func(copyMsg types.Message, members []string) {
			m.updateCacheGroup(ctx, &copyMsg, members)
			m.dispatchToGroup(ctx, &copyMsg, members)
		}(imMsg, memberUIDs)

	default:
		log.L.Error(fmt.Sprintf("[MQ] 未知 SessionType=%d, msg_id=%d", imMsg.SessionType, imMsg.Id))
	}
	return nil
}
func (m *MessageSubscribe) doBatchPush(ctx context.Context, uid int, msg *types.Message, targetUID int) {
	trace := fmt.Sprintf("[PUSH msg=%d from=%d to=%d]", msg.Id, msg.SenderID, targetUID)

	// 获取路由
	routeMap, err := m.GetUserRoute(ctx, targetUID)
	if err != nil {
		log.L.Error("获取用户路由失败", zap.Error(err), zap.String("trace", trace))
		return
	}

	if len(routeMap) == 0 {
		log.L.Info("用户离线", zap.String("trace", trace))
		return
	}

	payload, _ := json.Marshal(msg)

	for sid, cids := range routeMap {
		cli, err := m.getRpcClient(sid)
		if err != nil {
			log.L.Error("获取 RPC 客户端失败", zap.String("sid", sid), zap.Error(err))
			continue
		}

		// 转换 CID
		cidsInt64 := make([]int64, 0, len(cids))
		for _, s := range cids {
			if id, err := strconv.ParseInt(s, 10, 64); err == nil {
				cidsInt64 = append(cidsInt64, id)
			}
		}

		if len(cidsInt64) == 0 {
			continue
		}

		// 调用 BatchPush
		_, err = cli.BatchPushToClient(ctx, &push.BatchPushRequest{
			Cids:    cidsInt64,
			Uid:     int32(targetUID),
			Payload: string(payload),
			Event:   "chat",
		})

		if err != nil {
			log.L.Error("推送失败", zap.Error(err), zap.String("sid", sid), zap.Int("target", targetUID))
			// 可选：如果 RPC 报错说明 sid 可能挂了，可以选择清理 clientCache
			// clientCache.Delete(sid)
		} else {
			log.L.Info("推送成功", zap.String("trace", trace), zap.String("sid", sid), zap.Int("count", len(cidsInt64)))
		}
	}
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
	//ctx := context.Background()
	uid = int(msg.SenderID)
	msg.Status = types.MsgStatusSuccess
	trace := fmt.Sprintf("[DISPATCH msg=%d from=%d to=%d]", msg.Id, msg.SenderID, msg.TargetID)

	routeMap, err := m.GetUserRoute(ctx, uid)
	if err != nil {
		log.L.Error("获取用户路由失败", zap.Error(err), zap.String("trace", trace))
		return
	}

	if len(routeMap) == 0 {
		log.L.Info("用户离线", zap.String("trace", trace))
		return
	}

	log.L.Info("用户在线", zap.String("trace", trace), zap.Int("server_count", len(routeMap)))

	payload, _ := json.Marshal(msg)

	for sid, cids := range routeMap {
		cli, err := m.getRpcClient(sid)
		if err != nil {
			log.L.Error("获取 RPC 客户端失败", zap.String("sid", sid), zap.Error(err))
			continue
		}
		log.L.Debug("开始推送", zap.String("sid", sid), zap.Any("cids", cids))
		cidsInt64 := make([]int64, 0, len(cids))
		// 2. 遍历并转换
		for _, s := range cids {
			// 使用 strconv.ParseInt 转换为 10 进制的 int64
			id, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				// 如果转换失败（例如字符串里有非法字符），记录日志并跳过
				log.L.Warn("非法 CID 格式", zap.String("cid", s), zap.Error(err))
				continue
			}
			cidsInt64 = append(cidsInt64, id)
		}
		_, err = cli.BatchPushToClient(ctx, &push.BatchPushRequest{
			Cids:    cidsInt64,
			Uid:     int32(uid),
			Payload: string(payload),
			Event:   "chat",
		})
		if err != nil {
			//log.L.Error("推送失败", zap.Error(err), zap.String("sid", sid), zap.Int64("cid", cid))
			//m.Redis.HDel(ctx, fmt.Sprintf("im:user:location:%d", uid), cidStr)
		} else {
			log.L.Info("推送成功", zap.String("trace", trace), zap.Any("cids", cids))
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

func (m *MessageSubscribe) GetUserRoute(ctx context.Context, uid int) (map[string][]string, error) {
	// 1. 一键获取该用户所有的 clientId 和对应的 serverId
	// Key: im:user:location:100 -> { "c1": "sid_A", "c2": "sid_A", "c3": "sid_B" }
	results, err := m.Redis.HGetAll(ctx, fmt.Sprintf("im:user:location:%d", uid)).Result()
	if err != nil {
		return nil, err
	}

	// 2. 在内存中按 sid 进行分组
	routeMap := make(map[string][]string)
	for cid, sid := range results {
		routeMap[sid] = append(routeMap[sid], cid)
	}

	// 返回结果示例: map["sid_A"]: ["c1", "c2"], map["sid_B"]: ["c3"]
	return routeMap, nil
}

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

			m.doBatchPush(ctx, int(msg.SenderID), msg, receiver)
		}(uid)
	}

	wg.Wait()
}

//已弃用，推送改为用doBatchPush，群聊推送也调用doBatchPush
//// 私聊推送
//func (m *MessageSubscribe) dispatchToUser(ctx context.Context, uid int, msg *types.Message) {
//	key := RouterKey{}
//	msg.Status = types.MsgStatusSuccess
//
//	// 1) 取出用户在哪些 sid 上在线
//	sids, err := m.Redis.HKeys(ctx, key.UserLocation(uid)).Result()
//	if err != nil || len(sids) == 0 {
//		log.L.Error(fmt.Sprintf("[Router] uid=%d offline or no route (no sids), err=%v", uid, err))
//		return
//	}
//
//	payload, _ := json.Marshal(msg)
//
//	for _, sid := range sids {
//
//		userKey := key.UserClients(sid, msg.Channel, uid)
//
//		// 2) 这里做“短暂重试”，避免刚连接时 set 还没写完
//		var cids []string
//		for attempt := 0; attempt < 3; attempt++ {
//			cids, err = m.Redis.SMembers(ctx, userKey).Result()
//			if err == nil && len(cids) > 0 {
//				break
//			}
//			// 10ms, 30ms, 50ms（总共 < 100ms）
//			time.Sleep(time.Duration(10+attempt*20) * time.Millisecond)
//		}
//
//		if err != nil {
//			log.L.Error(fmt.Sprintf("[Router] uid=%d sid=%s read cids err=%v", uid, sid, err))
//			continue
//		}
//
//		// 关键：不要因为 cids 空就删除 location
//		if len(cids) == 0 {
//			log.L.Error(fmt.Sprintf("[Router] uid=%d sid=%s has route but empty cids (skip, no delete)", uid, sid))
//			continue
//		}
//
//		// 3) RPC 推送
//		cli, err := m.getRpcClient(sid)
//		if err != nil {
//			log.L.Error(fmt.Sprintf("[RPC] get client fail sid=%s err=%v", sid, err))
//			continue
//		}
//
//		for _, cidStr := range cids {
//			cid, _ := strconv.ParseInt(cidStr, 10, 64)
//			_, err := cli.PushToClient(ctx, &push.PushRequest{
//				Cid:     cid,
//				Uid:     int32(uid),
//				Payload: string(payload),
//				Event:   "chat",
//			})
//			if err != nil {
//				// 只有 RPC 明确不可达，才清理 sid（这个清理是合理的）
//				log.L.Error(fmt.Sprintf("[RPC] push fail uid=%d sid=%s cid=%d err=%v, cleanup route", uid, sid, cid, err))
//				m.Redis.HDel(ctx, key.UserLocation(uid), sid)
//				clientCache.Delete(sid)
//				break
//			}
//		}
//	}
//}

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
		log.L.Error(fmt.Sprintf("[Cache] group 更新失败 group=%d err=%v", groupID, err))
	}
}
