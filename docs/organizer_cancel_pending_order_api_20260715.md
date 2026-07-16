# 主办方取消未支付订单接口

更新时间：2026-07-15

## 目的

PC 主办方订单页不能调用用户侧的 `POST /api/v1/order/:order_no/cancel`。该接口按购票用户校验订单归属，主办方 Token 调用会返回订单不属于当前用户。

主办方端请改用本接口，仅用于取消当前主办方名下活动的未支付订单。已支付订单必须进入退款售后流程。

## 接口

```http
POST /api/v1/organizer/orders/:order_no/cancel
Authorization: Bearer <organizer_access_token>
Content-Type: application/json
```

路径参数：

| 参数 | 说明 |
| --- | --- |
| `order_no` | 待取消的票务订单号 |

请求体：

```json
{
  "reason_id": 1
}
```

`reason_id` 取自既有取消原因接口：

```http
GET /api/v1/order/cancel-reasons
Authorization: Bearer <access_token>
```

## 成功响应

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "order_no": "T2026053114300012ab34cd",
    "status": 3,
    "cancel_reason_id": 1,
    "cancelled_at": "2026-07-15T15:30:00+08:00"
  }
}
```

字段说明：

| 字段 | 说明 |
| --- | --- |
| `status` | 固定为 `3`，表示已取消 |
| `cancel_reason_id` | 本次提交的取消原因 ID |
| `cancelled_at` | 订单取消时间 |

## 业务规则

1. 仅已登录主办方可调用，且订单必须属于当前主办方名下活动。
2. 不属于当前主办方的订单返回 `404`，不泄露其他主办方订单信息。
3. 仅 `status=0` 的待支付订单允许取消。
4. 已支付、待使用、已核销、退款中、已退款等非待支付订单返回 `409`：

```json
{
  "code": 409,
  "msg": "已支付订单请通过退款售后流程处理"
}
```

5. 成功取消后，后端会自动回滚票种销量、返还该订单已使用的积分，并关闭未支付支付流水。
6. 同一订单已经取消时重复调用保持幂等，不会重复释放库存或重复返还积分。
7. 后端会写入主办方操作日志，动作编码为 `cancel_pending_order`。

## 前端接入要求

1. 将原调用地址从：

```http
POST /api/v1/order/:order_no/cancel
```

替换为：

```http
POST /api/v1/organizer/orders/:order_no/cancel
```

2. 仅订单 `status === 0` 时展示“取消订单”按钮。
3. 非待支付订单不展示取消按钮；已支付订单改为进入订单详情后申请退款或跳转售后处理。
4. 取消成功后刷新订单详情与订单列表，按返回的 `status=3` 展示“已取消”。
5. 收到 `409` 时直接展示后端消息，不要再次尝试取消。

## 验收

- 主办方可以取消自己活动的待支付订单。
- 取消后可售库存恢复，积分抵扣原路返还。
- 主办方不能取消其他主办方订单。
- 已支付订单不能被主办方直接取消。
- 连续点击或网络重试不会造成库存、积分重复变化。
