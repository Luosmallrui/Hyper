package types

import "time"

type PrepayRequest struct {
	Description string `json:"description"`  // 商品/票券描述
	OutTradeNo  string `json:"out_trade_no"` // 商户订单号
	OrderNo     string `json:"order_no"`     // 票务订单号，传这个时走 ticket_orders 支付
	Amount      int64  `json:"amount"`       // 金额（分），商品支付必传；票务支付由后端读取订单金额
	Openid      string `json:"openid"`       // 用户openid，由后端从 token 写入
	UserId      int    `json:"user_id"`
	Attach      string `json:"attach"`     // 附加数据（可选）
	ProductId   uint64 `json:"product_id"` // 购买的商品ID，旧商品支付使用
	Quantity    uint32 `json:"quantity"`   // 购买数量，旧商品支付使用
}

type PrepayWithRequestPaymentResponse struct {
	PrepayId   string `json:"prepay_id"` // 预支付交易会话标识
	Appid      string `json:"appId"`     // 应用ID
	TimeStamp  string `json:"timeStamp"` // 时间戳
	NonceStr   string `json:"nonceStr"`  // 随机字符串
	Package    string `json:"package"`   // 订单详情扩展字符串
	SignType   string `json:"signType"`  // 签名方式
	PaySign    string `json:"paySign"`   // 签名
	OutTradeNo string `json:"out_trade_no"`
}

type OrderDetail struct {
	Name       string            `json:"name"`         //商品名
	Avatar     string            `json:"avatar"`       //商品图标
	Price      int64             `json:"price"`        //价格 单位是分
	Quantity   int               `json:"quantity"`     //数量 单位是个
	OutTradeNo string            `json:"out_trade_no"` //订单编号
	Status     int8              `json:"status"`       // 支付状态 ：10:待支付, 20:已支付
	PayedAt    time.Time         `json:"payed_at"`     // 支付时间
	Attach     map[string]string `json:"exp"`          //附加信息 比如卖的是票 那就是活动的时间
}
