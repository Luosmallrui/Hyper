package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/pkg/log"
	"Hyper/pkg/response"
	utilBase "Hyper/pkg/utils"
	"Hyper/service"
	"Hyper/types"
	base "context"
	"crypto/rsa"
	"fmt"
	"net/http"

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
	wechatClient    *core.Client
	MchPrivateKey   *rsa.PrivateKey
	DB              *gorm.DB
}

func (p *Pay) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(p.Config.Jwt.Secret))
	pay := r.Group("/v1/pay")
	{
		pay.POST("/prepay", authorize, context.Wrap(p.Prepay))
		pay.POST("/notify", context.Wrap(p.PayNotify)) // 支付回调

		pay.GET("/query/:out_trade_no", context.Wrap(p.QueryOrder))     // 查询订单
		pay.GET("/receipt", authorize, context.Wrap(p.GetOrderReceipt)) // 获取订单电子回执
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

// initWechatClient 初始化微信支付客户端
func (p *Pay) initWechatClient() error {
	// 加载商户私钥
	mchPrivateKey, err := utils.LoadPrivateKeyWithPath(p.WechatPayConfig.MchPrivateKeyPath)
	if err != nil {
		return fmt.Errorf("加载商户私钥失败: %w", err)
	}
	p.MchPrivateKey = mchPrivateKey
	// 创建微信支付客户端
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

	return nil
}

// Prepay 预支付下单
func (p *Pay) Prepay(c *gin.Context) error {
	ctx := c.Request.Context()
	var req types.PrepayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		//return response.NewError(400, "参数错误: "+err.Error())
	}

	userId := c.GetInt("user_id")
	if req.OutTradeNo == "" {
		req.OutTradeNo = utilBase.GenerateOrderSn(userId)
	}
	openId := c.GetString("openid")
	req.UserId = userId
	req.Openid = openId

	resp, err := p.PayService.PreWeChatPay(ctx, p.wechatClient, req)
	if err != nil {
		log.L.Error("创建订单失败", zap.Error(err))
		return response.NewError(500, err.Error())
	}
	response.Success(c, resp)
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
	// --- ✅ 临时方案：直接解析 JSON 模拟 transaction 对象 ---
	//transaction := new(payments.Transaction)
	//if err := c.ShouldBindJSON(transaction); err != nil {
	//	log.L.Error("模拟回调解析 JSON 失败", zap.Error(err))
	//	return response.NewError(400, "参数格式错误")
	//}
	//
	//// 记录模拟数据日志
	//log.L.Info("收到模拟支付回调信号",
	//	zap.String("order_sn", *transaction.OutTradeNo),
	//	zap.String("tran_id", *transaction.TransactionId),
	//)

	err = p.PayService.ProcessOrderPaySuccess(ctx, transaction)
	if err != nil {
		log.L.Error("处理订单回调业务失败", zap.Error(err))
		return response.NewError(500, "process failed")
	}

	response.Success(c, notifyReq)
	//response.Success(c, gin.H{"status": "SUCCESS", "message": "模拟核销成功"})
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

// 在 RegisterRouter 中增加
// prod.GET("/receipt", auth, context.Wrap(p.GetOrderReceipt))

func (p *Pay) GetOrderReceipt(c *gin.Context) error {
	orderSn := c.Query("order_sn")
	if orderSn == "" {
		return response.NewError(http.StatusBadRequest, "订单号不能为空")
	}

	// 从中间件获取当前登录用户的 ID
	userId := c.GetInt("userID")
	//c.Set("user_id", 1)
	//// 这里会读取到你上面 Set 的 6
	//userId := c.GetInt("user_id")
	res, err := p.PayService.GetOrderReceipt(c.Request.Context(), orderSn, userId)
	if err != nil {
		return response.NewError(http.StatusNotFound, err.Error())
	}

	response.Success(c, res)
	return nil
}
