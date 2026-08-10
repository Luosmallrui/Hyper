# 2026-08-10 后端修复与前端接入说明

更新时间：2026-08-10

本文对应 `backend_fix_requests_20260810.md`，说明已修复内容、已存在契约和发布前需要执行的数据库迁移。

## 1. 已修复：主办方资料接口 500

接口：

```http
GET /api/v1/organizer/info
Authorization: Bearer <organizer_access_token>
```

根因是主办方等级信息回写时，SQL 的 `WHERE id = ?` 漏传了 `organizer_id`。生成的 `UPDATE` 语句包含四个参数，实际只传入三个，因此报：

```text
sql: expected 4 arguments, got 3
```

现已修复。接口会正常返回主办方名称、Logo、审核状态、入驻天数、当前等级、服务费比例、已完成活动数和下一等级门槛。

## 2. 已存在的登录与活动详情契约

### 密码登录 Token 续期

```http
POST /api/v1/auth/login-password
```

响应已包含：

```json
{
  "access_token": "...",
  "refresh_token": "...",
  "access_expire": 1780000000,
  "refresh_expire": 1780600000
}
```

前端收到 `401` 时，使用当前 `refresh_token` 调用：

```http
POST /api/v1/auth/refresh
Authorization: Bearer <refresh_token>
```

### 活动详情主办方用户 ID

```http
GET /api/v1/activity/:id
```

活动详情顶层已返回主办方账号 ID，同时 `organizer.user_id` 也保留：

```json
{
  "id": 15,
  "user_id": 35,
  "organizer": {
    "id": 7,
    "user_id": 35,
    "name": "Hyper Club"
  }
}
```

对象关注应优先使用本接口返回的 `follow_target_type` 和 `follow_target_id`，不能再只用 `user_id` 推导活动或场地关注目标。完整约定见 [content_follow_api_20260810.md](content_follow_api_20260810.md)。

## 3. 新增：主办方订单销售汇总

```http
GET /api/v1/organizer/orders/summary?start_date=2026-08-01&end_date=2026-08-10
Authorization: Bearer <organizer_access_token>
```

查询参数：

| 参数 | 必填 | 说明 |
|---|---|---|
| `start_date` | 否 | 支付时间起点，支持 `YYYY-MM-DD` 或完整日期时间。 |
| `end_date` | 否 | 支付时间终点；仅传日期时包含当天。 |

成功响应：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "total_amount": 32800,
    "order_count": 4,
    "ticket_count": 6,
    "average_order_amount": 8200,
    "activity_ranks": [
      {
        "activity_id": 15,
        "activity_name": "周末电音派对",
        "order_count": 3,
        "ticket_count": 5,
        "total_amount": 26400
      }
    ]
  }
}
```

字段说明：

- 金额单位均为分。
- 只统计当前状态为 `1`（待使用）或 `2`（已使用）的订单。
- 已退款、退款中、退款驳回和已取消订单不计入本接口的当前成交额。
- 日期范围按 `pay_time` 过滤；未传日期时统计该主办方全部历史成交订单。
- `activity_ranks` 是全量活动排行，不受订单列表分页大小限制。

销售数据页应改用此接口获取成交额、订单数、客单价及活动排行，不再拉取 `/organizer/orders` 的前 100 条后在客户端聚合。

## 4. 内容关注发布前 SQL

对象关注代码已支持，但生产库必须先创建 `content_follows` 表，否则地图、活动和场地接口无法计算 `follow_count`。

```sql
CREATE TABLE IF NOT EXISTS `content_follows` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `user_id` bigint unsigned NOT NULL COMMENT '关注用户ID',
  `target_type` varchar(20) NOT NULL COMMENT 'activity/venue/party',
  `target_id` bigint unsigned NOT NULL COMMENT '活动ID、场地主办方ID或旧派对ID',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_content_follow` (`user_id`, `target_type`, `target_id`),
  KEY `idx_content_follow_target` (`target_type`, `target_id`),
  KEY `idx_content_follow_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
```

发布该版本并执行 SQL 后，以下接口会返回 `follow_target_type`、`follow_target_id`、`follow_count`：

```http
GET /api/v1/map/markers
GET /api/v1/activity/:id
GET /api/v1/merchant/:id
GET /api/v1/venues
GET /api/v1/venues/:id
```

## 5. `poster_wechat` 现状

`poster_wechat` 当前会上传、保存到 `activities.poster_wechat`，并随活动详情返回；后端没有使用它生成 IM 卡片、分享海报或社群二维码。

因此客户端可以下线“活动微信社群”上传入口，不影响购票、分享或核销。历史字段会继续保留，以免影响已有活动数据。

## 6. 本轮未新增的可选数据

以下能力仍需要产品确认后再单独开发，当前客户端已隐藏相关入口：

- 活动浏览量/独立访客数，因此暂无可靠转化率。
- 订单销售渠道筛选（微信、抖音等）。
- 订单或活动维度的提现状态筛选。
