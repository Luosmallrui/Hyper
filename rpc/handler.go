package handler

import (
	"Hyper/pkg/log"
	"Hyper/pkg/socket"
	"Hyper/rpc/kitex_gen/im/push"
	"Hyper/types"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PushServiceImpl implements the last service interface defined in the IDL.
type PushServiceImpl struct {
	Db    *gorm.DB
	Redis *redis.Client
}

// PushToClient implements the PushServiceImpl interface.
func (s *PushServiceImpl) PushToClient(
	ctx context.Context,
	req *push.PushRequest,
) (*push.PushResponse, error) {

	// trace 前缀：一条消息贯穿全链路
	trace := fmt.Sprintf(
		"[PUSH msg=%s uid=%d cid=%d event=%s]",
		req.Payload, //q 如果太长，下面会优化
		req.Uid,
		req.Cid,
		req.Event,
	)
	log.L.Info("enter PushToClient", zap.Any("trace", trace))

	ch := socket.Session.Chat
	if ch == nil {
		log.L.Error("ch is nil")
		return &push.PushResponse{
			Success: false,
			Msg:     "chat channel not initialized",
		}, nil
	}

	client, ok := ch.Client(req.Cid)
	if !ok {
		log.L.Error("client  not found")
		return &push.PushResponse{
			Success: false,
			Msg:     "client offline",
		}, nil
	}

	// 解析 payload
	var m struct {
		Id          int64                  `json:"msg_id,string"`
		ClientMsgID string                 `json:"client_msg_id"`
		SenderID    int64                  `json:"sender_id,string"`
		TargetID    int64                  `json:"target_id,string"`
		SessionID   string                 `json:"session_id"`
		SessionType int                    `json:"session_type"`
		MsgType     int                    `json:"msg_type"`
		Content     string                 `json:"content"`
		ParentMsgID int64                  `json:"parent_msg_id,string"`
		Timestamp   int64                  `json:"timestamp"`
		Status      int                    `json:"status"`
		Ext         map[string]interface{} `json:"ext"`
	}

	if err := json.Unmarshal([]byte(req.Payload), &m); err != nil {
		log.L.Error("unmarshal payload", zap.Error(err), zap.Any("payload", req.Payload))
		return &push.PushResponse{
			Success: false,
			Msg:     "invalid payload",
		}, err
	}

	extBytes, err := json.Marshal(m.Ext)
	if err != nil {
		log.L.Error("marshal ext", zap.Error(err), zap.Any("ext", m.Ext))
		extBytes = []byte("{}")
	}

	dto := &types.MessageDTO{
		MsgID:       strconv.FormatInt(m.Id, 10),
		ClientMsgID: m.ClientMsgID,
		SenderID:    strconv.FormatInt(m.SenderID, 10),
		TargetID:    strconv.FormatInt(m.TargetID, 10),
		SessionID:   m.SessionID,
		SessionType: m.SessionType,
		MsgType:     m.MsgType,
		Content:     m.Content,
		Timestamp:   m.Timestamp,
		Status:      m.Status,
		Ext:         extBytes,
	}

	if m.ParentMsgID != 0 {
		dto.ParentMsgID = strconv.FormatInt(m.ParentMsgID, 10)
	}

	res := &socket.ClientResponse{
		IsAck:   ok,
		Event:   req.Event,
		Content: dto,
		IsSelf:  false,
	}

	if m.SenderID == int64(client.Uid()) {
		res.IsSelf = true
	}
	if err := client.Write(res); err != nil {

		log.L.Error("write response", zap.Error(err))

		return &push.PushResponse{
			Success: false,
			Msg:     err.Error(),
		}, nil
	}
	return &push.PushResponse{Success: true}, nil
}

func (s *PushServiceImpl) BatchPushToClient(ctx context.Context, req *push.BatchPushRequest) (r *push.PushResponse, err error) {
	// 1. 基础检查
	ch := socket.Session.Chat
	if ch == nil {
		log.L.Error("ch is nil")
		return &push.PushResponse{Success: false, Msg: "chat channel not initialized"}, nil
	}

	// 2. 公共逻辑提取：消息解析与 DTO 转换（只做一次）
	// 这里直接复用你之前的解析逻辑
	var m struct {
		Id          int64                  `json:"msg_id,string"`
		ClientMsgID string                 `json:"client_msg_id"`
		SenderID    int64                  `json:"sender_id,string"`
		TargetID    int64                  `json:"target_id,string"`
		SessionID   string                 `json:"session_id"`
		SessionType int                    `json:"session_type"`
		MsgType     int                    `json:"msg_type"`
		Content     string                 `json:"content"`
		ParentMsgID int64                  `json:"parent_msg_id,string"`
		Timestamp   int64                  `json:"timestamp"`
		Status      int                    `json:"status"`
		Ext         map[string]interface{} `json:"ext"`
	}

	if err := json.Unmarshal([]byte(req.Payload), &m); err != nil {
		log.L.Error("batch unmarshal payload failed", zap.Error(err))
		return &push.PushResponse{Success: false, Msg: "invalid payload"}, nil
	}

	extBytes, _ := json.Marshal(m.Ext)
	dto := &types.MessageDTO{
		MsgID:       strconv.FormatInt(m.Id, 10),
		ClientMsgID: m.ClientMsgID,
		SenderID:    strconv.FormatInt(m.SenderID, 10),
		TargetID:    strconv.FormatInt(m.TargetID, 10),
		SessionID:   m.SessionID,
		SessionType: m.SessionType,
		MsgType:     m.MsgType,
		Content:     m.Content,
		Timestamp:   m.Timestamp,
		Status:      m.Status,
		Ext:         extBytes,
	}
	if m.ParentMsgID != 0 {
		dto.ParentMsgID = strconv.FormatInt(m.ParentMsgID, 10)
	}

	// 3. 循环推送给不同的 CID
	successCount := 0
	failCount := 0
	var user types.UserProfile
	_ = s.Db.WithContext(ctx).Table("users").Select("avatar", "nickname").Where("id = ?", m.SenderID).Take(&user).Error
	for _, cid := range req.Cids {
		client, ok := ch.Client(cid)
		if !ok {
			failCount++
			continue
		}

		// 构造响应对象
		res := &socket.ClientResponse{
			IsAck:    true,
			Event:    req.Event,
			Content:  dto,
			IsSelf:   m.SenderID == int64(client.Uid()),
			NickName: user.Nickname,
			Avatar:   user.Avatar,
		}
		log.L.Info("push res", zap.Any("res", res))

		// 执行写入
		if err := client.Write(res); err != nil {
			log.L.Error("batch write error", zap.Int64("cid", cid), zap.Error(err))
			failCount++
		} else {
			successCount++
		}
	}

	log.L.Info("batch push finished",
		zap.Int("total", len(req.Cids)),
		zap.Int("success", successCount),
		zap.Int("fail", failCount))

	return &push.PushResponse{
		Success: successCount > 0, // 只要有一个成功就算成功，或者根据业务自定义
		Msg:     fmt.Sprintf("success:%d, fail:%d", successCount, failCount),
	}, nil
}

func (s *PushServiceImpl) BatchGetUserInfo(ctx context.Context, uids []uint64) map[uint64]types.UserProfile {
	result := make(map[uint64]types.UserProfile)
	if len(uids) == 0 {
		return result
	}

	// 1. Redis MGet 批量获取
	keys := make([]string, len(uids))
	for i, id := range uids {
		keys[i] = fmt.Sprintf("user:info:%d", id)
	}

	cacheRes, _ := s.Redis.MGet(ctx, keys...).Result()

	missingIds := make([]uint64, 0)
	for i, val := range cacheRes {
		if val != nil {
			var info types.UserProfile
			_ = json.Unmarshal([]byte(val.(string)), &info)
			result[uids[i]] = info
		} else {
			missingIds = append(missingIds, uids[i])
		}
	}

	// 2. 如果有缓存缺失，查数据库
	if len(missingIds) > 0 {
		var dbUsers []struct {
			Id       uint64
			Avatar   string
			Nickname string
		}
		s.Db.Table("users").Where("id IN ?", missingIds).Find(&dbUsers)

		pipe := s.Redis.Pipeline()
		for _, user := range dbUsers {
			info := types.UserProfile{Avatar: user.Avatar, Nickname: user.Nickname, UserID: user.Id}
			result[user.Id] = info

			// 写入缓存供下次使用
			data, _ := json.Marshal(info)
			pipe.Set(ctx, fmt.Sprintf("user:info:%d", user.Id), data, 15*time.Second)
		}
		_, _ = pipe.Exec(ctx)
	}

	return result
}
