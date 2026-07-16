# 管理端退款详情接口

更新时间：2026-07-15

## 使用场景

管理端 PC 的“售后订单”列表可根据退款单号直接进入退款详情页，不再需要先查询订单详情后从 `refunds` 数组中筛选。

## 接口

```http
GET /api/v1/admin/refunds/:refund_no
Authorization: Bearer <admin_access_token>
```

路径参数：

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `refund_no` | 是 | 退款单号，例如 `R2026071515300012ab34cd` |

## 权限

- 管理员必须已登录。
- 要求 RBAC 权限：`admin.orders`。
- 管理端会记录查看退款详情的操作日志，动作编码为 `admin.refund.view`。

## 成功响应

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "refund": {
      "id": 101,
      "order_id": 2001,
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
      "total_price": 17600,
      "actual_price": 17600,
      "points_amount": 100,
      "points_discount": 100,
      "quantity": 2,
      "user_id": 1001,
      "user_name": "用户昵称",
      "user_mobile": "138****8000",
      "activity_id": 10,
      "activity_name": "周末电音派对",
      "ticket_spec_id": 3,
      "ticket_spec_name": "早鸟票",
      "pay_method": "JSAPI",
      "pay_time": "2026-07-15T14:22:00+08:00",
      "created_at": "2026-07-15T14:20:00+08:00"
    },
    "viewers": [
      {
        "viewer_id": 12,
        "real_name": "张三",
        "id_card": "5101**********1234",
        "id_card_masked": "5101**********1234",
        "phone": "138****8000",
        "phone_masked": "138****8000",
        "type": 1
      }
    ],
    "refund_logs": [
      {
        "id": 1,
        "refund_id": 101,
        "status": "pending_review",
        "description": "用户提交退款申请",
        "created_at": "2026-07-15T15:30:00+08:00"
      }
    ],
    "pay_records": [],
    "verification_records": []
  }
}
```

## 字段说明

### `refund`

当前退款单本身：

| 字段 | 说明 |
| --- | --- |
| `refund_no` | 退款单号 |
| `refund_amount` | 实际退款金额，单位为分 |
| `deduct_amount` | 退款手续费或扣减金额，单位为分 |
| `reason` | 用户申请退款原因 |
| `status` | 退款状态，见下方枚举 |
| `wechat_refund_id` | 微信退款单号 |
| `wechat_status` | 微信退款状态 |
| `reject_reason` | 管理员驳回原因 |
| `expect_arrive_date` | 预计到账日期 |

退款状态：

| 值 | 说明 |
| --- | --- |
| `0` | 待审核 |
| `1` | 退款中 |
| `2` | 已退款 |
| `3` | 已驳回 |
| `4` | 已取消 |

### 关联数据

| 字段 | 说明 |
| --- | --- |
| `order` | 关联订单、买家、活动、票种和支付信息 |
| `viewers` | 订单实名观演人列表 |
| `refund_logs` | 仅当前 `refund_no` 的退款状态变更记录，按时间正序返回 |
| `pay_records` | 关联订单支付流水 |
| `verification_records` | 关联订单核销记录 |

金额字段均以分为单位，前端展示时除以 `100` 并按货币格式化。

## 失败响应

退款单不存在：

```json
{
  "code": 404,
  "msg": "退款单不存在"
}
```

无权限：

```json
{
  "code": 403,
  "msg": "无权限"
}
```

## 前端接入

1. 售后订单列表的详情按钮传递 `refund_no`。
2. 详情页请求 `GET /api/v1/admin/refunds/${refundNo}`。
3. 根据 `refund.status` 展示退款状态、审核按钮或微信退款进度。
4. `refund_logs` 用作退款处理时间线，不要混入同一订单的其他退款单日志。
5. 审核通过、驳回后刷新当前详情接口，保证退款状态和处理日志同步。

既有审核接口保持不变：

```http
POST /api/v1/admin/orders/:order_no/refund/approve
POST /api/v1/admin/orders/:order_no/refund/reject
```
