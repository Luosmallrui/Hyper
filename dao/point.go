package dao

import (
	"Hyper/models"
	"context"

	"gorm.io/gorm"
)

type Point struct {
	Repo[models.UserPoint]
}

func NewPoint(db *gorm.DB) *Point {
	return &Point{
		Repo: NewRepo[models.UserPoint](db),
	}
}

func (p *Point) CheckLogExists(ctx context.Context, userID uint64, sourceID string, changeType int) (bool, error) {
	var count int64
	err := p.Db.WithContext(ctx).Model(&models.PointsLog{}).
		Where("user_id = ? AND source_id = ? AND change_type = ?", userID, sourceID, changeType).
		Count(&count).Error
	return count > 0, err
}

// GetAccountForUpdate 获取账户信息（带锁或在事务中）
func (p *Point) GetAccount(ctx context.Context, userID uint64) (*models.UserPoint, error) {
	var account models.UserPoint
	err := p.Db.WithContext(ctx).Where("user_id = ?", userID).First(&account).Error
	return &account, err
}

// CreateAccount 初始化账户（针对新用户）
func (p *Point) CreateAccount(ctx context.Context, userID uint64, initialPoints int64) error {
	newAccount := &models.UserPoint{
		UserID:      userID,
		Balance:     initialPoints,
		TotalEarned: uint64(initialPoints),
		TotalUsed:   0,
	}
	return p.Db.WithContext(ctx).Create(newAccount).Error
}
func (p *Point) CreatePointLog(ctx context.Context, log *models.PointsLog) error {
	return p.Db.WithContext(ctx).Create(log).Error
}

func (p *Point) UpdateBalance(ctx context.Context, userID uint64, amount int64) (int64, error) {
	result := p.Db.WithContext(ctx).Model(&models.UserPoint{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			// gorm.Expr 保证了并发下的原子加减，避免数据覆盖
			"balance":      gorm.Expr("balance + ?", amount),
			"total_earned": gorm.Expr("total_earned + ?", amount),
		})

	// 返回受影响的行数，用于 Service 判断是否需要“自动开户”
	return result.RowsAffected, result.Error
}

// GetPendingStats 统计待入账数据
func (p *Point) GetPendingStats(ctx context.Context, userID uint64) (count int64, amount int64, err error) {
	var res struct {
		Count  int64
		Amount int64
	}
	err = p.Db.WithContext(ctx).Table("point_logs"). // 建议使用模型的 TableName()
								Select("COUNT(*) AS count, IFNULL(SUM(amount), 0) AS amount").
								Where("user_id = ? AND status = ?", userID, 0).
								Scan(&res).Error
	return res.Count, res.Amount, err
}

// ListRecords 分页筛选查询
func (p *Point) ListRecords(ctx context.Context, userID uint64, action string, cursor int64, limit int) ([]models.PointsLog, error) {
	var logs []models.PointsLog
	query := p.Db.WithContext(ctx).Where("user_id = ?", userID)

	switch action {
	case "income":
		query = query.Where("amount > ?", 0)
	case "expense":
		query = query.Where("amount < ?", 0)
	}

	if cursor > 0 {
		query = query.Where("id < ?", cursor)
	}

	err := query.Order("id DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

// CheckHasMore 检查是否还有更多数据
func (p *Point) GetNextCursorRecord(ctx context.Context, userID uint64, action string, lastID int64) (bool, error) {
	var count int64
	query := p.Db.WithContext(ctx).Model(&models.PointsLog{}).Where("user_id = ? AND id < ?", userID, lastID)

	switch action {
	case "income":
		query = query.Where("amount > ?", 0)
	case "expense":
		query = query.Where("amount < ?", 0)
	}

	err := query.Count(&count).Error
	return count > 0, err
}
