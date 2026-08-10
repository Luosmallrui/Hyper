# 活动流量统计与提现状态接口

更新时间：2026-08-10

Base URL：`/api/v1`

本文面向商家端和小程序端前端，说明活动 PV/UV、转化率，以及订单和活动的提现状态字段。

## 1. 发布前置条件

后端部署前必须执行本次新增的两张表：

```sql
CREATE TABLE IF NOT EXISTS `activity_daily_stats` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `activity_id` bigint unsigned NOT NULL,
  `stat_date` date NOT NULL,
  `view_count` bigint NOT NULL DEFAULT 0,
  `visitor_count` bigint NOT NULL DEFAULT 0,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_activity_daily_stat` (`activity_id`, `stat_date`),
  KEY `idx_activity_daily_stat_date` (`stat_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `activity_daily_visitors` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `activity_id` bigint unsigned NOT NULL,
  `stat_date` date NOT NULL,
  `visitor_key` varchar(80) NOT NULL,
  `user_id` bigint unsigned NOT NULL DEFAULT 0,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_activity_daily_visitor` (`activity_id`, `stat_date`, `visitor_key`),
  KEY `idx_activity_daily_visitor_date` (`stat_date`),
  KEY `idx_activity_daily_visitor_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `organizer_withdraw_allocations` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `withdraw_id` bigint unsigned NOT NULL,
  `organizer_id` bigint unsigned NOT NULL,
  `order_id` bigint unsigned NOT NULL,
  `order_no` varchar(30) NOT NULL,
  `activity_id` bigint unsigned NOT NULL,
  `amount` bigint NOT NULL,
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '0提现审核中 1已提现 2已释放',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_withdraw_order` (`withdraw_id`, `order_id`),
  KEY `idx_withdraw_allocation_organizer` (`organizer_id`),
  KEY `idx_withdraw_allocation_order` (`order_id`),
  KEY `idx_withdraw_allocation_activity` (`activity_id`),
  KEY `idx_withdraw_allocation_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
```

金额字段均为分。

## 2. 活动浏览记录

浏览记录不需要前端额外请求。成功读取公开活动详情后，后端自动记录一次 PV：

```http
GET /api/v1/activity/:id
```

### 游客 UV 标识

已登录用户由后端使用 Token 中的用户 ID 去重，无需额外字段。

游客请求活动详情时，前端应在本地长期保存一个随机 UUID，并每次放入请求头：

```http
X-Visitor-Id: 550e8400-e29b-41d4-a716-446655440000
```

规则：

- 同一活动、同一访客、同一天只计一次 UV。
- 每次成功打开详情都计一次 PV。
- 后端只保存游客 ID 的 SHA-256 哈希，不保存原始访客标识。
- 未携带 `X-Visitor-Id` 的游客仍计 PV，但不能计入 UV。
- 请勿在每次请求时重新生成 UUID，否则会虚增 UV。

## 3. 单活动统计

```http
GET /api/v1/activity/:id/statistics
Authorization: Bearer <organizer_access_token>
```

仅当前活动所属主办方可查询。

新增/补齐字段：

```json
{
  "code": 200,
  "data": {
    "ticket_count": 30,
    "buyer_count": 20,
    "gross_amount": 300000,
    "refund_amount": 10000,
    "net_amount": 290000,

    "view_count": 1280,
    "visitor_count": 860,
    "paid_order_count": 24,
    "conversion_rate": 0.027907,

    "available_withdraw_amount": 120000,
    "pending_withdraw_amount": 50000,
    "withdrawn_amount": 120000
  }
}
```

| 字段 | 说明 |
|---|---|
| `view_count` | 该活动累计 PV。 |
| `visitor_count` | 该活动累计 UV。 |
| `paid_order_count` | 已成功支付订单数，包含后续进入退款流程的订单。 |
| `conversion_rate` | `paid_order_count / visitor_count`，小数形式；例如 `0.027907` 表示约 `2.79%`。 |
| `available_withdraw_amount` | 按活动归集后仍可提现的金额。 |
| `pending_withdraw_amount` | 已提交提现、尚待平台审核的订单金额。 |
| `withdrawn_amount` | 平台已审核通过的提现金额。 |

### 每日趋势

```http
GET /api/v1/activity/:id/statistics/daily?days=7
Authorization: Bearer <organizer_access_token>
```

原有订单趋势项保留，新增：

```json
{
  "date": "2026-08-10",
  "amount": 32800,
  "ticket_count": 6,
  "order_count": 4,
  "view_count": 210,
  "visitor_count": 156
}
```

## 4. 商家销售汇总

```http
GET /api/v1/organizer/orders/summary?start_date=2026-08-01&end_date=2026-08-10
Authorization: Bearer <organizer_access_token>
```

新增汇总字段：

```json
{
  "total_amount": 32800,
  "order_count": 4,
  "ticket_count": 6,
  "average_order_amount": 8200,

  "view_count": 1280,
  "visitor_count": 860,
  "paid_order_count": 24,
  "conversion_rate": 0.027907,

  "activity_ranks": [
    {
      "activity_id": 15,
      "activity_name": "周末电音派对",
      "order_count": 3,
      "ticket_count": 5,
      "total_amount": 26400,
      "view_count": 900,
      "visitor_count": 600,
      "paid_order_count": 16,
      "conversion_rate": 0.026667,
      "available_withdraw_amount": 13000,
      "pending_withdraw_amount": 5000,
      "withdrawn_amount": 8400
    }
  ]
}
```

说明：

- `total_amount`、`order_count`、活动排行的原有成交金额只统计当前订单状态为待使用或已使用的订单。
- `paid_order_count` 用于转化率，统计曾成功支付的订单，退款中的订单也会保留在该指标中。
- `start_date`、`end_date` 同时影响支付订单和按日流量统计；日期范围按自然日统计。

## 5. 订单提现状态与筛选

订单列表沿用原接口：

```http
GET /api/v1/organizer/orders?page=1&size=20&withdraw_status=available
Authorization: Bearer <organizer_access_token>
```

新增查询参数：

| 参数 | 值 | 说明 |
|---|---|---|
| `withdraw_status` | `available` | 可提现，尚未被任何有效提现申请锁定。 |
| `withdraw_status` | `pending_withdraw` | 已分配到待审核提现申请。 |
| `withdraw_status` | `withdrawn` | 已分配到审核通过的提现申请。 |

不传 `withdraw_status` 保持原订单列表行为。

订单列表与 `GET /api/v1/organizer/orders/:order_no` 都会返回：

```json
{
  "withdraw_status": "pending_withdraw",
  "withdraw_amount": 5000
}
```

- `withdraw_amount` 是该订单已被待审核或已审核通过提现单分配的金额，单位为分。
- 订单实际金额可能大于 `withdraw_amount`，表示该订单还有剩余可提现余额。
- `unavailable` 是返回态，不支持作为筛选值，表示待支付、已取消、退款中或已退款等不可提现订单。

## 6. 提现状态流转

商家申请提现仍使用原接口：

```http
POST /api/v1/organizer/withdraws
```

后端会在同一事务中：

1. 校验商家可提现余额和收款账户。
2. 创建提现申请。
3. 按订单支付时间 FIFO 分配订单金额至该提现单。
4. 将这些订单即时标记为 `pending_withdraw`。

管理端审核提现：

- 审核通过：相关分配项变更为 `withdrawn`。
- 审核拒绝：相关分配项释放，订单重新显示为 `available`。

### 历史数据说明

本功能上线前创建的提现申请没有订单分配明细，无法安全倒推到具体订单或活动。新的订单级/活动级提现状态从本次版本上线后的提现申请开始准确追踪；历史提现仍会保留在原有商家资金汇总中。

## 7. 前端接入顺序

1. 在小程序本地初始化并持久化 `visitor_id`，活动详情请求带 `X-Visitor-Id`。
2. 销售数据页读取 `/organizer/orders/summary` 的流量与转化字段。
3. 活动统计页读取 `/activity/:id/statistics` 和 `/activity/:id/statistics/daily` 的新增字段。
4. 商家订单页增加提现状态筛选，并显示 `withdraw_status`、`withdraw_amount`。
5. 前端将 `conversion_rate` 乘以 `100` 后格式化为百分比显示。
