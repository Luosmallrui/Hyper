package process

import (
	"Hyper/dao/cache"
	"Hyper/models"
	"Hyper/rpc/kitex_gen/im/push"
	"Hyper/rpc/kitex_gen/im/push/pushservice"
	"time"

	"Hyper/pkg/log"
	"Hyper/service"
	"Hyper/types"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

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
		return nil
	}

	// 业务逻辑处理
	if imMsg.SessionType == types.SessionTypeSingle {
		extBytes, _ := json.Marshal(imMsg.Ext)
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

		// 1. 同步持久化
		if err := m.MessageService.SaveMessage(immsg); err != nil {
			log.L.Error("[MQ] 消息入库失败", zap.Error(err))
			// 数据库挂了建议让 MQ 重试
			//return rmq_client.ActionNack, err/
		}

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
