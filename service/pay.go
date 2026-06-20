package service

import (
	"Hyper/config"
	"Hyper/models"
	"Hyper/pkg/log"
	utilBase "Hyper/pkg/utils"
	"Hyper/types"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ IPayService = (*PayService)(nil)

type PayService struct {
	DB     *gorm.DB
	Config *config.Config
}

func (p *PayService) OrderDetail(ctx context.Context, OrderId string) (*types.OrderDetail, error) {
	var resq types.OrderDetail
	var order models.Order
	var orderItems models.OrderItem

	err := p.DB.WithContext(ctx).Where("order_sn = ?", OrderId).First(&order).Error
	if err != nil {
		return nil, fmt.Errorf("获取订单失败: %w", err)
	}
	err = p.DB.WithContext(ctx).Where("order_sn = ?", order.OrderSn).First(&orderItems).Error
	if err != nil {
		return nil, fmt.Errorf("获取订单明细失败: %w", err)
	}
	resq = types.OrderDetail{
		Name:       orderItems.ProductName,
		Avatar:     orderItems.CoverImage,
		Price:      int64(orderItems.ProductPrice),
		Quantity:   int(orderItems.Quantity),
		Status:     order.Status,
		OutTradeNo: order.OrderSn,
	}
	if order.PaidAt != nil {
		resq.PayedAt = *order.PaidAt
	}
	if orderItems.ConsumeType == "ticket" {
		resq.Attach = map[string]string{
			"event_time": "2024-12-31 19:00:00",
		}
	}
	if resq.Attach == nil {
		resq.Attach = make(map[string]string)
	}
	return &resq, nil
}

type IPayService interface {
	ProcessOrderPaySuccess(ctx context.Context, notify *payments.Transaction) error
	ApplyWechatRefund(ctx context.Context, weChatClient *core.Client, refundNo string) error
	SyncWechatRefund(ctx context.Context, weChatClient *core.Client, refundNo string) (*models.Refund, error)
	ProcessRefundNotify(ctx context.Context, refund *refunddomestic.Refund) error
	PreWeChatPay(ctx context.Context, weChatClient *core.Client, req types.PrepayRequest) (types.PrepayWithRequestPaymentResponse, error)
	GetOrderReceipt(ctx context.Context, orderSn string, userId int) (*types.OrderReceiptResponse, error)
	OrderDetail(ctx context.Context, OrderId string) (*types.OrderDetail, error)
}

func (p *PayService) PreWeChatPay(
	ctx context.Context,
	weChatClient *core.Client,
	req types.PrepayRequest,
) (types.PrepayWithRequestPaymentResponse, error) {

	var respData types.PrepayWithRequestPaymentResponse

	// 0. 参数校验
	if req.Openid == "" {
		return respData, fmt.Errorf("openid is required")
	}
	if req.OrderNo != "" {
		return p.prepayTicketOrder(ctx, weChatClient, req)
	}
	if req.ProductId == 0 {
		return respData, fmt.Errorf("product_id is required")
	}
	if req.Quantity == 0 {
		return respData, fmt.Errorf("quantity is required")
	}
	if req.Amount <= 0 {
		return respData, fmt.Errorf("invalid pay amount")
	}

	wechatCfg := p.Config.WechatPayConfig
	orderSn := utilBase.GenerateOrderSn(req.UserId)

	// 1. 事务内：创建订单 + 订单明细 + 支付流水（只做 DB）
	err := p.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var product models.Product
		if err := tx.Where("id = ? AND status = 1", req.ProductId).First(&product).Error; err != nil {
			return fmt.Errorf("商品不存在或已下架")
		}

		// 服务端金额校验：确保前端传的金额 = 商品单价 × 数量
		//expectedAmount := int64(product.Price) * int64(req.Quantity)
		//if req.Amount != expectedAmount {
		//	return fmt.Errorf("金额不匹配: 期望 %d, 实际 %d", expectedAmount, req.Amount)
		//}

		// 幂等检查：同一个用户对同一个商品，如果已有未支付订单，直接复用
		var existingOrder models.Order
		if err := tx.Where("user_id = ? AND status = 10 AND order_sn LIKE ?",
			req.UserId, fmt.Sprintf("%%_%d", req.UserId)).
			First(&existingOrder).Error; err == nil && existingOrder.ID > 0 {
			// 检查是否超过5分钟，超过则不算幂等
			if time.Since(existingOrder.CreatedAt) < 5*time.Minute {
				// 查找对应的 order item 确认是同一商品
				var existingItem models.OrderItem
				if err := tx.Where("order_sn = ? AND product_id = ?", existingOrder.OrderSn, req.ProductId).
					First(&existingItem).Error; err == nil {
					orderSn = existingOrder.OrderSn
					return nil // 复用已有订单
				}
			}
		}

		// 库存预扣减（下单时锁定库存）
		res := tx.Model(&models.Product{}).
			Where("id = ? AND stock >= ?", product.ID, req.Quantity).
			UpdateColumn("stock", gorm.Expr("stock - ?", req.Quantity))
		if res.Error != nil {
			return fmt.Errorf("扣减库存失败: %w", res.Error)
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("库存不足")
		}

		// A. 创建订单
		order := &models.Order{
			UserID:      req.UserId,
			OrderSn:     orderSn,
			TotalAmount: uint64(req.Amount),
			Description: product.Description,
			Status:      10, // 待支付
		}
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		orderItem := &models.OrderItem{
			OrderSn:        order.OrderSn,
			ProductID:      product.ID,
			ProductName:    product.ProductName,
			ProductPrice:   product.Price,
			Quantity:       req.Quantity,
			SubtotalAmount: product.Price * req.Quantity,
			CoverImage:     product.CoverImage,
		}
		if err := tx.Create(orderItem).Error; err != nil {
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
		Appid:      *wxResp.Appid,
		TimeStamp:  *wxResp.TimeStamp,
		NonceStr:   *wxResp.NonceStr,
		Package:    *wxResp.Package,
		SignType:   *wxResp.SignType,
		PaySign:    *wxResp.PaySign,
		PrepayId:   *wxResp.PrepayId,
		OutTradeNo: orderSn,
	}

	return respData, nil
}

func (p *PayService) prepayTicketOrder(
	ctx context.Context,
	weChatClient *core.Client,
	req types.PrepayRequest,
) (types.PrepayWithRequestPaymentResponse, error) {
	var respData types.PrepayWithRequestPaymentResponse
	var order models.TicketOrder
	var activity models.Activity
	var ticketSpec models.TicketSpec

	err := p.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("order_no = ? AND user_id = ?", req.OrderNo, req.UserId).
			First(&order).Error; err != nil {
			return fmt.Errorf("票务订单不存在: %w", err)
		}
		if order.Status != models.TicketOrderStatusPending {
			return fmt.Errorf("当前订单状态不可支付")
		}
		//if !order.ExpireTime.IsZero() && time.Now().After(order.ExpireTime) {
		//	return fmt.Errorf("订单已过期")
		//}
		if err := tx.First(&activity, order.ActivityID).Error; err != nil {
			return fmt.Errorf("活动不存在: %w", err)
		}
		if err := tx.First(&ticketSpec, order.TicketSpecID).Error; err != nil {
			return fmt.Errorf("票券不存在: %w", err)
		}
		req.Amount = order.ActualPrice
		req.Description = activity.Name + " - " + ticketSpec.Name
		var record models.PayRecord
		err := tx.Where("order_sn = ?", order.OrderNo).First(&record).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&models.PayRecord{
				OrderSn:     order.OrderNo,
				PayPlatform: 1,
				AmountTotal: uint64(order.ActualPrice),
				PayStatus:   0,
			}).Error
		}
		if err != nil {
			return err
		}
		if record.PayStatus == 2 {
			return fmt.Errorf("订单已支付")
		}
		if record.AmountTotal != uint64(order.ActualPrice) {
			return tx.Model(&record).Updates(map[string]any{
				"amount_total": order.ActualPrice,
				"pay_status":   0,
			}).Error
		}
		return nil
	})
	if err != nil {
		return respData, err
	}

	wechatCfg := p.Config.WechatPayConfig
	svc := jsapi.JsapiApiService{Client: weChatClient}
	wxResp, _, err := svc.PrepayWithRequestPayment(ctx, jsapi.PrepayRequest{
		Appid:       core.String(wechatCfg.AppID),
		Mchid:       core.String(wechatCfg.MchID),
		Description: core.String(req.Description),
		OutTradeNo:  core.String(order.OrderNo),
		NotifyUrl:   core.String(wechatCfg.NotifyURL),
		Amount: &jsapi.Amount{
			Total: core.Int64(order.ActualPrice),
		},
		Payer: &jsapi.Payer{
			Openid: core.String(req.Openid),
		},
	})
	if err != nil {
		return respData, fmt.Errorf("wechat prepay failed: %w", err)
	}

	if err := p.DB.WithContext(ctx).
		Model(&models.PayRecord{}).
		Where("order_sn = ?", order.OrderNo).
		Update("out_request_no", *wxResp.PrepayId).Error; err != nil {
		return respData, err
	}

	return types.PrepayWithRequestPaymentResponse{
		Appid:      *wxResp.Appid,
		TimeStamp:  *wxResp.TimeStamp,
		NonceStr:   *wxResp.NonceStr,
		Package:    *wxResp.Package,
		SignType:   *wxResp.SignType,
		PaySign:    *wxResp.PaySign,
		PrepayId:   *wxResp.PrepayId,
		OutTradeNo: order.OrderNo,
	}, nil
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
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("order_sn = ?", orderSn).First(&record).Error
		if err != nil {
			return err
		}

		// 如果流水状态已经是"已支付"，直接返回，不再执行后续逻辑
		if record.PayStatus == 2 { // 2 代表已支付
			log.L.Info("订单已处理过，跳过", zap.String("order_sn", orderSn))
			return nil
		}

		// 2. 更新支付流水表
		rawJson, _ := json.Marshal(notify)
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
		if strings.HasPrefix(orderSn, "T") {
			return p.processTicketOrderPaySuccess(tx, orderSn, tradeType)
		}
		var order models.Order
		if err := tx.Where("order_sn = ? AND status = ?", orderSn, 10).First(&order).Error; err != nil {
			return fmt.Errorf("获取订单失败: %w", err)
		}
		// 3. 更新主订单状态
		result := tx.Model(&models.Order{}).
			Where("order_sn = ? AND status = ?", orderSn, 10).
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
		var items []models.OrderItem
		if err := tx.Where("order_sn = ?", orderSn).Find(&items).Error; err != nil {
			return fmt.Errorf("获取订单明细失败: %w", err)
		}
		// 库存在下单时已预扣减，这里只做销量更新和自动下架检测
		for _, item := range items {
			// A. 增加商品销量
			if err := tx.Model(&models.Product{}).
				Where("id = ?", item.ProductID).
				UpdateColumn("sales_volume", gorm.Expr("sales_volume + ?", item.Quantity)).Error; err != nil {
				return fmt.Errorf("更新商品销量失败: %w", err)
			}

			// B. 自动下架检测：如果库存为0，自动将状态设为下架 (0)
			var prod models.Product
			if err := tx.Select("stock").First(&prod, item.ProductID).Error; err == nil {
				if prod.Stock == 0 {
					tx.Model(&prod).Update("status", 0)
					log.L.Info("检测到库存为0，商品已自动下架", zap.Uint64("product_id", item.ProductID))
				}
			}
		}

		return nil
	})
}

