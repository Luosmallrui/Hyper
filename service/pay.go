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
		PayedAt:    *order.PaidAt,
		OutTradeNo: order.OrderSn,
	}
	if orderItems.ConsumeType == "ticket" {
		// 把活动开始时间作为附加信息返回给前端(先固定了，后续加上活动时间字段)
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
	if req.Amount <= 0 {
		return respData, fmt.Errorf("invalid pay amount")
	}
	if req.Openid == "" {
		return respData, fmt.Errorf("openid is required")
	}

	wechatCfg := p.Config.WechatPayConfig
	orderSn := utilBase.GenerateOrderSn(req.UserId)

	// 1. 事务内：创建订单 + 订单明细 + 支付流水（只做 DB）
	err := p.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var product models.Product
		if err := tx.Where("id = ? AND status = 1", req.ProductId).First(&product).Error; err != nil {
			return fmt.Errorf("订单不存在或已下架")
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
		var order models.Order
		if err := tx.Where("order_sn = ? AND status = ?", orderSn, 10).First(&order).Error; err != nil {
			return fmt.Errorf("获取订单失败: %w", err)
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
		var items []models.OrderItem
		if err := tx.Where("order_sn = ?", order.ID).Find(&items).Error; err != nil {
			return fmt.Errorf("获取订单明细失败: %w", err)
		}
		for _, item := range items {
			res := tx.Model(&models.Product{}).
				Where("id = ? AND stock >= ?", item.ProductID, item.Quantity).
				UpdateColumn("stock", gorm.Expr("stock - ?", item.Quantity))
			if res.Error != nil {
				return fmt.Errorf("商品 [%s] 扣减库存失败: %w", item.ProductName, res.Error)
			}
			if res.RowsAffected == 0 {
				return fmt.Errorf("商品 [%s] 库存不足，无法完成支付回调业务", item.ProductName)
			}
			// B. 增加商品销量
			if err := tx.Model(&models.Product{}).
				Where("id = ?", item.ProductID).
				UpdateColumn("sales_volume", gorm.Expr("sales_volume + ?", item.Quantity)).Error; err != nil {
				return fmt.Errorf("更新商品销量失败: %w", err)
			}

			// C. 自动下架检测：如果库存扣减后变为 0，自动将状态设为下架 (0)
			var prod models.Product
			if err := tx.Select("stock").First(&prod, item.ProductID).Error; err == nil {
				if prod.Stock == 0 {
					tx.Model(&prod).Update("status", 0)
					log.L.Info("检测到库存为0，商品已自动下架", zap.Uint64("product_id", item.ProductID))
				}
			}
		}
		// 4. 【关键】执行后续业务逻辑
		// 例如：给用户账户加钱、增加积分、如果是商品则通知仓库发货
		// if err := s.AccountService.AddBalance(record.UserID, record.AmountTotal); err != nil {
		//     return err
		// }

		return nil
	})
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
	if err := p.DB.WithContext(ctx).Where("order_sn", order.ID).Find(&items).Error; err != nil {
		return nil, fmt.Errorf("获取订单明细失败: %w", err)
	}
	//4、组装响应体
	resp := &types.OrderReceiptResponse{
		OrderSn:       order.OrderSn,
		TransactionId: payRecord.TransactionId,
		Status:        order.Status,
		StatusText:    "支付成功", // 可根据 order.Status 进一步判断
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
