package service

import (
	"Hyper/models"
	"Hyper/pkg/log"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var _ IPayService = (*PayService)(nil)

type PayService struct {
	DB *gorm.DB
}

type IPayService interface {
	ProcessOrderPaySuccess(ctx context.Context, notify *payments.Transaction) error
}

func (s *PayService) ProcessOrderPaySuccess(ctx context.Context, notify *payments.Transaction) error {
	// 获取微信的订单号和支付状态
	orderSn := *notify.OutTradeNo
	transactionId := *notify.TransactionId
	tradeState := *notify.TradeState
	tradeType := *notify.TradeType
	var openid string
	if notify.Payer != nil && notify.Payer.Openid != nil {
		openid = *notify.Payer.Openid
	}
	// 只有当支付状态为 SUCCESS 时才处理逻辑
	if tradeState != "SUCCESS" {
		log.L.Info("支付未成功，跳过处理", zap.String("order_sn", orderSn), zap.String("state", tradeState))
		return nil
	}

	// 开启事务
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 幂等检查：查询流水表并锁定行（SELECT FOR UPDATE）
		var record models.PayRecord
		err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("order_sn = ?", orderSn).First(&record).Error
		if err != nil {
			return err
		}

		// 如果流水状态已经是“已支付”，直接返回，不再执行后续逻辑
		if record.PayStatus == 2 { // 2 代表已支付
			log.L.Info("订单已处理过，跳过", zap.String("order_sn", orderSn))
			return nil
		}

		// 2. 更新支付流水表
		rawJson, _ := json.Marshal(notify) // 将整个回调原始数据存入JSON字段
		updateRecord := map[string]interface{}{
			"transaction_id":  transactionId,
			"pay_status":      2,
			"raw_trade_state": tradeState,
			"pay_method":      tradeType,
			"payer_id":        openid,
			"notify_raw":      rawJson,
			"finished_at":     time.Now(),
		}
		if err := tx.Model(&record).Updates(updateRecord).Error; err != nil {
			return err
		}

		// 3. 更新主订单状态
		result := tx.Model(&models.Order{}).
			Where("order_sn = ? AND status = ?", orderSn, 10). // 只有待支付状态才能变更为已支付
			Updates(map[string]interface{}{
				"status":  20, // 已支付
				"paid_at": time.Now(),
			})

		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("更新订单状态失败或订单状态已改变")
		}

		// 4. 【关键】执行后续业务逻辑
		// 例如：给用户账户加钱、增加积分、如果是商品则通知仓库发货
		// if err := s.AccountService.AddBalance(record.UserID, record.AmountTotal); err != nil {
		//     return err
		// }

		return nil
	})
}
