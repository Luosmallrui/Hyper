package service

import (
	"Hyper/config"
	"Hyper/models"
	"Hyper/pkg/log"
	utilBase "Hyper/pkg/utils"
	"Hyper/types"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var _ IPayService = (*PayService)(nil)

type PayService struct {
	DB     *gorm.DB
	Config *config.Config
}

type IPayService interface {
	ProcessOrderPaySuccess(ctx context.Context, notify *payments.Transaction) error
	PreWeChatPay(ctx context.Context, weChatClient *core.Client, req types.PrepayRequest) (types.PrepayWithRequestPaymentResponse, error)
}

func (p *PayService) PreWeChatPay(
	ctx context.Context,
	weChatClient *core.Client,
	req types.PrepayRequest,
) (types.PrepayWithRequestPaymentResponse, error) {

	var respData types.PrepayWithRequestPaymentResponse

	// 0. 参数校验
	if req.Amount <= 0 {
		return respData, fmt.Errorf("invalid pay amount")
	}
	if req.Openid == "" {
		return respData, fmt.Errorf("openid is required")
	}

	wechatCfg := p.Config.WechatPayConfig
	orderSn := utilBase.GenerateOrderSn(req.UserId)

	// 1. 事务内：创建订单 + 支付流水（只做 DB）
	err := p.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// A. 创建订单
		order := &models.Order{
			UserID:      req.UserId,
			OrderSn:     orderSn,
			TotalAmount: uint64(req.Amount),
			Description: req.Description,
			Status:      10, // 待支付
		}
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		// B. 创建支付流水
		payRecord := &models.PayRecord{
			OrderSn:     orderSn,
			PayPlatform: 1, // 微信
			AmountTotal: uint64(req.Amount),
			PayStatus:   0, // 待支付
		}
		if err := tx.Create(payRecord).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return respData, err
	}

	// 2. 事务外：调用微信 JSAPI 预下单
	svc := jsapi.JsapiApiService{Client: weChatClient}
	prepayReq := jsapi.PrepayRequest{
		Appid:       core.String(wechatCfg.AppID),
		Mchid:       core.String(wechatCfg.MchID),
		Description: core.String(req.Description),
		OutTradeNo:  core.String(orderSn),
		NotifyUrl:   core.String(wechatCfg.NotifyURL),
		Amount: &jsapi.Amount{
			Total: core.Int64(req.Amount),
		},
		Payer: &jsapi.Payer{
			Openid: core.String(req.Openid),
		},
	}

	wxResp, _, err := svc.PrepayWithRequestPayment(ctx, prepayReq)
	if err != nil {
		return respData, fmt.Errorf("wechat prepay failed: %w", err)
	}

	// 3. 更新支付流水（写入 prepay_id）
	if err := p.DB.WithContext(ctx).
		Model(&models.PayRecord{}).
		Where("order_sn = ?", orderSn).
		Update("out_request_no", *wxResp.PrepayId).Error; err != nil {
		return respData, err
	}

	// 4. 组装返回给前端的支付参数
	respData = types.PrepayWithRequestPaymentResponse{
		Appid:     *wxResp.Appid,
		TimeStamp: *wxResp.TimeStamp,
		NonceStr:  *wxResp.NonceStr,
		Package:   *wxResp.Package,
		SignType:  *wxResp.SignType,
		PaySign:   *wxResp.PaySign,
		PrepayId:  *wxResp.PrepayId,
	}

	return respData, nil
}

func (p *PayService) ProcessOrderPaySuccess(ctx context.Context, notify *payments.Transaction) error {
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
	return p.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
