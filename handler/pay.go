package handler

import (
	"Hyper/config"
	"Hyper/pkg/context"
	"Hyper/pkg/log"
	"Hyper/pkg/response"
	"Hyper/service"
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
)

// PrepayRequest 预支付请求参数
type PrepayRequest struct {
	Description string `json:"description" binding:"required"`  // 商品描述
	OutTradeNo  string `json:"out_trade_no" binding:"required"` // 商户订单号
	Amount      int64  `json:"amount" binding:"required,min=1"` // 金额（分）
	Openid      string `json:"openid" binding:"required"`       // 用户openid
	Attach      string `json:"attach"`                          // 附加数据（可选）
}

type Pay struct {
	Config        *config.WechatPayConfig
	PayService    service.IPayService
	wechatClient  *core.Client // 微信支付客户端（复用）
	MchPrivateKey *rsa.PrivateKey
}

func (p *Pay) RegisterRouter(r gin.IRouter) {
	pay := r.Group("/v1/pay")
	{
		pay.POST("/prepay", context.Wrap(p.Prepay))
		pay.POST("/notify", context.Wrap(p.PayNotify))              // 支付回调
		pay.GET("/query/:out_trade_no", context.Wrap(p.QueryOrder)) // 查询订单
	}
}

// NewPay 创建支付处理器
func NewPay(cfg *config.WechatPayConfig, payService service.IPayService) *Pay {
	p := &Pay{
		Config:     cfg,
		PayService: payService,
	}

	// 初始化时创建微信支付客户端
	if err := p.initWechatClient(); err != nil {
		return nil
	}

	return p
}

// initWechatClient 初始化微信支付客户端（只执行一次）
func (p *Pay) initWechatClient() error {
	// 1. 加载商户私钥
	mchPrivateKey, err := utils.LoadPrivateKeyWithPath(p.Config.MchPrivateKeyPath)
	if err != nil {
		return fmt.Errorf("加载商户私钥失败: %w", err)
	}
	p.MchPrivateKey = mchPrivateKey

	// 2. 加载微信支付公钥（如果有公钥文件的话）
	// 注意：新版建议使用平台证书序列号模式，而不是公钥文件
	// 这里保留了公钥加载的逻辑，实际使用时可以根据需要选择
	wechatPayPublicKey, err := utils.LoadPublicKeyWithPath(p.Config.WechatPayPublicKeyPath)
	if err != nil {
		return fmt.Errorf("加载微信支付公钥失败: %w", err)
	}

	// 3. 创建微信支付客户端
	opts := []core.ClientOption{
		option.WithWechatPayPublicKeyAuthCipher(
			p.Config.MchID,
			p.Config.MchCertificateSerialNumber,
			mchPrivateKey,
			p.Config.WechatPayPublicKeyID,
			wechatPayPublicKey,
		),
	}

	client, err := core.NewClient(base.Background(), opts...)
	if err != nil {
		return fmt.Errorf("创建微信支付客户端失败: %w", err)
	}

	p.wechatClient = client

	return nil
}

