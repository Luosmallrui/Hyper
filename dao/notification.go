package dao

import (
	"Hyper/models"
	"context"

	"gorm.io/gorm"
)

// NotificationDAO 用户通知收件箱
type NotificationDAO struct {
	Repo[models.UserNotification]
}

func NewNotificationDAO(db *gorm.DB) *NotificationDAO {
	return &NotificationDAO{Repo: NewRepo[models.UserNotification](db)}
}

// byUser 构造按用户（可选类型）过滤的基础查询
func (d *NotificationDAO) byUser(ctx context.Context, userID int64, notifyType string) *gorm.DB {
	query := d.Model(ctx).Where("user_id = ?", userID)
	if notifyType != "" {
		query = query.Where("type = ?", notifyType)
	}
	return query
}

// Create 写入一条通知
func (d *NotificationDAO) Create(ctx context.Context, notification *models.UserNotification) error {
	return d.Db.WithContext(ctx).Create(notification).Error
}

// ListByUser 按 id DESC 分页拉取用户通知，notifyType 为空表示全部类型，同时返回总数
func (d *NotificationDAO) ListByUser(ctx context.Context, userID int64, notifyType string, page, size int) ([]models.UserNotification, int64, error) {
	var total int64
	if err := d.byUser(ctx, userID, notifyType).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.UserNotification
	err := d.byUser(ctx, userID, notifyType).
		Order("id DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&list).Error
	return list, total, err
}

// UnreadCountByType 按 type 分组统计未读数
func (d *NotificationDAO) UnreadCountByType(ctx context.Context, userID int64) (map[string]int64, error) {
	type countRow struct {
		Type  string
		Count int64
	}
	var rows []countRow
	err := d.Model(ctx).
		Select("type, COUNT(*) AS count").
		Where("user_id = ? AND is_read = 0", userID).
		Group("type").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int64, len(rows))
	for _, row := range rows {
		counts[row.Type] = row.Count
	}
	return counts, nil
}

// MarkReadByIDs 按 ID 批量标记已读（仅限本人的通知）
func (d *NotificationDAO) MarkReadByIDs(ctx context.Context, userID int64, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	return d.Model(ctx).
		Where("user_id = ? AND id IN ? AND is_read = 0", userID, ids).
		Update("is_read", 1).Error
}

// MarkAllRead 全部标记已读，notifyType 为空表示全部类型
func (d *NotificationDAO) MarkAllRead(ctx context.Context, userID int64, notifyType string) error {
	return d.byUser(ctx, userID, notifyType).
		Where("is_read = 0").
		Update("is_read", 1).Error
}
