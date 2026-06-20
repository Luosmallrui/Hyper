# 订单列表、继续支付与入驻审核接口说明

本文对应小程序订单页与用户中心“我要入驻”流程。

默认请求头：

```http
Authorization: Bearer <access_token>
Content-Type: application/json
```

## 1. 我的订单列表

```http
GET /api/v1/order/list?page=1&size=10
```

支持筛选：

```http
GET /api/v1/order/list?page=1&size=10&status=0
GET /api/v1/order/list?page=1&size=10&refund_status=refunding
```

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| page | number | 否 | 页码，默认 `1` |
| size | number | 否 | 每页数量，默认 `10` |
| status | number | 否 | 普通订单状态 |
| refund_status | string | 否 | 售后状态 |

普通订单状态：

| status | 说明 |
|---|---|
| 0 | 待支付 |
| 1 | 待使用 |
| 2 | 已使用 |
| 3 | 已取消 |
| 4 | 退款中 |
| 5 | 已退款 |
| 6 | 已驳回 |

售后状态：

| refund_status | 说明 |
|---|---|
| pending_review | 待审核 |
| refunding | 退款中 |
| refunded | 已退款 |
| rejected | 已驳回 |
| cancelled | 已取消 |

### 响应

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "list": [
      {
        "order_no": "T20260619145616f26eeff2",
        "status": 0,
        "refund_status": "",
        "refund_status_text": "",
        "actual_price": 100,
        "total_price": 100,
        "quantity": 1,
        "buyer_name": "张三",
        "buyer_id_card": "5001********0817",
        "created_at": "2026-06-19T14:56:00+08:00",
        "expire_time": "2026-06-19T15:11:00+08:00",
        "pay_time": null,
        "activity": {
          "id": 10,
          "name": "jjjj",
          "poster_list": "https://cdn.hypercn.cn/..."
        },
        "ticket_spec": {
          "id": 1,
          "name": "1元"
        },
        "viewers": []
      }
    ],
    "total": 1
  }
}
```

说明：

- 不传 `status` 和 `refund_status` 时返回全部订单。
- 传 `status=0` 返回待支付订单。
- 传 `status=1` 返回待使用订单。
- 传 `refund_status` 时按最新售后单状态筛选。
- 0 元订单或积分全额抵扣订单会自动归为 `status=1`。

## 2. 待支付订单继续支付

```http
POST /api/v1/pay/prepay
```

请求：

```json
{
  "order_no": "T20260619145616f26eeff2"
}
```

响应：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "prepay_id": "wx...",
    "appId": "wx...",
    "timeStamp": "1781929913",
    "nonceStr": "random-string",
    "package": "prepay_id=wx...",
    "signType": "RSA",
    "paySign": "signature",
    "out_trade_no": "T20260619145616f26eeff2"
  }
}
```

失败示例：

```json
{
  "code": 400,
  "msg": "当前订单状态不可支付"
}
```

限制：

- 只允许 `status=0` 的票务订单继续支付。
- 已支付、已取消、退款中、已退款、已使用订单不能继续支付。
- 后端不会重新创建订单。

## 3. 查询入驻审核状态

```http
GET /api/v1/organizer/audit-status
```

未提交：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "status": 0,
    "reject_reason": ""
  }
}
```

审核中：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "type": "venue",
    "status": 1,
    "reject_reason": "",
    "submitted_at": "2026-06-20T13:21:00+08:00"
  }
}
```

已通过：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "type": "venue",
    "status": 2,
    "reject_reason": "",
    "submitted_at": "2026-06-20T13:21:00+08:00",
    "reviewed_at": "2026-06-20T14:00:00+08:00"
  }
}
```

已驳回：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "type": "venue",
    "status": 3,
    "reject_reason": "资料不完整",
    "submitted_at": "2026-06-20T13:21:00+08:00",
    "reviewed_at": "2026-06-20T14:00:00+08:00"
  }
}
```

状态说明：

| status | 说明 | 前端行为 |
|---|---|---|
| 0 | 未提交 | 展示入驻申请表单 |
| 1 | 审核中 | 展示审核中，不允许重复提交，不进入管理后台 |
| 2 | 已通过 | 允许进入管理后台 |
| 3 | 已驳回 | 展示驳回原因，允许重新提交 |

## 4. 提交入驻申请

```http
POST /api/v1/organizer/apply
```

请求：

```json
{
  "type": "venue",
  "name": "主办方名称",
  "logo": "https://cdn.hypercn.cn/...",
  "province": "广东省",
  "city": "广州市",
  "district": "越秀区"
}
```

成功响应：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "application_id": 123,
    "status": 1,
    "submitted_at": "2026-06-20T13:21:00+08:00"
  }
}
```

重复提交响应：

```json
{
  "code": 409,
  "msg": "入驻申请正在审核中，请勿重复提交"
}
```

规则：

- 审核中 `status=1` 不允许重复提交。
- 已通过 `status=2` 不允许重复提交。
- 已驳回 `status=3` 允许重新提交，提交后状态变为 `1`。
- 提交成功后，再次查询 `/api/v1/organizer/audit-status` 必须返回 `status=1`。
