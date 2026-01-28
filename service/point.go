package service

import (
	"Hyper/config"
	"Hyper/dao"
	"Hyper/models"
	"Hyper/types"
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type PointService struct {
	Config   *config.Config
	DB       *gorm.DB
	Redis    *redis.Client
	PointDAO *dao.Point
}

var _ IPointService = (*PointService)(nil)

type IPointService interface {
	ConsumePoints(ctx context.Context, userID uint64, amount int64, changeType int, sourceID string, remark string) (*types.PointsAccount, error)
	RewardPoints(ctx context.Context, userID uint64, amount int64, changeType int, sourceID string, remark string, isPending bool) (*types.PointsAccount, error)
	getAccountStats(ctx context.Context, userID uint64) (*types.PointsAccount, error)

	// 查询
	GetAccountDashboard(ctx context.Context, userID uint64) (*types.PointsAccount, error)
	ListPointRecords(ctx context.Context, userID uint64, action string, cursor int64, limit int) (*types.ListPointsRecord, error)
}

func (p *PointService) RewardPoints(ctx context.Context, userID uint64, amount int64, changeType int, sourceID string, remark string, isPending bool) (*types.PointsAccount, error) {
	if amount <= 0 {
		return nil, errors.New("奖励积分数额必须大于0")
	}

	var finalAccount models.UserPoint

	err := p.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		// 1. 幂等检查
		exists, err := p.PointDAO.CheckLogExists(ctx, userID, sourceID, changeType)
		if err != nil {
			return errors.New("检查积分变动记录失败" + err.Error())
		}
		if exists {
			return errors.New("该业务已处理，请勿重复操作")
		}

		targetStatus := int8(1) // 默认直接入账
		if isPending {
			targetStatus = int8(0)
		}
		// 确实入账更新
		if targetStatus == 1 {
			rowsAffected, err := p.PointDAO.UpdateBalance(ctx, userID, amount)
			if err != nil {
				return errors.New("更新用户积分余额失败: " + err.Error())
			}

			if rowsAffected == 0 {
				if err := p.PointDAO.CreateAccount(ctx, userID, amount); err != nil {
					// 开户失败必须捕获，防止出现有流水没账户的情况
					return errors.New("新用户积分账户创建失败: " + err.Error())
				}
			}
		}

		// 3. 获取账户快照
		acc, err := p.PointDAO.GetAccount(ctx, userID)
		if err != nil {
			acc = &models.UserPoint{UserID: userID, Balance: 0}
		}
		finalAccount = *acc
		// 4. 记录积分变动日志
		logRecord := &models.PointsLog{
			UserID:     userID,
			Amount:     amount,
			Balance:    acc.Balance,
			ChangeType: int8(changeType),
			SourceID:   sourceID,
			Remark:     remark,
			Status:     targetStatus,
		}
		return p.PointDAO.CreatePointLog(ctx, logRecord)
	})

	if err != nil {
		return nil, err
	}
	return &types.PointsAccount{
		Balance:     int(finalAccount.Balance),
		TotalEarned: int(finalAccount.TotalEarned),
		TotalUsed:   int(finalAccount.TotalUsed),
	}, nil
}

func (p *PointService) ConsumePoints(ctx context.Context, userID uint64, amount int64, changeType int, sourceID string, remark string) (*types.PointsAccount, error) {
	if amount <= 0 {
		return nil, errors.New("消费积分数额必须大于0")
	}
	var finalAccount models.UserPoint
	err := p.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 幂等检查
		exists, err := p.PointDAO.CheckLogExists(ctx, userID, sourceID, changeType)
		if err != nil {
			return errors.New("检查积分变动记录失败" + err.Error())
		}
		if exists {
			return errors.New("该业务已处理，请勿重复操作")
		}
		account, err := p.PointDAO.GetAccount(ctx, userID)
		if err != nil {
			return errors.New("用户积分账户不存在")
		}
		if account.Balance < amount {
			return errors.New("积分余额不足")
		}
		rows, err := p.PointDAO.UpdateBalance(ctx, userID, -amount)
		if err != nil || rows == 0 {
			return errors.New("更新用户积分余额失败: " + err.Error())
		}
		finalAccount = *account
		finalAccount.Balance -= amount

		targetStatus := int8(1)
		logRecord := &models.PointsLog{
			UserID:     userID,
			Amount:     -amount,
			Balance:    finalAccount.Balance,
			ChangeType: int8(changeType),
			SourceID:   sourceID,
			Remark:     remark,
			Status:     targetStatus,
		}
		return p.PointDAO.CreatePointLog(ctx, logRecord)
	})
	if err != nil {
		return nil, err
	}
	return &types.PointsAccount{
		Balance:     int(finalAccount.Balance),
		TotalEarned: int(finalAccount.TotalEarned),
		TotalUsed:   int(finalAccount.TotalUsed),
	}, nil
}
func (p *PointService) GetAccountDashboard(ctx context.Context, userID uint64) (*types.PointsAccount, error) {
	account, err := p.PointDAO.GetAccount(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果没记录，说明是新用户，直接返回初始状态
			return &types.PointsAccount{
				Balance: 0, TotalEarned: 0, TotalUsed: 0,
				PendingCount: 0, PendingAmount: 0,
			}, nil
		}
		return nil, errors.New("查询积分账户失败")
	}
	pCount, pAmount, err := p.PointDAO.GetPendingStats(ctx, userID)
	if err != nil {
		pCount, pAmount = 0, 0
	}
	return &types.PointsAccount{
		Balance:       int(account.Balance),
		TotalEarned:   int(account.TotalEarned),
		TotalUsed:     int(account.TotalUsed),
		PendingCount:  int(pCount),
		PendingAmount: int(pAmount),
	}, nil
}
func (p *PointService) getAccountStats(ctx context.Context, userID uint64) (*types.PointsAccount, error) {
	account, err := p.PointDAO.GetAccount(ctx, userID)
	if err != nil {
		return nil, errors.New("用户积分账户不存在")
	}
	pCount, pAmount, err := p.PointDAO.GetPendingStats(ctx, userID)
	if err != nil {
		return nil, errors.New("统计待入账数据失败")
	}

	return &types.PointsAccount{
		Balance:       int(account.Balance),
		TotalEarned:   int(account.TotalEarned),
		TotalUsed:     int(account.TotalUsed),
		PendingCount:  int(pCount),
		PendingAmount: int(pAmount),
	}, nil
}

func (p *PointService) ListPointRecords(ctx context.Context, userID uint64, action string, cursor int64, limit int) (*types.ListPointsRecord, error) {
	logs, err := p.PointDAO.ListRecords(ctx, userID, action, cursor, limit+1)
	if err != nil {
		return nil, errors.New("查询积分流水失败")
	}

	resp := &types.ListPointsRecord{
		Records: make([]types.PointRecord, 0),
		HasMore: false,
	}

	if len(logs) > limit {
		resp.HasMore = true
		logs = logs[:limit]
		resp.NextCursor = int64(logs[len(logs)-1].ID)
	}

	for _, l := range logs {
		orderType := "INCOME"
		if l.Amount < 0 {
			orderType = "EXPENSE"
		}
		resp.Records = append(resp.Records, types.PointRecord{
			ID:          int(l.ID),
			Amount:      int(l.Amount),
			Description: l.Remark,
			OrderType:   orderType,
			Status:      int(l.Status),
			CreatedAt:   l.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return resp, nil
}
