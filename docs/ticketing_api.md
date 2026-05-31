# 派对票务系统 API 对接文档

Base URL:

```text
/api/v1
```

除特别说明外，所有接口都需要：

```http
Authorization: Bearer <access_token>
Content-Type: application/json
```

统一响应结构沿用后端现有 `response.Success`：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {}
}
```

错误响应：

```json
{
  "code": 500,
  "msg": "错误信息"
}
```

金额字段统一使用 **分**，例如 `8800` 表示 `88.00` 元。

---

## 1. 状态枚举

### 主办方状态 `organizer.status`

| 值 | 说明 |
|---|---|
| 0 | 待审核 |
| 1 | 审核中 |
| 2 | 已认证 |
| 3 | 审核未通过 |

### 活动状态 `activity.status`

| 值 | 说明 |
|---|---|
| 0 | 草稿 |
| 1 | 待审核 |
| 2 | 审核中 |
| 3 | 已上架 |
| 4 | 审核未通过 |

### 票务订单状态 `order.status`

| 值 | 说明 |
|---|---|
| 0 | 待支付 |
| 1 | 待使用 |
| 2 | 已使用 |
| 3 | 已取消 |
| 4 | 退款中 |
| 5 | 退款成功 |
| 6 | 退款拒绝 |

### 退款状态 `refund.status`

| 值 | 说明 |
|---|---|
| 0 | 审核中 |
| 1 | 退款中 |
| 2 | 退款成功 |
| 3 | 退款拒绝 |

### 核销员状态 `verifier.status`

| 值 | 说明 |
|---|---|
| 0 | 未激活 |
| 1 | 已激活 |

---

## 2. 文件上传

### 通用文件上传

```http
POST /api/v1/upload
Content-Type: multipart/form-data
Authorization: Bearer <access_token>
```

FormData:

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| file | File | 是 | 文件 |
| type | string | 否 | `poster_detail` / `poster_long` / `poster_list` / `poster_wechat` / `qualification_doc` / `avatar` / `organizer_logo` |

响应：

```json
{
  "code": 200,
  "data": {
    "url": "https://cdn.hypercn.cn/ticketing/poster_list/2026/05/31/123.jpg"
  }
}
```

---

## 3. 主办方模块

### 获取主办方信息

```http
GET /api/v1/organizer/info
```

响应：

```json
{
  "code": 200,
  "data": {
    "id": 1,
    "name": "Hyper Club",
    "logo": "https://cdn.xxx/logo.png",
    "status": 2,
    "level": "LV1",
    "service_fee_rate": 0,
    "join_days": 12,
    "basic_info": {
      "province": "北京市",
      "city": "北京市",
      "district": "朝阳区"
    },
    "account_info": {
      "bank_account_name": "张三",
      "bank_account_no": "6222...",
      "bank_name": "招商银行"
    }
  }
}
```

### 入驻申请

```http
POST /api/v1/organizer/apply
```

请求：

```json
{
  "name": "Hyper Club",
  "logo": "https://cdn.xxx/logo.png",
  "province": "北京市",
  "city": "北京市",
  "district": "朝阳区"
}
```

响应：

```json
{ "code": 200, "data": { "success": true } }
```

### 入驻审核状态

```http
GET /api/v1/organizer/audit-status
```

响应：

```json
{
  "code": 200,
  "data": {
    "status": 1,
    "reject_reason": ""
  }
}
```

### 更新主办方基本信息

```http
PUT /api/v1/organizer/basic
```

请求：

```json
{
  "name": "新主办方名称",
  "logo": "https://cdn.xxx/logo.png",
  "province": "上海市",
  "city": "上海市",
  "district": "黄浦区"
}
```

### 获取提现信息

```http
GET /api/v1/organizer/withdraw-info
```

响应：

```json
{
  "code": 200,
  "data": {
    "bank_account_name": "张三",
    "bank_account_no": "6222...",
    "bank_name": "招商银行"
  }
}
```

### 修改提现信息

```http
PUT /api/v1/organizer/withdraw-info
```

请求：

```json
{
  "bank_account_name": "张三",
  "bank_account_no": "6222000000000000000",
  "bank_name": "招商银行"
}
```

---

## 4. 活动模块

### 创建/分步保存活动

```http
POST /api/v1/activity/create
```

说明：

- 首次创建不传 `activity_id`
- 后续步骤传上一次返回的 `activity_id`
- 时间支持：`2026-05-31T20:00`、`2026-05-31 20:00:00`、RFC3339
- `price` 使用分

Step 1 请求：

```json
{
  "step": 1,
  "name": "周末电音派对",
  "share_title": "一起蹦到凌晨",
  "start_time": "2026-06-12T20:00",
  "end_time": "2026-06-13T02:00",
  "real_name_mode": 1,
  "minor_check": 1,
  "description": "<p>活动介绍</p>"
}
```

Step 2 请求：

```json
{
  "activity_id": 1,
  "step": 2,
  "province": "北京市",
  "city": "北京市",
  "district": "朝阳区",
  "address": "工体西路",
  "latitude": 39.9333,
  "longitude": 116.4533
}
```

Step 3 请求：

```json
{
  "activity_id": 1,
  "step": 3,
  "poster_detail": "https://cdn.xxx/detail.jpg",
  "poster_long": "https://cdn.xxx/long.jpg",
  "poster_list": "https://cdn.xxx/list.jpg",
  "poster_wechat": "https://cdn.xxx/wechat.jpg"
}
```

Step 4 请求：

```json
{
  "activity_id": 1,
  "step": 4,
  "ticket_specs": [
    {
      "name": "早鸟票",
      "is_enabled": 1,
      "sale_start": "2026-06-01T10:00",
      "sale_end": "2026-06-12T18:00",
      "price": 8800,
      "stock": 100,
      "purchase_limit": 2,
      "max_attendees": 1
    }
  ]
}
```

Step 5 请求：

```json
{
  "activity_id": 1,
  "step": 5,
  "qualification_doc": "https://cdn.xxx/doc.pdf"
}
```

响应：

```json
{
  "code": 200,
  "data": {
    "activity_id": 1
  }
}
```

### 获取活动详情

```http
GET /api/v1/activity/:id
```

响应核心结构：

```json
{
  "code": 200,
  "data": {
    "id": 1,
    "organizer_id": 1,
    "name": "周末电音派对",
    "status": 0,
    "ticket_specs": [
      {
        "id": 1,
        "activity_id": 1,
        "name": "早鸟票",
        "price": 8800,
        "stock": 100,
        "sold_count": 0
      }
    ],
    "organizer": {
      "id": 1,
      "name": "Hyper Club"
    }
  }
}
```

### 我的活动列表

```http
GET /api/v1/activity/my-list?page=1&size=10
```

响应：

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": 1,
        "name": "周末电音派对",
        "poster_list": "https://cdn.xxx/list.jpg",
        "start_time": "2026-06-12T20:00:00+08:00",
        "end_time": "2026-06-13T02:00:00+08:00",
        "status": 0
      }
    ],
    "total": 1
  }
}
```

