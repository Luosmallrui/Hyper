# 订单待支付取消、自动取消与退款状态修改说明

更新日期：2026-06-22

本文说明小程序端订单模块本次后端调整内容，方便前端同步对接。

## 1. 修改背景

本次针对小程序端用户订单列表与订单详情的以下问题做后端补齐：

- 待支付订单需要支持取消订单。
- 待支付订单取消后，需要返还下单时抵扣的积分。
- 待支付订单超过 15 分钟未支付，需要自动取消，并同样返还抵扣积分。
- 偶现微信支付无法唤起或卡顿时，需要降低重复支付风险。
- 用户提交退款申请后，实际状态是待审核，页面不应直接提示“退款中”。

## 2. 待支付订单取消

接口不变：

```http
POST /api/v1/order/:order_no/cancel
Authorization: Bearer <access_token>
```

请求：

```json
{
  "reason_id": 1
}
```

后端规则：

- 仅允许取消 `status=0` 的待支付订单。
- 取消成功后，订单状态更新为 `status=3` 已取消。
- 后端回滚票券销量 `ticket_specs.sold_count`。
- 如果订单使用了积分抵扣，后端返还 `ticket_orders.points_amount` 到用户积分余额。
- 积分返还会写入 `point_logs`，并按 `order_no + TypeOrderRefund` 做幂等，避免重复返还。
- 待支付流水会关闭：`pay_records.pay_status=4`。

## 3. 15 分钟自动取消

订单创建时已设置：

```text
expire_time = created_at + 15 分钟
```

本次新增自动取消逻辑：

- API 服务启动后，每分钟扫描过期票务订单。
- 命中条件：

```sql
status = 0
AND actual_price > 0
AND expire_time <= NOW()
```

- 自动取消后：
  - 订单状态更新为 `status=3`。
  - `cancel_reason` 写入：`超时未支付自动取消`。
  - 回滚票券销量。
  - 返还抵扣积分。
  - 关闭待支付流水。

兜底触发：

- 查询订单列表时会处理当前用户过期待支付订单。
- 查询订单详情时会处理当前用户过期待支付订单。
- 继续支付时如果订单已过期，会先取消订单，再返回错误。

## 4. 继续支付与重复支付防护

接口不变：

```http
POST /api/v1/pay/prepay
Authorization: Bearer <access_token>
```

请求：

```json
{
  "order_no": "T20260622120000abcd1234"
}
```

后端规则：

- 仅 `status=0` 待支付订单可继续支付。
- 如果订单已超过 `expire_time`，后端自动取消并返回错误，不再请求微信预支付。
- 后端始终使用同一个 `order_no` 作为微信支付 `out_trade_no`。
- 微信侧同一个 `out_trade_no` 不会生成两笔不同订单支付。
- 支付回调按订单状态幂等处理：
  - 待支付订单支付成功后更新为 `status=1`。
  - 如果订单已经是已支付/已使用，重复回调直接视为成功处理。
  - 如果订单已取消/已退款，则拒绝继续支付状态变更。

前端建议：

- 用户点击继续支付后，按钮进入 loading，避免连续点击。
- 如果接口返回订单已过期或已取消，刷新订单列表/详情。

## 5. 退款申请后的状态展示

接口不变：

```http
POST /api/v1/refund/apply
Authorization: Bearer <access_token>
```

请求：

```json
{
  "order_no": "T20260622120000abcd1234",
  "reason_id": 1
}
```

本次调整：

- 用户提交退款申请后，只创建 `refunds.status=0` 的退款单。
- 主订单 `ticket_orders.status` 暂不立即改成 `4` 退款中。
- 商家/管理员审核通过并发起退款时，主订单才更新为 `status=4` 退款中。
- 微信退款成功后，退款单更新为 `status=2`，订单更新为 `status=5` 已退款。

前端展示口径：

| 字段 | 值 | 展示文案 |
|---|---|---|
| `refund_status` | `pending_review` | 待审核 |
| `refund_status` | `refunding` | 退款中 |
| `refund_status` | `refunded` | 已退款 |
| `refund_status` | `rejected` | 已驳回 |
| `refund_status` | `cancelled` | 已取消 |

列表和详情应优先使用：

```json
{
  "refund_status": "pending_review",
  "refund_status_text": "待审核",
  "refund_no": "R20260622120000abcd"
}
```

不要仅凭主订单 `status=4` 判断“退款中”。

## 6. 订单列表/详情建议展示

待支付订单：

- `status=0`
- 展示“取消订单”按钮。
- 展示“继续支付”按钮。
- 若当前时间超过 `expire_time`，后端会自动取消；前端刷新后展示已取消。

退款待审核订单：

- 主订单可能仍为 `status=1`。
- 但存在：

```json
{
  "refund_status": "pending_review",
  "refund_status_text": "待审核"
}
```

- 页面展示“退款待审核”或“售后待审核”。
- 列表页可隐藏“申请退款”按钮。
- 详情页可展示售后进度。

退款中订单：

- 主订单 `status=4`。
- `refund_status=refunding`。
- 页面展示“退款中”。

## 7. 核销保护

本次新增核销保护：

- 如果订单存在 `待审核` 或 `退款中` 的退款单，扫码识别会返回失败。
- 确认核销时也会再次校验。

原因：

- 避免用户已提交退款申请后，线下仍被核销，导致售后状态和票务状态冲突。

## 8. 影响的后端文件

- `service/ticketing.go`
- `service/pay.go`
- `pkg/server/task.go`
- `pkg/server/gin.go`
- `cmd/api-server/wire_gen.go`

## 9. 验收点

- 待支付订单点击取消后，订单变为已取消。
- 使用积分抵扣的待支付订单取消后，积分余额增加。
- 待支付订单超过 15 分钟后，刷新列表/详情展示已取消。
- 过期订单点击继续支付，不能拉起微信支付。
- 快速重复点击继续支付，不会产生重复订单支付。
- 提交退款申请后，列表/详情展示“待审核”，不是“退款中”。
- 商家审核通过并发起退款后，列表/详情展示“退款中”。
- 退款待审核/退款中的订单不能被核销。
