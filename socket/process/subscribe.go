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
	"errors"
	"fmt"
	rmq_client "github.com/apache/rocketmq-clients/golang/v5"
	"github.com/cloudwego/kitex/client"
	mysqlerr "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"sync"
	"time"
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

		// 2) 入库：重复插入视为成功（幂等）
		if err := m.MessageService.SaveMessage(immsg); err != nil {
			if !isDupKeyErr(err) {
				log.L.Error("[MQ] 消息入库失败", zap.Error(err))
				return err
			}
			// duplicate => 说明之前已成功入库，继续补后续步骤
			log.L.Warn("[MQ] 消息重复入库(幂等)", zap.Int64("msg_id", imMsg.Id))
		}
		// 2) DB 会话更新：必须成功，否则 return err 让 MQ 重试
		if err := m.SessionService.UpdateSingleSession(ctx, &imMsg); err != nil {
			log.L.Error("[MQ] update session error", zap.Error(err), zap.Int64("msg_id", imMsg.Id))
			return err
		}
		// 3) Redis cache：推荐尽力而为（失败不阻塞 MQ）
		if err := m.updateUserCache(ctx, &imMsg); err != nil {
			log.L.Warn("[MQ] update cache failed (degraded)", zap.Error(err), zap.Int64("msg_id", imMsg.Id))
		}

		// 4) doneKey：放在关键路径之后
		if err := m.Redis.Set(ctx, doneKey, 1, 24*time.Hour).Err(); err != nil {
			return err
		}
		// 5) 推送：doneKey 后异步尽力而为
		// 注意：闭包一定要传参，防止 imMsg 在循环/协程中被污染
		go func(msg types.Message) {
			// 推送前可以使用新的 background ctx，防止父 context 取消导致推送中断
			bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

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
		// 1) 入库：重复插入视为成功
		if err := m.MessageService.SaveGroupMessage(gdb); err != nil {
			if !isDupKeyErr(err) {
				log.L.Error("[MQ] 群聊消息入库失败", zap.Error(err))
				return err
			}
			log.L.Warn("[MQ] 群聊消息重复入库(幂等)", zap.Int64("msg_id", imMsg.Id))
		}

		// 2) 查群成员：必须成功（否则无法更新会话未读）
		memberIDs, err := m.GroupMemberDAO.GetMemberIds(ctx, int(imMsg.TargetID))
		if err != nil {
			log.L.Error("[MQ] query group members failed",
				zap.Error(err),
				zap.Int64("group_id", imMsg.TargetID),
			)
			return err
		}
		// 群无成员：通常认为无需后续会话/推送，直接 done
		if len(memberIDs) == 0 {
			if err := m.Redis.Set(ctx, doneKey, 1, 24*time.Hour).Err(); err != nil {
				return err
			}
			return nil
		}

		// 3) DB 会话更新（含 unread_count）：必须成功
		if err := m.SessionService.UpsertGroupSessions(ctx, &imMsg, memberIDs); err != nil {
			log.L.Error("[MQ] upsert group sessions error", zap.Error(err), zap.Int64("group_id", imMsg.TargetID))
			return err
		}
		// 4) Redis cache：尽力而为
		memberUIDs := make([]string, 0, len(memberIDs))
		for _, id := range memberIDs {
			memberUIDs = append(memberUIDs, strconv.Itoa(id))
		}
		if err := m.updateCacheGroup(ctx, &imMsg, memberUIDs); err != nil {
			log.L.Warn("[MQ] update group cache failed (degraded)", zap.Error(err), zap.Int64("group_id", imMsg.TargetID))
		}
		// 5) doneKey 放在关键路径之后
		if err := m.Redis.Set(ctx, doneKey, 1, 24*time.Hour).Err(); err != nil {
			return err
		}
		// 6) 推送异步
		go func(copyMsg types.Message, members []string) {
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

func (m *MessageSubscribe) updateUserCache(ctx context.Context, msg *types.Message) error {
	// 1. 确定接收者 ID
	// 自己发给自己：不需要未读，也不必写两份摘要
	if msg.SenderID == msg.TargetID {
		return nil
	}
	receiverID := int(msg.TargetID)
	senderID := int(msg.SenderID)

	summary := &cache.LastCacheMessage{
		Content:   truncateContent(msg.Content, 50), // 截断内容防止浪费内存
		Timestamp: msg.Timestamp,
	}

	// 只维护“最后一条消息摘要”
	if err := m.MessageStorage.Set(ctx, types.SessionTypeSingle, receiverID, senderID, summary); err != nil {
		log.L.Error("set last message (receiver) error", zap.Error(err))
		return err
	}
	if err := m.MessageStorage.Set(ctx, types.SessionTypeSingle, senderID, receiverID, summary); err != nil {
		log.L.Error("set last message (sender) error", zap.Error(err))
		return err
	}

	return nil
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
	//测试
	log.L.Info("[DEBUG] HGetAll route raw",
		zap.Int("uid", uid),
		zap.Int("count", len(results)),
		zap.Any("raw", results),
	)

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

func (m *MessageSubscribe) updateCacheGroup(ctx context.Context, msg *types.Message, members []string) error {
	groupID := int(msg.TargetID)

	summary := &cache.LastCacheMessage{
		Content:   truncateContent(msg.Content, 50),
		Timestamp: msg.Timestamp,
	}
	// 群聊摘要：所有成员都更新一份 last_message（包括 sender）
	for _, uidStr := range members {
		uid, err := strconv.Atoi(uidStr)
		if err != nil || uid == 0 {
			continue
		}

		if err := m.MessageStorage.Set(ctx, types.GroupChatSessionTypeGroup, uid, groupID, summary); err != nil {
			log.L.Error("[Cache] set group last message error",
				zap.Error(err),
				zap.Int("group_id", groupID),
				zap.Int("uid", uid),
			)
			return err
		}
	}

	return nil
}

func isDupKeyErr(err error) bool {
	// MySQL duplicate key = 1062
	var me *mysqlerr.MySQLError
	if errors.As(err, &me) {
		return me.Number == 1062
	}
	// 兜底（有些场景 gorm 包装后不方便 As）
	return strings.Contains(err.Error(), "Duplicate entry")
}