### 活动搜索

```http
GET /api/v1/activity/search?keyword=电音
```

只返回已上架活动。

### 删除活动

```http
DELETE /api/v1/activity/:id
```

### 提交活动审核

```http
POST /api/v1/activity/:id/submit-audit
```

---

## 5. 票券配置模块

### 获取活动票券列表

```http
GET /api/v1/activity/:id/ticket-specs
```

响应：

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": 1,
        "name": "早鸟票",
        "is_enabled": 1,
        "sale_start": "2026-06-01T10:00:00+08:00",
        "sale_end": "2026-06-12T18:00:00+08:00",
        "price": 8800,
        "stock": 100,
        "sold_count": 0,
        "purchase_limit": 2,
        "max_attendees": 1
      }
    ]
  }
}
```

### 保存票券配置

```http
POST /api/v1/activity/:id/ticket-specs
```

请求：

```json
{
  "specs": [
    {
      "id": 1,
      "name": "早鸟票",
      "is_enabled": 1,
      "sale_start": "2026-06-01T10:00",
      "sale_end": "2026-06-12T18:00",
      "price": 8800,
      "stock": 100,
      "purchase_limit": 2,
      "max_attendees": 1
    }
  ]
}
```

说明：

- 有 `id` 表示更新
- 无 `id` 表示新增

### 删除票券规格

```http
DELETE /api/v1/ticket-spec/:id
```

---

## 6. 订单模块

### 创建票务订单

```http
POST /api/v1/order/create
```

请求：

```json
{
  "activity_id": 1,
  "ticket_spec_id": 1,
  "quantity": 1,
  "buyer_name": "罗小瑞",
  "buyer_id_card": "500101199811040817"
}
```

响应：

```json
{
  "code": 200,
  "data": {
    "success": true,
    "order_no": "T2026053114300012ab34cd"
  }
}
```

说明：

- 活动必须为 `status=3` 已上架
- 票券必须启用并且在销售时间内
- 开启实名模式时，必须传 `buyer_name` 和 `buyer_id_card`
- 开启未成年人校验时，18 岁以下不可购票

### 发起微信支付

票务订单创建成功后，直接使用现有支付接口：

```http
POST /api/v1/pay/prepay
```

请求：

```json
{
  "order_no": "T2026053114300012ab34cd"
}
```

响应：

```json
{
  "code": 200,
  "data": {
    "prepay_id": "wx201410272009395522657a690389285100",
    "appId": "wx8888888888888888",
    "timeStamp": "1414561699",
    "nonceStr": "5K8264ILTKCH16CQ2502SI8ZNMTM67VS",
    "package": "prepay_id=wx201410272009395522657a690389285100",
    "signType": "RSA",
    "paySign": "xxx",
    "out_trade_no": "T2026053114300012ab34cd"
  }
}
```

说明：

- 票务支付只需要传 `order_no`
- 后端会从 `ticket_orders` 读取金额，前端不需要传 `amount`
- 微信支付成功回调后，后端会把 `ticket_orders.status` 从 `0` 更新为 `1`
- `pay_method` 会写入微信回调里的交易类型，`pay_time` 写入回调处理时间

### 获取订单详情

```http
GET /api/v1/order/:order_no
```

响应：

```json
{
  "code": 200,
  "data": {
    "order_no": "T2026053114300012ab34cd",
    "status": 0,
    "total_price": 8800,
    "actual_price": 8800,
    "quantity": 1,
    "activity": {
      "id": 1,
      "name": "周末电音派对",
      "start_time": "2026-06-12T20:00:00+08:00",
      "end_time": "2026-06-13T02:00:00+08:00",
      "poster_list": "https://cdn.xxx/list.jpg"
    },
    "ticket_spec": {
      "name": "早鸟票"
    },
    "buyer_name": "罗小瑞",
    "buyer_id_card": "500101199811040817",
    "pay_method": "",
    "pay_time": null,
    "created_at": "2026-05-31T14:30:00+08:00",
    "qr_code": "TICKET:T2026053114300012ab34cd:xxxx",
    "expire_time": "2026-05-31T14:45:00+08:00"
  }
}
```

### 获取取消原因列表

```http
GET /api/v1/order/cancel-reasons
```

### 取消订单

```http
POST /api/v1/order/:order_no/cancel
```

请求：

```json
{
  "reason_id": 1
}
```

---

## 7. 退款模块

### 获取退款原因列表

```http
GET /api/v1/refund/reasons
```

### 申请退款

```http
POST /api/v1/refund/apply
```

请求：

```json
{
  "order_no": "T2026053114300012ab34cd",
  "reason_id": 1
}
```

响应：

```json
{
  "code": 200,
  "data": {
    "success": true,
    "refund_no": "R20260531150000abcd"
  }
}
```

### 获取退款详情

```http
GET /api/v1/refund/:refund_no
```

### 取消退款

```http
POST /api/v1/refund/:refund_no/cancel
```

---

## 8. 核销模块

### 核销员列表

```http
GET /api/v1/organizer/verifiers?page=1&size=10
```

### 添加核销员

```http
POST /api/v1/organizer/verifier
```

请求：

```json
{
  "name": "核销员A",
  "phone": "13800138000"
}
```

### 删除核销员

```http
DELETE /api/v1/organizer/verifier/:id
```

### 获取核销员激活码

```http
GET /api/v1/organizer/verifier/:id/activation-qr
```

响应：

```json
{
  "code": 200,
  "data": {
    "wechat_qr": "hyper://verifier/activate?id=1&channel=wechat",
    "douyin_qr": "hyper://verifier/activate?id=1&channel=douyin"
  }
}
```

### 核销员激活

```http
POST /api/v1/verifier/activate
```

请求：

```json
{
  "phone": "13800138000",
  "channel": "wechat"
}
```

### 扫码识别订单

```http
POST /api/v1/verifier/scan
```

请求：

```json
{
  "qr_code": "TICKET:T2026053114300012ab34cd:xxxx",
  "activity_id": 1
}
```

响应成功：

```json
{
  "code": 200,
  "data": {
    "success": true,
    "order": {
      "activity_name": "周末电音派对",
      "ticket_spec_name": "早鸟票",
      "quantity": 1,
      "buyer_name_masked": "罗**",
      "buyer_id_card_masked": "5001**********0817"
    }
  }
}
```

响应失败：

```json
{
  "code": 200,
  "data": {
    "success": false,
    "error_code": "ALREADY_VERIFIED"
  }
}
```

错误码：

| error_code | 说明 |
|---|---|
| INVALID_QR | 无效订单码 |
| ALREADY_VERIFIED | 已核销 |
| NOT_VERIFIABLE_TIME | 非可核销时间 |
| ORDER_NOT_FOUND | 订单不存在 |
| ORDER_CANCELLED | 订单已取消 |
| WRONG_ACTIVITY | 订单不属于当前活动 |

### 确认核销

```http
POST /api/v1/verifier/confirm
X-Verifier-Id: <verifier_id>
```

请求：

```json
{
  "order_no": "T2026053114300012ab34cd"
}
```

说明：当前版本先用 `X-Verifier-Id` 传核销员 ID，后续可以替换为独立核销员 token。

### 已核销列表

```http
GET /api/v1/verifier/verified-list?page=1&size=10
X-Verifier-Id: <verifier_id>
```

---

## 9. 前端推荐流程

### 活动发布

1. `POST /api/v1/organizer/apply`
2. 上传海报：`POST /api/v1/upload`
3. 分步保存：`POST /api/v1/activity/create`
4. 保存票券：`POST /api/v1/activity/:id/ticket-specs`
5. 提交审核：`POST /api/v1/activity/:id/submit-audit`

### 用户购票

1. `GET /api/v1/activity/:id`
2. `POST /api/v1/order/create`
3. `POST /api/v1/pay/prepay`，只传 `order_no` 发起微信支付
4. `GET /api/v1/order/:order_no` 查询订单和二维码

### 核销

1. 主办方添加核销员：`POST /api/v1/organizer/verifier`
2. 核销员激活：`POST /api/v1/verifier/activate`
3. 扫码识别：`POST /api/v1/verifier/scan`
4. 确认核销：`POST /api/v1/verifier/confirm`

---

## 10. 当前实现边界

- 新票务订单使用 `ticket_orders`，旧商品订单仍使用 `orders/products/pay`。
- 微信支付已支持通过 `/api/v1/pay/prepay` + `order_no` 支付 `ticket_orders`。
- 微信退款资金流暂未接入微信退款 API，目前实现的是退款申请和状态流转。
- 核销员当前临时使用 `X-Verifier-Id`，后续可升级为独立核销员 token。
