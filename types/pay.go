package types

type PrepayRequest struct {
	Description string `json:"description" binding:"required"`  // 商品描述
	OutTradeNo  string `json:"out_trade_no" binding:"required"` // 商户订单号
	Amount      int64  `json:"amount" binding:"required,min=1"` // 金额（分）
	Openid      string `json:"openid" binding:"required"`       // 用户openid
	UserId      int    `json:"user_id"`
	Attach      string `json:"attach"`                            // 附加数据（可选）
	ProductId   uint64 `json:"product_id" binding:"required"`     // 购买的商品ID
	Quantity    uint32 `json:"quantity" binding:"required,min=1"` // 购买数量
}

type PrepayWithRequestPaymentResponse struct {
	PrepayId  string `json:"prepay_id"` // 预支付交易会话标识
	Appid     string `json:"appId"`     // 应用ID
	TimeStamp string `json:"timeStamp"` // 时间戳
	NonceStr  string `json:"nonceStr"`  // 随机字符串
	Package   string `json:"package"`   // 订单详情扩展字符串
	SignType  string `json:"signType"`  // 签名方式
	PaySign   string `json:"paySign"`   // 签名
}