func (p *PayService) processTicketOrderPaySuccess(tx *gorm.DB, orderNo string, tradeType string) error {
	now := time.Now()
	var paidOrder models.TicketOrder
	result := tx.Model(&models.TicketOrder{}).
		Where("order_no = ? AND status = ?", orderNo, models.TicketOrderStatusPending).
		Updates(map[string]any{
			"status":     models.TicketOrderStatusUsable,
			"pay_method": tradeType,
			"pay_time":   now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		var order models.TicketOrder
		if err := tx.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
			return fmt.Errorf("获取票务订单失败: %w", err)
		}
		if order.Status == models.TicketOrderStatusUsable || order.Status == models.TicketOrderStatusUsed {
			return nil
		}
		return fmt.Errorf("票务订单状态不可支付: %d", order.Status)
	}
	if err := tx.Where("order_no = ?", orderNo).First(&paidOrder).Error; err != nil {
		return fmt.Errorf("获取票务订单失败: %w", err)
	}
	return p.rewardTicketOrderPoints(tx, paidOrder)
}

func (p *PayService) rewardTicketOrderPoints(tx *gorm.DB, order models.TicketOrder) error {
	reward := (order.ActualPrice + 500) / 1000
	if reward <= 0 {
		return nil
	}
	var exists int64
	if err := tx.Model(&models.PointsLog{}).
		Where("user_id = ? AND source_id = ? AND change_type = ?", uint64(order.UserID), order.OrderNo, models.TypeOrderReward).
		Count(&exists).Error; err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}
	var account models.UserPoint
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", uint64(order.UserID)).First(&account).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		account = models.UserPoint{UserID: uint64(order.UserID), Balance: 0}
		if err := tx.Create(&account).Error; err != nil {
			return err
		}
	}
	newBalance := account.Balance + reward
	if err := tx.Model(&models.UserPoint{}).Where("user_id = ?", uint64(order.UserID)).Updates(map[string]any{
		"balance":      newBalance,
		"total_earned": gorm.Expr("total_earned + ?", reward),
	}).Error; err != nil {
		return err
	}
	return tx.Create(&models.PointsLog{
		UserID:     uint64(order.UserID),
		Amount:     reward,
		Balance:    newBalance,
		ChangeType: models.TypeOrderReward,
		SourceID:   order.OrderNo,
		Remark:     "票务订单消费返积分",
		Status:     1,
	}).Error
}

