# 管理端 PC 0709 后端接口更新说明

更新时间：2026-07-09

## 1. 更新范围

本次针对管理端 PC 0709 优化需求，补齐后端已能直接支持的接口能力：

- 订单列表筛选与订单详情字段增强
- 售后状态筛选
- 用户参加/订阅记录切换到新票务活动数据源
- 消息列表筛选与发布校验
- 商家对账字段增强
- 距离分类配置后端校验

未做假数据兜底：动态分享记录仍缺少真实分享日志表，详见文末“仍需后端后续补齐”。

## 2. 订单与售后

### 订单列表

```http
GET /api/v1/admin/orders?page=1&pageSize=20&keyword=&status=&refund_status=&activity_id=
Authorization: Bearer <admin_access_token>
```

Query:

| 参数 | 必填 | 说明 |
|---|---|---|
| page | 否 | 默认 `1` |
| pageSize | 否 | 默认 `20`，最大 `100` |
| keyword | 否 | 支持订单号、买家姓名、身份证、用户昵称、手机号、活动名、票种名 |
| status | 否 | 订单主状态 |
| refund_status | 否 | 售后状态，支持数字或字符串 |
| activity_id | 否 | 活动 ID |

`refund_status` 字符串枚举：

| 值 | 说明 |
|---|---|
| pending_review | 待审核 |
| refunding | 退款中 |
| refunded | 已退款 |
| rejected | 已驳回 |
| cancelled | 已取消 |

响应列表项新增/确认字段：

```json
{
  "order_no": "T202607090001",
  "status": 1,
  "total_price": 10000,
  "actual_price": 9000,
  "points_amount": 100,
  "points_discount": 1000,
  "quantity": 1,
  "user_id": 2,
  "user_name": "用户昵称",
  "user_mobile": "13800000000",
  "buyer_name": "张三",
  "buyer_id_card": "510xxxxxxxxxxxxxxx",
  "activity_id": 10,
  "activity_name": "周末电音派对",
  "ticket_spec_id": 3,
  "ticket_spec_name": "早鸟票",
  "pay_method": "wechat",
  "pay_time": "2026-07-09 12:00:00",
  "expire_time": "2026-07-09 12:15:00",
  "created_at": "2026-07-09 12:00:00"
}
```

### 订单详情

```http
GET /api/v1/admin/orders/:order_no
Authorization: Bearer <admin_access_token>
```

响应新增字段：

```json
{
  "order": {},
  "viewers": [
    {
      "viewer_id": 1,
      "real_name": "张三",
      "id_card": "510xxxxxxxxxxxxxxx",
      "id_card_masked": "510***********1234",
      "phone": "13800000000",
      "phone_masked": "138****0000",
      "type": 1
    }
  ],
  "refunds": [],
  "refund_logs": [],
  "verification_records": [
    {
      "id": 1,
      "order_id": 20,
      "verifier_id": 5,
      "verifier_name": "核销员",
      "verifier_phone": "13800000000",
      "activity_id": 10,
      "activity_name": "周末电音派对",
      "verified_at": "2026-07-09T12:30:00+08:00"
    }
  ],
  "pay_records": []
}
```

说明：

- `viewers` 用于管理端展示实名观演人。
- `verification_records` 用于订单详情展示核销信息。
- `pay_records` 用于展示微信支付流水；积分全额抵扣或 0 元订单可能没有微信支付流水。

## 3. 用户行为记录

### 参加活动记录

```http
GET /api/v1/admin/users/:id/records/attends?page=1&pageSize=20
```

数据源已从旧 `party_attends` 切换为新票务订单 `ticket_orders`。

仅返回已支付/已使用/售后相关的票务订单：

- 待使用
- 已使用
- 退款中
- 退款成功
- 退款拒绝

返回字段包含：

```json
{
  "id": 20,
  "order_no": "T202607090001",
  "status": 1,
  "quantity": 1,
  "actual_price": 9000,
  "activity_id": 10,
  "activity_name": "周末电音派对",
  "poster_list": "https://cdn.xxx/poster.png",
  "ticket_spec_name": "早鸟票",
  "created_at": "2026-07-09T12:00:00+08:00"
}
```

### 订阅活动记录

```http
GET /api/v1/admin/users/:id/records/subscribes?page=1&pageSize=20
```

数据源已从旧 `subscribes` 切换为新活动订阅 `activity_subscriptions`。

返回字段包含：

