# 票务订单多人观演人接口更新

本文说明本次新增的“一个订单购买多张票，并绑定多位观演人”后端能力。

## 创建订单

```http
POST /api/v1/order/create
```

请求示例：

```json
{
  "activity_id": 1,
  "ticket_spec_id": 1,
  "quantity": 2,
  "viewer_ids": [12, 13],
  "viewers": [
    {
      "id": 12,
      "real_name": "罗小瑞",
      "id_card": "500101199811040817",
      "phone": "13800138000"
    },
    {
      "id": 13,
      "real_name": "农子健",
      "id_card": "500101199901010818",
      "phone": "13800138001"
    }
  ]
}
```

规则：

- `quantity` 可以大于 1，但不能超过票档 `purchase_limit` 和剩余库存。
- 实名活动下，后端要求最终观演人数必须等于 `quantity`。
- `viewer_ids` 读取当前账号已保存的观演人。
- `viewers` 支持直接提交实名快照；如果元素里带 `id` 或 `viewer_id`，以后端保存的观演人信息为准。
- 前端同时提交 `viewer_ids` 和同 ID 的 `viewers` 时，后端会自动去重。
- 旧字段 `buyer_name`、`buyer_id_card` 仍兼容单人购票；多人购票请使用 `viewer_ids/viewers`。

响应会返回取票码和观演人脱敏列表：

```json
{
  "code": 200,
  "data": {
    "order_no": "T2026061314300012ab34cd",
    "total_price": 17600,
    "actual_price": 17600,
    "qr_code": "TICKET:T2026061314300012ab34cd:xxxx",
    "qr_code_url": "https://cdn.hypercn.cn/ticket/qrcode/2026/06/13/T2026061314300012ab34cd.png",
    "viewers": [
      {
        "viewer_id": 12,
        "real_name": "罗小瑞",
        "id_card_masked": "5001**********0817",
        "phone_masked": "138****8000",
        "type": 1
      }
    ]
  }
}
```

## 订单详情和列表

`GET /api/v1/order/:order_no` 会返回 `viewers`，其中详情页包含完整 `id_card` 和 `phone`，同时也返回脱敏字段。

`GET /api/v1/order/list` 会返回 `viewers` 脱敏列表，方便订单卡片展示“多人实名”。

## 核销扫码

`POST /api/v1/verifier/scan` 返回的 `order.viewers` 会包含全部观演人脱敏信息，线下核验时可以确认该订单包含几位观演人。
