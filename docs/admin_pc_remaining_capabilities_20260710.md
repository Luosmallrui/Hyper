# 管理端 PC 剩余能力接口说明

更新时间：2026-07-10

本次补齐动态真实分享记录、平台资金流水和管理员人工调整积分。所有管理端接口均需管理员 Token：

```http
Authorization: Bearer <admin_access_token>
```

## 1. 动态真实分享记录

### 用户侧写入分享行为

```http
POST /api/v1/note/:note_id/share
Authorization: Bearer <user_access_token>
Content-Type: application/json
```

```json
{
  "channel": "wechat_session"
}
```

`channel` 支持：`wechat_session`、`wechat_timeline`、`copy_link`、`poster`、`other`。

每次成功调用都会新增一条 `note_shares` 日志，同时原子增加 `note_stats.share_count`。同一用户多次分享会保留多条记录。

### 管理端查询分享明细

```http
GET /api/v1/admin/notes/:id/records/shares?page=1&pageSize=20
```

响应：

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": 1,
        "note_id": 100,
        "user_id": 20,
        "user_name": "用户昵称",
        "user_mobile": "138****0000",
        "channel": "wechat_session",
        "created_at": "2026-07-10T12:00:00+08:00"
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 20
  }
}
```

## 2. 平台资金流水

```http
GET /api/v1/admin/finance/platform-flows?page=1&pageSize=20&type=order_income&organizer_id=10&start_date=2026-07-01&end_date=2026-07-10&keyword=T20260710
```

查询参数：

| 参数 | 说明 |
| --- | --- |
| `page` / `pageSize` | 分页，`pageSize` 最大 100 |
| `type` | `order_income`、`refund`、`service_fee`、`settlement`、`withdraw`、`manual_adjustment` |
| `keyword` | 流水号、订单号、退款单号、商家名称 |
| `organizer_id` | 商家 ID |
| `start_date` / `end_date` | `YYYY-MM-DD`，结束日期包含当天 |

响应中的金额均为分：

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": 1,
        "flow_no": "PF206...",
        "type": "service_fee",
        "direction": "income",
        "amount": 500,
        "order_no": "T202607100001",
        "refund_no": "",
        "withdraw_id": 0,
        "organizer_id": 10,
        "organizer_name": "Hyper Club",
        "pay_method": "JSAPI",
        "remark": "活动10服务费，费率5.00%",
        "occurred_at": "2026-07-10T12:00:00+08:00"
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 20
  }
}
```

写入规则：支付成功写入订单收入、服务费和商家可结算金额；退款成功写入退款；提现审核通过写入提现。流水使用业务键幂等，历史记录不更新，修正应追加冲正记录。

注意：新表上线后开始记录新业务。若管理端需要展示上线前历史订单，后端运维需单独执行一次数据回填。

## 3. 管理员人工调整积分

```http
POST /api/v1/admin/points/adjust
Content-Type: application/json
```

```json
{
  "user_id": 1001,
  "points": 500,
  "reason": "活动补偿",
  "request_no": "PA202607100001"
}
```

- `points > 0`：补发积分。
- `points < 0`：扣减积分，余额不能扣为负数。
- `request_no` 幂等；重复提交返回首次处理后的余额，不会再次变动。
- 成功后会写入积分流水及管理员操作日志。

响应：

```json
{
  "code": 200,
  "data": {
    "user_id": 1001,
    "change_points": 500,
    "balance": 1500,
    "request_no": "PA202607100001"
  }
}
```

## 部署要求

先执行 [config/table.sql](/Users/luosmallrui/Hyper/config/table.sql) 中的 `note_shares`、`platform_finance_flows` 建表语句，以及 `point_logs` 唯一索引迁移，再部署服务。

旧的 `POST /api/v1/points/reward` 已取消注册，客户端和管理端都不要再调用；后台人工加减积分统一使用 `/api/v1/admin/points/adjust`。