func (p *PayService) ApplyWechatRefund(ctx context.Context, weChatClient *core.Client, refundNo string) error {
	var refund models.Refund
	var order models.TicketOrder
	var payRecord models.PayRecord

	if err := p.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("refund_no = ?", refundNo).First(&refund).Error; err != nil {
			return fmt.Errorf("退款单不存在: %w", err)
		}
		if refund.Status != models.RefundStatusAuditing {
			if refund.Status == models.RefundStatusRunning {
				return nil
			}
			return fmt.Errorf("当前退款状态不可发起微信退款")
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", refund.OrderID).First(&order).Error; err != nil {
			return fmt.Errorf("订单不存在: %w", err)
		}
		if order.Status != models.TicketOrderStatusRefunding {
			return fmt.Errorf("订单状态不可退款")
		}
		if order.ActualPrice > 0 {
			if err := tx.Where("order_sn = ? AND pay_status = 2", order.OrderNo).First(&payRecord).Error; err != nil {
				return fmt.Errorf("支付流水不存在或未支付: %w", err)
			}
		}
		return nil
	}); err != nil {
		return err
	}
	if order.ActualPrice <= 0 || refund.RefundAmount <= 0 {
		return p.completeLocalTicketRefund(ctx, refund, order, "纯积分/零元订单无需微信退款")
	}

	notifyURL := p.Config.WechatPayConfig.RefundNotifyURL
	if notifyURL == "" {
		notifyURL = strings.Replace(p.Config.WechatPayConfig.NotifyURL, "/notify", "/refund-notify", 1)
	}
	currency := "CNY"
	svc := refunddomestic.RefundsApiService{Client: weChatClient}
	wxRefund, _, err := svc.Create(ctx, refunddomestic.CreateRequest{
		OutTradeNo:  core.String(order.OrderNo),
		OutRefundNo: core.String(refund.RefundNo),
		Reason:      core.String(refund.Reason),
		NotifyUrl:   core.String(notifyURL),
		Amount: &refunddomestic.AmountReq{
			Refund:   core.Int64(refund.RefundAmount),
			Total:    core.Int64(order.ActualPrice),
			Currency: core.String(currency),
		},
	})
	if err != nil {
		return fmt.Errorf("微信退款申请失败: %w", err)
	}

	updates := map[string]any{
		"status":        models.RefundStatusRunning,
		"wechat_status": "PROCESSING",
	}
	if wxRefund != nil {
		if wxRefund.RefundId != nil {
			updates["wechat_refund_id"] = *wxRefund.RefundId
		}
		if wxRefund.Status != nil {
			updates["wechat_status"] = string(*wxRefund.Status)
			updates["status"] = refundStatusFromWechat(*wxRefund.Status)
		}
	}

	return p.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Refund{}).Where("refund_no = ?", refund.RefundNo).Updates(updates).Error; err != nil {
			return err
		}
		status := updates["status"].(int8)
		if err := tx.Model(&models.TicketOrder{}).Where("id = ?", refund.OrderID).Update("status", orderStatusFromRefund(status)).Error; err != nil {
			return err
		}
		if status == models.RefundStatusSuccess {
			if err := p.refundTicketOrderPoints(tx, order); err != nil {
				return err
			}
		}
		return tx.Create(&models.RefundLog{
			RefundID:    refund.ID,
			Status:      refundLogStatus(status),
			Description: "微信退款已受理",
		}).Error
	})
}

