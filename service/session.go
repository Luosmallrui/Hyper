package service

import (
	"Hyper/dao/cache"
	"Hyper/models"
	"Hyper/types"
	"context"
	"errors"
	"strconv"
	"time"

	"gorm.io/gorm"
)

var _ ISessionService = (*SessionService)(nil)

type SessionService struct {
	DB             *gorm.DB
	MessageStorage *cache.MessageStorage
	UnreadStorage  *cache.UnreadStorage
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

	var convs []models.Session
	err := s.DB.WithContext(ctx).
		Where("user_id = ?", userId).
		Order("is_top DESC, last_msg_time DESC").
		Limit(limit).
		Find(&convs).Error
	if err != nil {
		return nil, err
	}

	result := make([]*types.SessionDTO, 0, len(convs))

	for _, c := range convs {
		dto := &types.SessionDTO{
			SessionType: c.SessionType,
			PeerId:      c.PeerId,
			LastMsg:     c.LastMsgContent,
			LastMsgTime: c.LastMsgTime,
			Unread:      c.UnreadCount,
			IsTop:       c.IsTop,
			IsMute:      c.IsMute,
		}

		unread := s.UnreadStorage.Get(ctx, int(userId), c.SessionType, int(c.PeerId))
		dto.Unread = uint32(unread)
		last, err := s.MessageStorage.Get(ctx, c.SessionType, int(userId), int(c.PeerId))
		if err == nil {
			dto.LastMsg = last.Content
			if ts, err := strconv.ParseInt(last.Datetime, 10, 64); err == nil {
				dto.LastMsgTime = ts
			}
		}
		result = append(result, dto)
	}

	return result, nil
}
