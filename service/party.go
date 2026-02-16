package service

import (
	"Hyper/models"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type MerchantService struct {
	Redis *redis.Client
	DB    *gorm.DB
}

// 确保 MerchantService 实现了 IMerchantService 接口
var _ IMerchantService = (*MerchantService)(nil)

type IMerchantService interface {
	SubcribParty(ctx context.Context, userId int, partyId int) error
	UnsubcribParty(ctx context.Context, userId int, partyId int) error
}

// GenerateKey 生成 Redis Key，例如 user:party:subscribe:100
func (s *MerchantService) GenerateKey(userId int) string {
	return fmt.Sprintf("user:party:subscribe:%d", userId)
}

// SubcribParty 订阅派对
func (s *MerchantService) SubcribParty(ctx context.Context, userId int, partyId int) error {
	lockkey := fmt.Sprintf("user:party:subscribe:%d:%d", userId, partyId)
	ok, err := s.Redis.SetNX(ctx, lockkey, 1, 5*time.Second).Result()
	if err != nil {
		return kerrors.NewBizStatusError(50000, "订阅失败")
	}
	if !ok {
		return kerrors.NewBizStatusError(50001, "请勿重复订阅")
	}
	defer s.Redis.Del(ctx, lockkey)

	return s.DB.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Table("party_likes").
			Where("user_id = ? AND party_id = ?", userId, partyId).
			Count(&count).Error; err != nil {
			return errors.New("查询订阅状态失败，请稍后再试")
		}
		if count > 0 {
			return errors.New("请勿重复订阅")
		}

		like := models.PartyLike{
			PartyID:   int64(partyId),
			UserID:    userId,
			CreatedAt: time.Now(),
		}
		if err := tx.Create(&like).Error; err != nil {
			return errors.New("订阅失败，请稍后再试")
		}
		// 更新派对的订阅数量
		//result := tx.Model(&models.MemberItem{}).
		//	Where("id = ?", partyId).
		//	UpdateColumn("current_count", gorm.Expr("current_count + ?", 1))
		//if result.Error != nil {
		//	return result.Error
		//}
		//if result.RowsAffected == 0 {
		//	return errors.New("派对不存在")
		//}
		userSubkey := fmt.Sprintf("user:party:subscribe:list:%d", userId)
		s.Redis.SAdd(ctx, userSubkey, like.PartyID)
		return nil
	})
}

// UnsubcribParty 取消订阅派对
func (s *MerchantService) UnsubcribParty(ctx context.Context, userId int, partyId int) error {
	lockkey := fmt.Sprintf("user:party:unsubscribe:%d:%d", userId, partyId)
	ok, err := s.Redis.SetNX(ctx, lockkey, 1, 5*time.Second).Result()
	if err != nil {
		return errors.New("系统繁忙")
	}
	if !ok {
		return errors.New("请勿重复操作")
	}
	defer s.Redis.Del(ctx, lockkey)
	return s.DB.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("user_id = ? AND party_id = ?", userId, partyId).Delete(&models.PartyLike{})
		if result.Error != nil {
			return errors.New("系统繁忙，请稍后再试")
		}
		if result.RowsAffected == 0 {
			return errors.New("未订阅该派对")
		}

		userSubkey := fmt.Sprintf("user:party:subscribe:list:%d", userId)
		s.Redis.SRem(ctx, userSubkey, partyId)
		userSubKeyV2 := s.GenerateKey(userId)
		s.Redis.SRem(ctx, userSubKeyV2, partyId)

		//// 4. 原子减少计数：SET current_count = current_count - 1
		//// 增加条件 current_count > 0 防止出现负数
		//updateRes := tx.Model(&models.Merchant{}).
		//	Where("id = ? AND current_count > 0", partyId).
		//	UpdateColumn("current_count", gorm.Expr("current_count - ?", 1))
		//
		//if updateRes.Error != nil {
		//	return updateRes.Error
		//}

		return nil
	})
}

//func(s *MerchantService) ShareParty(ctx context.Context,partyId int64) (types.SharePartyResq,error) {
//
//}