func (p *PayService) ProcessRefundNotify(ctx context.Context, wxRefund *refunddomestic.Refund) error {
	if wxRefund == nil || wxRefund.OutRefundNo == nil {
		return fmt.Errorf("退款回调参数不完整")
	}
	refundNo := *wxRefund.OutRefundNo
	status, wechatStatus, err := refundStatusFromWechatRefund(wxRefund)
	if err != nil {
		return err
	}

	return p.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var refund models.Refund
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("refund_no = ?", refundNo).First(&refund).Error; err != nil {
			return err
		}
		if refund.Status == models.RefundStatusSuccess {
			return tx.Model(&models.TicketOrder{}).
				Where("id = ? AND status <> ?", refund.OrderID, models.TicketOrderStatusRefundSuccess).
				Update("status", models.TicketOrderStatusRefundSuccess).Error
		}
		updates := map[string]any{
			"status":        status,
			"wechat_status": wechatStatus,
		}
		if wxRefund.RefundId != nil {
			updates["wechat_refund_id"] = *wxRefund.RefundId
		}
		if err := tx.Model(&refund).Updates(updates).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.TicketOrder{}).
			Where("id = ?", refund.OrderID).
			Update("status", orderStatusFromRefund(status)).Error; err != nil {
			return err
		}
		if status == models.RefundStatusSuccess {
			var order models.TicketOrder
			if err := tx.First(&order, refund.OrderID).Error; err == nil {
				_ = tx.Model(&models.TicketSpec{}).
					Where("id = ?", order.TicketSpecID).
					UpdateColumn("sold_count", gorm.Expr("GREATEST(sold_count - ?, 0)", order.Quantity)).Error
				if err := p.refundTicketOrderPoints(tx, order); err != nil {
					return err
				}
			}
		}
		return tx.Create(&models.RefundLog{
			RefundID:    refund.ID,
			Status:      refundLogStatus(status),
			Description: "微信退款回调：" + wechatStatus,
		}).Error
	})
}

