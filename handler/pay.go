package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/models"
	"Hyper/pkg/context"
	"Hyper/pkg/log"
	"Hyper/pkg/response"
	utilBase "Hyper/pkg/utils"
	"Hyper/service"
	"Hyper/types"
	base "context"
	"crypto/rsa"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PrepayRequest 预支付请求参数

type Pay struct {
	WechatPayConfig *config.WechatPayConfig
	Config          *config.Config
	PayService      service.IPayService
	wechatClient    *core.Client // 微信支付客户端（复用）
	MchPrivateKey   *rsa.PrivateKey
	MchPublicKey    *rsa.PublicKey
	DB              *gorm.DB
}

func (p *Pay) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(p.Config.Jwt.Secret))
	pay := r.Group("/v1/pay")
	{
		pay.POST("/prepay", authorize, context.Wrap(p.Prepay))
		pay.POST("/notify", context.Wrap(p.PayNotify))              // 支付回调
		pay.GET("/query/:out_trade_no", context.Wrap(p.QueryOrder)) // 查询订单
	}
}

// NewPay 创建支付处理器
func NewPay(cfg *config.Config, payService service.IPayService, Db *gorm.DB) *Pay {
	p := &Pay{
		WechatPayConfig: cfg.WechatPayConfig,
		PayService:      payService,
		DB:              Db,
		Config:          cfg,
	}

	// 初始化时创建微信支付客户端
	if err := p.initWechatClient(); err != nil {
		log.L.Info("init wechat client failed", zap.Error(err))
		return nil
	}

	return p
}

// initWechatClient 初始化微信支付客户端（只执行一次）
func (p *Pay) initWechatClient() error {
	// 1. 加载商户私钥
	mchPrivateKey, err := utils.LoadPrivateKeyWithPath(p.WechatPayConfig.MchPrivateKeyPath)
	if err != nil {
		return fmt.Errorf("加载商户私钥失败: %w", err)
	}
	p.MchPrivateKey = mchPrivateKey

	// 2. 加载微信支付公钥（如果有公钥文件的话）
	// 注意：新版建议使用平台证书序列号模式，而不是公钥文件
	// 这里保留了公钥加载的逻辑，实际使用时可以根据需要选择
	wechatPayPublicKey, err := utils.LoadPublicKeyWithPath(p.WechatPayConfig.WechatPayPublicKeyPath)
	if err != nil {
		return fmt.Errorf("加载微信支付公钥失败: %w", err)
	}

	p.MchPublicKey = wechatPayPublicKey
	// 3. 创建微信支付客户端
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(
			p.WechatPayConfig.MchID,
			p.WechatPayConfig.MchCertificateSerialNumber,
			mchPrivateKey,
			p.WechatPayConfig.MchAPIv3Key,
		),
	}

	client, err := core.NewClient(base.Background(), opts...)
	if err != nil {
		log.L.Error("new client failed", zap.Error(err))
		return fmt.Errorf("创建微信支付客户端失败: %w", err)
	}

	p.wechatClient = client
	log.L.Info("info", zap.Any("info", p.WechatPayConfig))

	return nil
}

