# 主办方退款详情接口

更新时间：2026-07-16

## 接口

```http
GET /api/v1/organizer/refunds/:refund_no
Authorization: Bearer <organizer_access_token>
```

该接口供主办方 PC 的退款详情页使用。不要使用用户侧 `GET /api/v1/refund/:refund_no`，后者只允许购票用户查询自己的退款单。

## 数据范围与权限

- 仅已登录主办方可调用。
- 仅返回当前主办方名下活动对应的退款单。
- 不属于当前主办方或退款单不存在时，统一返回 `404`：`退款单不存在或不属于当前主办方`。
- 查询成功会写入主办方操作日志，动作编码为 `organizer.refund.view`。

## 成功响应

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "refund": {
      "refund_no": "R2026071515300012ab34cd",
      "refund_amount": 17600,
      "deduct_amount": 0,
      "reason": "临时有事无法参加",
      "status": 0,
      "wechat_refund_id": "",
      "wechat_status": "",
      "reject_reason": "",
      "expect_arrive_date": "2026-07-16",
      "created_at": "2026-07-15T15:30:00+08:00",
      "updated_at": "2026-07-15T15:30:00+08:00"
    },
    "order": {
      "order_no": "T2026071514200012ab34cd",
      "status": 4,
      "actual_price": 17600,
      "quantity": 2,
      "user_name": "用户昵称",
      "user_mobile": "138****8000",
      "activity_name": "周末电音派对",
      "ticket_spec_name": "早鸟票",
      "pay_method": "JSAPI",
      "pay_time": "2026-07-15T14:22:00+08:00"
    },
    "viewers": [
      {
        "viewer_id": 12,
        "real_name": "张三",
        "id_card_masked": "5101**********1234",
        "phone_masked": "138****8000",
        "type": 1
      }
    ],
    "refund_logs": [],
    "pay_records": [],
    "verification_records": []
  }
}
```

## 字段说明

| 字段 | 说明 |
| --- | --- |
| `refund` | 当前退款单与微信退款进度。金额单位均为分。 |
| `order` | 关联订单、下单账号、活动和票种。`user_mobile` 已脱敏。 |
| `viewers` | 订单全部观演人，仅包含脱敏证件号与手机号。 |
| `refund_logs` | 仅当前退款单的审核和退款状态时间线，按时间正序。 |
| `pay_records` | 支付流水展示字段，不返回 OpenID 或微信原始回调数据。 |
| `verification_records` | 订单核销记录；核销员手机号已脱敏。 |

退款状态：

| 值 | 说明 |
| --- | --- |
| `0` | 待审核 |
| `1` | 退款中 |
| `2` | 已退款 |
| `3` | 已驳回 |
| `4` | 已取消 |

## 前端接入

1. 从 `GET /api/v1/organizer/refunds` 的列表项读取 `refund_no`。
2. 详情页请求 `GET /api/v1/organizer/refunds/${refundNo}`。
3. 使用 `refund_logs` 展示退款处理时间线，使用 `refund.wechat_status` 展示微信退款进度。
4. 审核通过、拒绝或取消退款后重新请求本接口刷新详情。