func (p *PayService) SyncWechatRefund(ctx context.Context, weChatClient *core.Client, refundNo string) (*models.Refund, error) {
	if refundNo == "" {
		return nil, fmt.Errorf("退款单号不能为空")
	}
	if weChatClient == nil {
		return nil, fmt.Errorf("微信支付客户端未初始化")
	}
	svc := refunddomestic.RefundsApiService{Client: weChatClient}
	wxRefund, _, err := svc.QueryByOutRefundNo(ctx, refunddomestic.QueryByOutRefundNoRequest{
		OutRefundNo: core.String(refundNo),
	})
	if err != nil {
		return nil, fmt.Errorf("查询微信退款失败: %w", err)
	}
	if err := p.ProcessRefundNotify(ctx, wxRefund); err != nil {
		return nil, err
	}
	var refund models.Refund
	if err := p.DB.WithContext(ctx).Where("refund_no = ?", refundNo).First(&refund).Error; err != nil {
		return nil, err
	}
	return &refund, nil
}

func (p *PayService) completeLocalTicketRefund(ctx context.Context, refund models.Refund, order models.TicketOrder, description string) error {
	return p.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Refund{}).Where("id = ? AND status <> ?", refund.ID, models.RefundStatusSuccess).Updates(map[string]any{
			"status":        models.RefundStatusSuccess,
			"wechat_status": "LOCAL_SUCCESS",
		}).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.TicketOrder{}).Where("id = ?", order.ID).Update("status", models.TicketOrderStatusRefundSuccess).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.TicketSpec{}).
			Where("id = ?", order.TicketSpecID).
			UpdateColumn("sold_count", gorm.Expr("GREATEST(sold_count - ?, 0)", order.Quantity)).Error; err != nil {
			return err
		}
		if err := p.refundTicketOrderPoints(tx, order); err != nil {
			return err
		}
		return tx.Create(&models.RefundLog{
			RefundID:    refund.ID,
			Status:      "退款成功",
			Description: description,
		}).Error
	})
}