// Prepay 预支付下单
// Prepay 预支付下单接口
func (p *Pay) Prepay(c *gin.Context) error {
	ctx := c.Request.Context()

	// 1. 参数绑定 (获取金额、描述、OpenID等)
	var req types.PrepayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(400, "参数错误: "+err.Error())
	}

	// 2. 获取当前用户 ID
	userId := c.GetInt("user_id")
	openId := c.GetString("openid")

	// 3. 生成全局唯一的业务订单号 (使用你喜欢的格式)
	orderSn := utilBase.GenerateOrderSn(userId)

	// 4. 开启数据库事务
	err := p.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// A. 创建本地主订单
		newOrder := &models.Order{
			UserID:      userId,
			OrderSn:     orderSn,
			TotalAmount: uint64(req.Amount), // 单位：分
			Description: req.Description,
			Status:      10, // 待支付
		}
		if err := tx.Create(newOrder).Error; err != nil {
			return err
		}

		// B. 创建初始支付流水
		payRecord := &models.PayRecord{
			OrderSn:     orderSn,
			PayPlatform: 1, // 1-微信
			AmountTotal: uint64(req.Amount),
			PayStatus:   0, // 支付中/待支付
		}
		if err := tx.Create(payRecord).Error; err != nil {
			return err
		}

		// C. 调用微信支付预下单 API
		svc := jsapi.JsapiApiService{Client: p.wechatClient}
		prepayReq := jsapi.PrepayRequest{
			Appid:       core.String(p.WechatPayConfig.AppID),
			Mchid:       core.String(p.WechatPayConfig.MchID),
			Description: core.String(req.Description),
			OutTradeNo:  core.String(orderSn), // 使用我们生成的单号
			NotifyUrl:   core.String(p.WechatPayConfig.NotifyURL),
			Amount: &jsapi.Amount{
				Total: core.Int64(req.Amount),
			},
			Payer: &jsapi.Payer{
				Openid: core.String(openId),
			},
		}

		resp, _, err := svc.PrepayWithRequestPayment(ctx, prepayReq)
		if err != nil {
			return fmt.Errorf("微信下单失败: %w", err)
		}

		// D. 将微信返回的 prepay_id 更新到流水表中，方便后续追踪
		if err := tx.Model(&models.PayRecord{}).
			Where("order_sn = ?", orderSn).
			Update("out_request_no", *resp.PrepayId).Error; err != nil {
			return err
		}

		// E. 将支付参数保存或直接准备返回
		c.Set("pay_params", resp)
		return nil
	})

	if err != nil {
		log.L.Error("创建订单失败", zap.Error(err))
		return response.NewError(500, "下单失败")
	}

	// 5. 返回给前端唤起支付所需的参数
	payParams, _ := c.Get("pay_params")
	response.Success(c, payParams)
	return nil
}

// PayNotify 支付回调处理
func (p *Pay) PayNotify(c *gin.Context) error {
	ctx := c.Request.Context()
	certificateVisitor := downloader.MgrInstance().GetCertificateVisitor(p.WechatPayConfig.MchID)
	handler, err := notify.NewRSANotifyHandler(p.WechatPayConfig.MchAPIv3Key, verifiers.NewSHA256WithRSAVerifier(certificateVisitor))
	if err != nil {
		log.L.Error("创建微信支付回调处理器失败", zap.Error(err))
		return response.NewError(500, err.Error())
	}

	transaction := new(payments.Transaction)
	notifyReq, err := handler.ParseNotifyRequest(ctx, c.Request, transaction)
	if err != nil {
		log.L.Error("微信支付回调验签或解密失败", zap.Error(err))
		return response.NewError(500, err.Error())
	}
	log.L.Info("pay notify", zap.Any("notifyReq", notifyReq), zap.Any("transaction", transaction))

	err = p.PayService.ProcessOrderPaySuccess(ctx, transaction)
	if err != nil {
		log.L.Error("处理订单回调业务失败", zap.Error(err))
		return response.NewError(500, "process failed")
	}

	response.Success(c, notifyReq)
	return nil
}

// QueryOrder 查询订单
func (p *Pay) QueryOrder(c *gin.Context) error {
	ctx := c.Request.Context()
	outTradeNo := c.Param("out_trade_no")
	if outTradeNo == "" {
		return response.NewError(400, "订单号不能为空")
	}
	svc := jsapi.JsapiApiService{Client: p.wechatClient}
	resp, result, err := svc.QueryOrderByOutTradeNo(ctx,
		jsapi.QueryOrderByOutTradeNoRequest{
			OutTradeNo: core.String(outTradeNo),
			Mchid:      core.String(p.WechatPayConfig.MchID),
		},
	)
	if err != nil {
		log.L.Error("查询订单失败",
			zap.String("out_trade_no", outTradeNo),
			zap.Error(err))
		return response.NewError(500, "查询订单失败")
	}
	log.L.Info("查询订单成功",
		zap.String("out_trade_no", outTradeNo),
		zap.Int("status", result.Response.StatusCode))
	response.Success(c, resp)
	return nil
}
