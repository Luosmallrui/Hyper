package service

import (
	"Hyper/dao"
	"Hyper/models"
	"Hyper/pkg/log"
	"Hyper/types"
	"context"
	"encoding/json"

	rmq_client "github.com/apache/rocketmq-clients/golang/v5"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// INotificationService 用户通知收件箱：系统/互动/支付消息统一落库 + 在线 WS 提醒。
// Notify 失败只记日志不返回错误，不能阻塞主业务流程。
type INotificationService interface {
	// Notify 写入一条通知并尝试向在线用户推送 WS 事件 notice.new（离线靠收件箱补拉）
	Notify(ctx context.Context, userID int64, notifyType, title, content, payload string)
	// List 分页拉取通知，notifyType 为空表示全部类型
	List(ctx context.Context, userID int64, notifyType string, page, size int) (*types.PageResponse[types.NotificationItem], error)
	// UnreadCount 未读数（按类型分组）
	UnreadCount(ctx context.Context, userID int64) (*types.UnreadCountResponse, error)
	// MarkRead 按 ID 批量标记已读（仅限本人的通知）
	MarkRead(ctx context.Context, userID int64, ids []int64) error
	// MarkAllRead 全部标记已读，notifyType 为空表示全部类型
	MarkAllRead(ctx context.Context, userID int64, notifyType string) error
}

const notificationTimeFormat = "2006-01-02 15:04:05"

// NotificationService 用户通知收件箱实现
type NotificationService struct {
	DB         *gorm.DB
	MqProducer rmq_client.Producer
}

var _ INotificationService = (*NotificationService)(nil)

func (s *NotificationService) notificationDAO() *dao.NotificationDAO {
	return dao.NewNotificationDAO(s.DB)
}

// Notify 写入一条通知并尝试向在线用户推送 WS 事件 notice.new（离线靠收件箱补拉）。
// 任何失败只记日志，不返回错误、不允许 panic，不能阻塞主业务流程。
func (s *NotificationService) Notify(ctx context.Context, userID int64, notifyType, title, content, payload string) {
	notification := &models.UserNotification{
		UserID:  userID,
		Type:    notifyType,
		Title:   title,
		Content: content,
		Payload: payload,
	}
	if err := s.notificationDAO().Create(ctx, notification); err != nil {
		log.L.Error("写入用户通知失败", zap.Error(err), zap.Int64("user_id", userID))
		return
	}

	if s.MqProducer == nil {
		return
	}
	data, err := json.Marshal(types.NotificationPayload{
		UserID:    userID,
		Type:      notifyType,
		Title:     title,
		Content:   content,
		Payload:   payload,
		CreatedAt: notification.CreatedAt.Format(notificationTimeFormat),
	})
	if err != nil {
		log.L.Error("序列化用户通知失败", zap.Error(err), zap.Int64("user_id", userID))
		return
	}
	body, err := json.Marshal(types.SystemMessage{Type: "user_notification", Data: json.RawMessage(data)})
	if err != nil {
		log.L.Error("序列化用户通知事件失败", zap.Error(err), zap.Int64("user_id", userID))
		return
	}
	if _, err := s.MqProducer.Send(ctx, &rmq_client.Message{Topic: types.SystemMessageTopic, Body: body}); err != nil {
		// 实时推送失败不影响收件箱，用户下次拉取仍能看到
		log.L.Error("发送用户通知MQ失败", zap.Error(err), zap.Int64("user_id", userID))
	}
}

// List 分页拉取通知，notifyType 为空表示全部类型
func (s *NotificationService) List(ctx context.Context, userID int64, notifyType string, page, size int) (*types.PageResponse[types.NotificationItem], error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 50 {
		size = 50
	}
	list, total, err := s.notificationDAO().ListByUser(ctx, userID, notifyType, page, size)
	if err != nil {
		return nil, err
	}
	items := make([]types.NotificationItem, 0, len(list))
	for _, n := range list {
		items = append(items, types.NotificationItem{
			ID:        n.ID,
			Type:      n.Type,
			Title:     n.Title,
			Content:   n.Content,
			Payload:   n.Payload,
			IsRead:    n.IsRead == 1,
			CreatedAt: n.CreatedAt.Format(notificationTimeFormat),
		})
	}
	return &types.PageResponse[types.NotificationItem]{List: items, Total: total}, nil
}

// UnreadCount 未读数（按类型分组）
func (s *NotificationService) UnreadCount(ctx context.Context, userID int64) (*types.UnreadCountResponse, error) {
	counts, err := s.notificationDAO().UnreadCountByType(ctx, userID)
	if err != nil {
		return nil, err
	}
	resp := &types.UnreadCountResponse{
		System:      counts[types.NotifyTypeSystem],
		Interaction: counts[types.NotifyTypeInteraction],
		Payment:     counts[types.NotifyTypePayment],
	}
	resp.Total = resp.System + resp.Interaction + resp.Payment
	return resp, nil
}

// MarkRead 按 ID 批量标记已读（仅限本人的通知）
func (s *NotificationService) MarkRead(ctx context.Context, userID int64, ids []int64) error {
	return s.notificationDAO().MarkReadByIDs(ctx, userID, ids)
}

// MarkAllRead 全部标记已读，notifyType 为空表示全部类型
func (s *NotificationService) MarkAllRead(ctx context.Context, userID int64, notifyType string) error {
	return s.notificationDAO().MarkAllRead(ctx, userID, notifyType)
}