func (p *PayService) refundTicketOrderPoints(tx *gorm.DB, order models.TicketOrder) error {
	if order.PointsAmount <= 0 {
		return nil
	}
	var count int64
	if err := tx.Model(&models.PointsLog{}).
		Where("user_id = ? AND source_id = ? AND change_type = ?", uint64(order.UserID), order.OrderNo, models.TypeOrderRefund).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	var account models.UserPoint
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", uint64(order.UserID)).
		First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			account = models.UserPoint{UserID: uint64(order.UserID)}
			if err := tx.Create(&account).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}
	newBalance := account.Balance + order.PointsAmount
	if err := tx.Model(&models.UserPoint{}).
		Where("user_id = ?", uint64(order.UserID)).
		Updates(map[string]any{
			"balance":    newBalance,
			"total_used": gorm.Expr("GREATEST(total_used - ?, 0)", order.PointsAmount),
		}).Error; err != nil {
		return err
	}
	return tx.Create(&models.PointsLog{
		UserID:     uint64(order.UserID),
		Amount:     order.PointsAmount,
		Balance:    newBalance,
		ChangeType: models.TypeOrderRefund,
		SourceID:   order.OrderNo,
		Remark:     "票务订单退款返还积分",
		Status:     1,
	}).Error
}

func refundStatusFromWechat(status refunddomestic.Status) int8 {
	switch status {
	case refunddomestic.STATUS_SUCCESS:
		return models.RefundStatusSuccess
	case refunddomestic.STATUS_CLOSED, refunddomestic.STATUS_ABNORMAL:
		return models.RefundStatusRejected
	default:
		return models.RefundStatusRunning
	}
}

func refundStatusFromWechatRefund(wxRefund *refunddomestic.Refund) (int8, string, error) {
	if wxRefund == nil {
		return 0, "", fmt.Errorf("退款回调参数不完整")
	}
	if wxRefund.Status != nil {
		wechatStatus := string(*wxRefund.Status)
		return refundStatusFromWechat(*wxRefund.Status), wechatStatus, nil
	}
	if wxRefund.SuccessTime != nil {
		return models.RefundStatusSuccess, string(refunddomestic.STATUS_SUCCESS), nil
	}
	return 0, "", fmt.Errorf("退款回调缺少退款状态")
}