```json
{
  "id": 1,
  "activity_id": 10,
  "activity_name": "周末电音派对",
  "poster_list": "https://cdn.xxx/poster.png",
  "activity_status": 3,
  "created_at": "2026-07-09T12:00:00+08:00"
}
```

## 4. 消息管理

### 消息列表

```http
GET /api/v1/admin/messages?page=1&pageSize=20&target=&type=
Authorization: Bearer <admin_access_token>
```

Query:

| 参数 | 必填 | 说明 |
|---|---|---|
| target | 否 | `all` / `user` / `merchant` / `organizer` |
| type | 否 | 消息类型，例如 `system` / `notice` / `activity` / `refund` |

### 发布消息

```http
POST /api/v1/admin/messages
Authorization: Bearer <admin_access_token>
Content-Type: application/json
```

请求：

```json
{
  "title": "系统通知",
  "content": "消息内容",
  "type": "notice",
  "target": "organizer",
  "organizer_ids": [1, 2],
  "target_user_ids": [],
  "status": 1
}
```

后端校验：

- `title` 必填，且不能为空字符串。
- `content` 必填，且不能为空字符串。
- `type` 不传默认 `system`。
- `target` 不传默认 `all`。

## 5. 商家对账

```http
GET /api/v1/admin/finance/settlements?page=1&pageSize=20&organizer_id=
Authorization: Bearer <admin_access_token>
```

Query:

| 参数 | 必填 | 说明 |
|---|---|---|
| organizer_id | 否 | 商家/主办方 ID；不传返回全部 |

返回字段：

```json
{
  "organizer_id": 1,
  "organizer_name": "Hyper Club",
  "gross_amount": 100000,
  "refund_amount": 10000,
  "withdraw_amount": 20000,
  "pending_withdraw_amount": 5000,
  "settle_amount": 90000,
  "pending_settle_amount": 65000,
  "order_count": 12,
  "updated_at": "2026-07-09 12:00:00"
}
```

字段说明：

| 字段 | 说明 |
|---|---|
| gross_amount | 订单收入，单位分 |
| refund_amount | 已退款金额，单位分 |
| withdraw_amount | 已审核通过提现金额，单位分 |
| pending_withdraw_amount | 待审核提现金额，单位分 |
| settle_amount | 订单收入 - 已退款金额 |
| pending_settle_amount | 订单收入 - 已退款金额 - 已提现 - 待审核提现 |
| order_count | 订单数 |
| updated_at | 最近更新时间 |

## 6. 分类管理：距离筛选校验

```http
POST /api/v1/admin/categories
PUT /api/v1/admin/categories/:id
```

当 `type=distance` 时：

- `value` 必填。
- `value` 必须是大于 0 的数字。
- 示例合法值：`1`、`3`、`5`、`10`。

错误示例：

```json
{
  "code": 500,
  "msg": "距离筛选项只能填写大于0的数字公里值"
}
```

## 7. 商家等级规则

该接口后端已存在，不再属于后端阻塞项。

```http
GET    /api/v1/admin/organizer-level-rules
POST   /api/v1/admin/organizer-level-rules
PUT    /api/v1/admin/organizer-level-rules/:id
DELETE /api/v1/admin/organizer-level-rules/:id
```

请求字段：

```json
{
  "level": 1,
  "name": "LV1",
  "fee_rate": 0.05,
  "required_activity_count": 0,
  "description": "默认等级",
  "benefits": "服务费 5%",
  "status": 1
}
```

## 8. 仍需后续补齐

### 动态分享记录

当前 `/api/v1/admin/notes/:id/records/shares` 仍只能基于 `note_stats` 返回聚合数据，不是真实分享日志。

如果管理端需要展示“哪个用户在什么时间分享了哪条动态”，后端还需要新增分享日志采集，例如：

```sql
CREATE TABLE note_shares (
  id bigint unsigned NOT NULL AUTO_INCREMENT,
  note_id bigint unsigned NOT NULL,
  user_id bigint unsigned NOT NULL,
  channel varchar(50) NOT NULL DEFAULT '',
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_note_id (note_id),
  KEY idx_user_id (user_id)
);
```

并将接口改为查询真实 `note_shares`。

### 平台流水明细

当前已有：

```http
GET /api/v1/admin/finance/summary
GET /api/v1/admin/finance/settlements
GET /api/v1/admin/withdraws
```

但还没有独立平台流水明细接口。如果前端需要“平台流水列表”，建议后续新增：

```http
GET /api/v1/admin/finance/platform-flows?page=1&pageSize=20&type=&start_date=&end_date=
```
