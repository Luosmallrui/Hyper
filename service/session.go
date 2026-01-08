package service

import (
	"Hyper/dao/cache"
	"Hyper/models"
	"Hyper/types"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

var _ ISessionService = (*SessionService)(nil)

type SessionService struct {
	DB             *gorm.DB
	MessageStorage *cache.MessageStorage
	UnreadStorage  *cache.UnreadStorage
	UserService    IUserService
}

type ISessionService interface {
	UpdateSingleSession(ctx context.Context, msg *types.Message) error
	ListUserSessions(ctx context.Context, userId uint64, limit int) ([]*types.SessionDTO, error)
}

func (s *SessionService) UpdateSingleSession(
	ctx context.Context,
	msg *types.Message,
) error {

	// 自己发给自己，通常不计未读
	if msg.SenderID == msg.TargetID {
		return nil
	}

	summary := truncateContent(msg.Content, 50)

	// 发送方会话（unread = 0）
	if err := s.upsertConversation(
		ctx,
		uint64(msg.SenderID),
		types.SessionTypeSingle,
		uint64(msg.TargetID),
		msg,
		summary,
		0,
	); err != nil {
		return err
	}

	// 接收方会话（unread +1）
	if err := s.upsertConversation(
		ctx,
		uint64(msg.TargetID),
		types.SessionTypeSingle,
		uint64(msg.SenderID),
		msg,
		summary,
		1,
	); err != nil {
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

func (s *SessionService) upsertConversation(
	ctx context.Context,
	userId uint64,
	sessionType int,
	peerId uint64,
	msg *types.Message,
	summary string,
	unreadDelta int,
) error {

	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var conv models.Session

		err := tx.Where(
			"user_id = ? AND session_type = ? AND peer_id = ?",
			userId, sessionType, peerId,
		).Take(&conv).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 新会话
				conv = models.Session{
					UserId:         userId,
					SessionType:    sessionType,
					PeerId:         peerId,
					LastMsgId:      uint64(msg.Id),
					LastMsgType:    msg.MsgType,
					LastMsgContent: summary,
					LastMsgTime:    msg.Timestamp,
					UnreadCount:    uint32(max(unreadDelta, 0)),
				}
				return tx.Create(&conv).Error
			}
			return err
		}

		// 已存在会话，更新
		update := map[string]interface{}{
			"last_msg_id":      msg.Id,
			"last_msg_type":    msg.MsgType,
			"last_msg_content": summary,
			"last_msg_time":    msg.Timestamp,
			"updated_at":       time.Now(),
		}

		if unreadDelta > 0 {
			update["unread_count"] = gorm.Expr(
				"unread_count + ?", unreadDelta,
			)
		} else {
			update["unread_count"] = 0
		}

		return tx.Model(&models.Session{}).
			Where("id = ?", conv.Id).
			Updates(update).Error
	})
}

func (s *SessionService) ListUserSessions(ctx context.Context, userId uint64, limit int) ([]*types.SessionDTO, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	// 1. 批量查询会话表
	var convs []models.Session
	err := s.DB.WithContext(ctx).
		Where("user_id = ?", userId).
		Order("is_top DESC, last_msg_time DESC").
		Limit(limit).
		Find(&convs).Error
	if err != nil {
		return nil, err
	}

	if len(convs) == 0 {
		return []*types.SessionDTO{}, nil
	}

	// 2. 提取所有需要查询的 PeerId (去重)
	peerIds := make([]uint64, 0, len(convs))
	for _, c := range convs {
		peerIds = append(peerIds, c.PeerId)
	}

	// 3. 批量获取用户信息 (从 Redis + DB 回源)
	// 建议封装成我们之前讨论的 BatchGetUserInfo
	userInfoMap := s.UserService.BatchGetUserInfo(ctx, peerIds)

	// 4. 批量获取 Redis 中的未读数和最后一条消息
	// 注意：这里推荐在 MessageService 内部实现一个 Pipeline 批量获取方法
	unreadMap := s.UnreadStorage.BatchGet(ctx, userId, convs)
	lastMsgMap := s.MessageStorage.BatchGet(ctx, userId, convs)

	// 5. 在内存中组装结果
	result := make([]*types.SessionDTO, 0, len(convs))
	for _, c := range convs {
		info := userInfoMap[c.PeerId]
		dto := &types.SessionDTO{
			SessionType: c.SessionType,
			PeerId:      c.PeerId,
			IsTop:       c.IsTop,
			IsMute:      c.IsMute,
			PeerName:    info.Nickname,
			PeerAvatar:  info.Avatar,
		}

		// 优先使用 Redis 中的实时数据，如果不存在再用数据库里的兜底
		if last, ok := lastMsgMap[c.PeerId]; ok {
			dto.LastMsg = last.Content
			dto.LastMsgTime = last.Timestamp
		} else {
			dto.LastMsg = c.LastMsgContent
			dto.LastMsgTime = c.LastMsgTime
		}

		dto.Unread = unreadMap[c.PeerId]

		result = append(result, dto)
	}

	return result, nil
}