func orderStatusFromRefund(status int8) int8 {
	switch status {
	case models.RefundStatusSuccess:
		return models.TicketOrderStatusRefundSuccess
	case models.RefundStatusRejected:
		return models.TicketOrderStatusRefundReject
	default:
		return models.TicketOrderStatusRefunding
	}
}

func refundLogStatus(status int8) string {
	switch status {
	case models.RefundStatusSuccess:
		return "退款成功"
	case models.RefundStatusRejected:
		return "退款拒绝"
	default:
		return "退款中"
	}
}

func (p *PayService) GetOrderReceipt(ctx context.Context, orderSn string, userId int) (*types.OrderReceiptResponse, error) {
	var order models.Order
	var payRecord models.PayRecord
	var items []models.OrderItem
	//1.获取订单信息
	if err := p.DB.WithContext(ctx).
		Model(&models.Order{}).Where("order_sn = ? AND user_id = ?", orderSn, userId).First(&order).Error; err != nil {
		return nil, fmt.Errorf("获取订单失败: %w", err)
	}
	//2、获取支付流水信息
	p.DB.WithContext(ctx).Where("order_sn = ?", orderSn).First(&payRecord)
	//3、获取订单明细
	if err := p.DB.WithContext(ctx).Where("order_sn = ?", orderSn).Find(&items).Error; err != nil {
		return nil, fmt.Errorf("获取订单明细失败: %w", err)
	}
	//4、组装响应体
	resp := &types.OrderReceiptResponse{
		OrderSn:       order.OrderSn,
		TransactionId: payRecord.TransactionId,
		Status:        order.Status,
		StatusText:    "支付成功",
		PayTime:       "",
		TotalAmount:   float64(order.TotalAmount) / 100.0,
	}

	if order.PaidAt != nil {
		resp.PayTime = order.PaidAt.Format("2006-01-02 15:04:05")
	}
	for _, item := range items {
		resp.Items = append(resp.Items, types.ReceiptItem{
			ProductName:  item.ProductName,
			ProductPrice: float64(item.ProductPrice) / 100.0,
			Quantity:     item.Quantity,
			Subtotal:     float64(item.SubtotalAmount) / 100.0,
			CoverImage:   item.CoverImage,
		})
	}
	return resp, nil
}

// CancelExpiredOrders 取消超过指定分钟数未支付的订单，并回滚库存
func (p *PayService) CancelExpiredOrders(ctx context.Context, expireMinutes int) (int64, error) {
	deadline := time.Now().Add(-time.Duration(expireMinutes) * time.Minute)

	var expiredOrders []models.Order
	if err := p.DB.WithContext(ctx).
		Where("status = 10 AND created_at < ?", deadline).
		Find(&expiredOrders).Error; err != nil {
		return 0, fmt.Errorf("查询过期订单失败: %w", err)
	}

	if len(expiredOrders) == 0 {
		return 0, nil
	}

	var cancelled int64
	for _, order := range expiredOrders {
		err := p.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			// 更新订单状态为已取消(30)
			result := tx.Model(&models.Order{}).
				Where("order_sn = ? AND status = 10", order.OrderSn).
				Update("status", 30)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return nil // 已被其他goroutine处理
			}

			// 回滚库存
			var items []models.OrderItem
			if err := tx.Where("order_sn = ?", order.OrderSn).Find(&items).Error; err != nil {
				return err
			}
			for _, item := range items {
				if err := tx.Model(&models.Product{}).
					Where("id = ?", item.ProductID).
					UpdateColumn("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {
					return fmt.Errorf("回滚库存失败: %w", err)
				}
			}

			// 更新支付流水状态
			tx.Model(&models.PayRecord{}).
				Where("order_sn = ? AND pay_status = 0", order.OrderSn).
				Update("pay_status", 4) // 4: 已关闭

			cancelled++
			return nil
		})
		if err != nil {
			log.L.Error("取消过期订单失败", zap.String("order_sn", order.OrderSn), zap.Error(err))
		}
	}

	return cancelled, nil
}
