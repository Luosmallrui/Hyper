package service

import (
	"Hyper/models"
	"Hyper/types"
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

var (
	ErrCustomerServiceNotConfigured   = errors.New("客服工作台未配置客服账号")
	ErrCustomerServiceSessionNotFound = errors.New("客服会话不存在或尚未由用户发起")
)

// customerServiceAccount returns the normal user account which represents the
// platform in client IM. Administrators never reply with their own identity.
func (s *AdminService) customerServiceAccount(ctx context.Context) (*models.Users, error) {
	if s.DB == nil {
		return nil, errors.New("客服工作台服务未初始化")
	}

	var setting models.PlatformSetting
	if err := s.DB.WithContext(ctx).Where("setting_key = ?", "customer_service_user_id").First(&setting).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerServiceNotConfigured
		}
		return nil, err
	}
	serviceUserID, err := strconv.ParseInt(strings.TrimSpace(setting.Value), 10, 64)
	if err != nil || serviceUserID <= 0 {
		return nil, ErrCustomerServiceNotConfigured
	}

	var account models.Users
	if err := s.DB.WithContext(ctx).Where("id = ? AND status = ?", serviceUserID, 1).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerServiceNotConfigured
		}
		return nil, err
	}
	return &account, nil
}

func (s *AdminService) customerServiceSessionExists(ctx context.Context, serviceUserID, customerID int64) error {
	var count int64
	if err := s.DB.WithContext(ctx).Model(&models.Session{}).
		Where("user_id = ? AND session_type = ? AND peer_id = ?", serviceUserID, types.SessionTypeSingle, customerID).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return ErrCustomerServiceSessionNotFound
	}
	return nil
}

// ListCustomerServiceSessions limits the workbench to sessions owned by the
// configured platform account, so arbitrary user-to-user IM is never exposed.
func (s *AdminService) ListCustomerServiceSessions(ctx context.Context, page, pageSize int, keyword string) (*types.AdminCustomerServiceSessionListResponse, error) {
	serviceAccount, err := s.customerServiceAccount(ctx)
	if err != nil {
		return nil, err
	}
	page, pageSize = normalizeAdminPage(page, pageSize)

	query := s.DB.WithContext(ctx).Table("im_session cs").
		Joins("JOIN users u ON u.id = cs.peer_id").
		Where("cs.user_id = ? AND cs.session_type = ? AND cs.peer_id <> ?", serviceAccount.Id, types.SessionTypeSingle, serviceAccount.Id)
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("u.nickname LIKE ? OR u.mobile LIKE ? OR CAST(u.id AS CHAR) LIKE ?", like, like, like)
	}

	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, err
	}
	list := make([]types.AdminCustomerServiceSessionItem, 0)
	if err := query.Select(`cs.peer_id AS user_id, u.nickname, u.avatar, u.mobile, u.status AS user_status,
		cs.last_msg_id, cs.last_msg_type, cs.last_msg_content AS last_msg, cs.last_msg_time, cs.unread_count AS unread`).
		Order("cs.is_top DESC, cs.last_msg_time DESC, cs.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&list).Error; err != nil {
		return nil, err
	}
	return &types.AdminCustomerServiceSessionListResponse{
		ServiceUserID: int64(serviceAccount.Id), List: list, Total: total, Page: page, PageSize: pageSize,
	}, nil
}

func (s *AdminService) ListCustomerServiceMessages(ctx context.Context, customerID uint64, cursor, since int64, limit int) (*types.AdminCustomerServiceMessageListResponse, error) {
	serviceAccount, err := s.customerServiceAccount(ctx)
	if err != nil {
		return nil, err
	}
	if customerID == 0 || customerID == uint64(serviceAccount.Id) {
		return nil, ErrCustomerServiceSessionNotFound
	}
	if err := s.customerServiceSessionExists(ctx, int64(serviceAccount.Id), int64(customerID)); err != nil {
		return nil, err
	}

	var customer models.Users
	if err := s.DB.WithContext(ctx).Where("id = ?", customerID).First(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerServiceSessionNotFound
		}
		return nil, err
	}
	if s.MessageService == nil {
		return nil, errors.New("客服消息服务未初始化")
	}
	messages, err := s.MessageService.ListMessages(ctx, uint64(serviceAccount.Id), customerID, types.SessionTypeSingle, cursor, since, limit)
	if err != nil {
		return nil, err
	}
	nextCursor := int64(0)
	if len(messages) > 0 {
		nextCursor = messages[0].Time
	}
	return &types.AdminCustomerServiceMessageListResponse{
		Service: types.AdminCustomerServiceContact{
			UserID: int64(serviceAccount.Id), Nickname: serviceAccount.Nickname, Avatar: serviceAccount.Avatar, Signature: serviceAccount.Motto,
		},
		Customer: types.AdminCustomerServiceContact{
			UserID: int64(customer.Id), Nickname: customer.Nickname, Avatar: customer.Avatar, Signature: customer.Motto,
		},
		List: messages, NextCursor: nextCursor,
	}, nil
}

// SendCustomerServiceMessage enters the same MQ, storage and Socket pipeline
// as a client IM reply, but the sender remains the platform account.
func (s *AdminService) SendCustomerServiceMessage(ctx context.Context, customerID uint64, req types.AdminCustomerServiceSendMessageRequest) (*types.Message, error) {
	serviceAccount, err := s.customerServiceAccount(ctx)
	if err != nil {
		return nil, err
	}
	if customerID == 0 || customerID == uint64(serviceAccount.Id) {
		return nil, ErrCustomerServiceSessionNotFound
	}
	if err := s.customerServiceSessionExists(ctx, int64(serviceAccount.Id), int64(customerID)); err != nil {
		return nil, err
	}
	if req.MsgType < types.MsgTypeText || req.MsgType > types.MsgTypeActivity {
		return nil, errors.New("不支持的客服消息类型")
	}
	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		return nil, errors.New("消息内容不能为空")
	}
	if len([]rune(req.Content)) > 5000 {
		return nil, errors.New("消息内容不能超过 5000 个字符")
	}
	if s.MessageService == nil {
		return nil, errors.New("客服消息服务未初始化")
	}

	message := &types.Message{
		SenderID:    int64(serviceAccount.Id),
		TargetID:    int64(customerID),
		SessionType: types.SessionTypeSingle,
		MsgType:     req.MsgType,
		Content:     req.Content,
		ParentMsgID: req.ParentMsgID,
		Ext:         req.Ext,
	}
	if err := s.MessageService.SendMessage(message); err != nil {
		return nil, err
	}
	return message, nil
}

// MarkCustomerServiceSessionRead clears only the configured service account's
// unread state. A stale read acknowledgement leaves newer messages unread.
func (s *AdminService) MarkCustomerServiceSessionRead(ctx context.Context, customerID uint64, readTime int64) error {
	serviceAccount, err := s.customerServiceAccount(ctx)
	if err != nil {
		return err
	}
	if customerID == 0 || customerID == uint64(serviceAccount.Id) {
		return ErrCustomerServiceSessionNotFound
	}
	if err := s.customerServiceSessionExists(ctx, int64(serviceAccount.Id), int64(customerID)); err != nil {
		return err
	}
	if readTime <= 0 {
		readTime = time.Now().UnixMilli()
	}
	return s.DB.WithContext(ctx).Model(&models.Session{}).
		Where("user_id = ? AND session_type = ? AND peer_id = ? AND last_msg_time <= ?", serviceAccount.Id, types.SessionTypeSingle, customerID, readTime).
		Update("unread_count", 0).Error
}