// Prepay 预支付下单
func (p *Pay) Prepay(c *gin.Context) error {
	ctx := c.Request.Context()

	// 1. 参数绑定和验证
	var req PrepayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(400, "参数错误: "+err.Error())
	}

	// 2. 业务逻辑验证（如订单号是否重复等）
	// 这里可以调用 PayService 进行业务验证
	// if err := p.PayService.ValidateOrder(ctx, req.OutTradeNo); err != nil {
	//     return response.NewError(400, "订单验证失败: "+err.Error())
	// }

	// 3. 调用微信支付预下单 API
	svc := jsapi.JsapiApiService{Client: p.wechatClient}

	prepayReq := jsapi.PrepayRequest{
		Appid:       core.String(p.Config.AppID),
		Mchid:       core.String(p.Config.MchID),
		Description: core.String(req.Description),
		OutTradeNo:  core.String(req.OutTradeNo),
		NotifyUrl:   core.String(p.Config.NotifyURL),
		Amount: &jsapi.Amount{
			Total: core.Int64(req.Amount),
		},
		Payer: &jsapi.Payer{
			Openid: core.String(req.Openid),
		},
	}

	// 如果有附加数据
	if req.Attach != "" {
		prepayReq.Attach = core.String(req.Attach)
	}

	// 发起请求
	resp, result, err := svc.PrepayWithRequestPayment(ctx, prepayReq)
	if err != nil {
		log.L.Error("微信预支付下单失败",
			zap.String("out_trade_no", req.OutTradeNo),
			zap.Error(err))
		return response.NewError(500, "支付下单失败")
	}

	// 4. 记录日志
	log.L.Info("微信预支付下单成功",
		zap.String("out_trade_no", req.OutTradeNo),
		zap.String("prepay_id", *resp.PrepayId),
		zap.Int("status", result.Response.StatusCode))

	// 5. 保存订单信息到数据库（可选）
	// if err := p.PayService.SaveOrder(ctx, req, resp); err != nil {
	//     log.L.Error("保存订单失败", zap.Error(err))
	// }

	// 6. 返回前端所需的支付参数
	response.Success(c, gin.H{
		"prepay_id": resp.PrepayId,
		"timestamp": resp.TimeStamp,
		"nonce_str": resp.NonceStr,
		"package":   resp.Package,
		"sign_type": resp.SignType,
		"pay_sign":  resp.PaySign,
	})
	return nil
}

// PayNotify 支付回调处理
func (p *Pay) PayNotify(c *gin.Context) error {
	ctx := c.Request.Context()
	// 1. 使用 `RegisterDownloaderWithPrivateKey` 注册下载器
	err := downloader.MgrInstance().RegisterDownloaderWithPrivateKey(ctx, p.MchPrivateKey, p.Config.MchCertificateSerialNumber, p.Config.MchID, p.Config.MchAPIv3Key)
	// 2. 获取商户号对应的微信支付平台证书访问器
	certificateVisitor := downloader.MgrInstance().GetCertificateVisitor(p.Config.MchID)
	// 3. 使用证书访问器初始化 `notify.Handler`
	handler, err := notify.NewRSANotifyHandler(p.Config.MchAPIv3Key, verifiers.NewSHA256WithRSAVerifier(certificateVisitor))
	if err != nil {
		log.L.Error("创建微信支付回调处理器失败", zap.Error(err))
		c.JSON(500, gin.H{
			"code":    "FAIL",
			"message": "init notify handler error",
		})
		return nil
	}

	transaction := new(payments.Transaction)
	notifyReq, err := handler.ParseNotifyRequest(ctx, c.Request, transaction)
	if err != nil {
		log.L.Error("微信支付回调验签或解密失败", zap.Error(err))
		c.JSON(500, gin.H{
			"code":    "FAIL",
			"message": err.Error(),
		})
		return nil
	}
	fmt.Println(notifyReq.Summary)
	fmt.Println(transaction.TransactionId)
	
	// TODO：幂等更新订单
	// p.PayService.PaySuccess(ctx, outTradeNo, transactionId, transaction)

	c.JSON(200, gin.H{
		"code": "SUCCESS",
		"data": notifyReq,
	})
	return nil
}

// QueryOrder 查询订单
func (p *Pay) QueryOrder(c *gin.Context) error {
	ctx := c.Request.Context()
	outTradeNo := c.Param("out_trade_no")

	if outTradeNo == "" {
		return response.NewError(400, "订单号不能为空")
	}

	// 调用微信支付查询订单 API
	svc := jsapi.JsapiApiService{Client: p.wechatClient}

	resp, result, err := svc.QueryOrderByOutTradeNo(ctx,
		jsapi.QueryOrderByOutTradeNoRequest{
			OutTradeNo: core.String(outTradeNo),
			Mchid:      core.String(p.Config.MchID),
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
