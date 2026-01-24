package types

type PrepayRequest struct {
	Description string `json:"description" binding:"required"`  // 商品描述
	OutTradeNo  string `json:"out_trade_no" binding:"required"` // 商户订单号
	Amount      int64  `json:"amount" binding:"required,min=1"` // 金额（分）
	Openid      string `json:"openid" binding:"required"`       // 用户openid
	Attach      string `json:"attach"`                          // 附加数据（可选）
}
