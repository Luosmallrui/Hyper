# 小程序前端接口更新说明（2026-08-10）

Base URL：

```text
/api/v1
```

除公开地图、活动和场地查询外，接口均携带：

```http
Authorization: Bearer <access_token>
Content-Type: application/json
```

金额字段单位统一为分。

## 1. 主办方账户信息

```http
GET /api/v1/organizer/info
```

用于商家端“账户/品牌卡”。接口已修复，可直接恢复调用；不再需要用 `/organizer/profile` 兜底等级与服务费信息。

响应示例：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "id": 7,
    "name": "Hyper Club",
    "logo": "https://cdn.example.com/logo.png",
    "status": 2,
    "reject_reason": "",
    "join_days": 31,
    "level": "LV2",
    "level_value": 2,
    "service_fee_rate": 0.03,
    "fee_rate": 0.03,
    "completed_activity_count": 6,
    "next_level_required_count": 10,
    "basic_info": {
      "province": "四川省",
      "city": "成都市",
      "district": "武侯区"
    }
  }
}
```

字段说明：

| 字段 | 说明 |
|---|---|
| `status` | `2` 为审核通过。 |
| `level` / `level_value` | 当前主办方等级。 |
| `service_fee_rate` / `fee_rate` | 手续费比例，小数表示，例如 `0.03` 是 3%。 |
| `completed_activity_count` | 已结束且已上架的活动数量。 |
| `next_level_required_count` | 下一等级所需活动数；最高等级时为 `0`。 |

## 2. 销售数据汇总

```http
GET /api/v1/organizer/orders/summary
```

可选查询参数：

| 参数 | 示例 | 说明 |
|---|---|---|
| `start_date` | `2026-08-01` | 支付时间起点；也支持完整日期时间。 |
| `end_date` | `2026-08-10` | 支付时间终点；仅传日期时包含当天。 |

示例：

```http
GET /api/v1/organizer/orders/summary?start_date=2026-08-01&end_date=2026-08-10
```

响应：

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

接入规则：

- 销售数据页的成交额、成交单数、客单价、活动排行全部改用本接口。
- 不再通过 `GET /organizer/orders?size=100` 拉取订单后在客户端聚合。
- 只统计当前订单状态为 `1`（待使用）和 `2`（已使用）的订单。
- 已退款、退款中、退款驳回、已取消订单不计入 `total_amount`。

## 3. 密码登录与自动续期

```http
POST /api/v1/auth/login-password
```

请求：

```json
{
  "phone": "13800000000",
  "password": "******"
}
```

响应始终包含：

```json
{
  "access_token": "...",
  "refresh_token": "...",
  "access_expire": 1780000000,
  "refresh_expire": 1780600000
}
```

收到 `401` 后，使用保存的 `refresh_token` 调用：

```http
POST /api/v1/auth/refresh
Authorization: Bearer <refresh_token>
```

刷新成功后替换本地的 `access_token` 和返回的 `refresh_token`；仅当刷新失败时清除登录态。

## 4. 活动、场地、派对对象关注

> 此项随本次后端代码一起部署。生产数据库必须先创建 `content_follows` 表；未部署前，客户端应维持原有用户关注兼容逻辑。

列表和详情响应会增加：

```json
{
  "is_follow": true,
  "follow_count": 28,
  "follow_target_type": "activity",
  "follow_target_id": 15
}
```

适用接口：

```http
GET /api/v1/map/markers
GET /api/v1/activity/:id
GET /api/v1/merchant/:id
GET /api/v1/venues
GET /api/v1/venues/:id
```

内容关注/取消关注沿用已有路径：

```http
POST /api/v1/follow/follow
POST /api/v1/follow/unfollow
```

请求体：

```json
{
  "target_type": "activity",
  "target_id": 15
}
```

`target_type` 取值：

| 值 | `target_id` |
|---|---|
| `activity` | 新派对/活动的 `activities.id`。 |
| `venue` | 场地主办方的 `organizers.id`。 |
| `party` | 旧派对的 `parties.id`。 |

接入规则：

- 必须直接使用接口返回的 `follow_target_type` 和 `follow_target_id` 发起关注，不要依据 `user_id`、`source_id` 或活动类型自行推导。
- 内容卡片的粉丝数显示 `follow_count`，不要展示用户资料里的 `stats.follower`。
- 用户主页关注仍按旧逻辑提交 `user_id`，不传 `target_type` 和 `target_id`。
- 完整对象关注说明见 [content_follow_api_20260810.md](content_follow_api_20260810.md)。

## 5. 活动详情主办方用户

```http
GET /api/v1/activity/:id
```

响应顶层和 `organizer` 内都提供主办方账号 ID：

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

- 展示主办方用户资料、跳转用户主页时使用 `user_id`。
- 关注活动/场地时不要使用 `user_id`，使用上一节的 `follow_target_*` 字段。

## 6. 活动时间与票券开售时间

活动 `start_time`、`end_time` 和票券 `sale_start`、`sale_end` 都支持精确到秒。

推荐提交格式：

```json
{
  "start_time": "2026-08-10T19:30:00+08:00",
  "end_time": "2026-08-10T23:30:00+08:00",
  "ticket_specs": [
    {
      "sale_start": "2026-08-01T10:00:00+08:00",
      "sale_end": "2026-08-10T19:25:00+08:00"
    }
  ]
}
```

也兼容 `YYYY-MM-DDTHH:mm`、`YYYY-MM-DD HH:mm`、`YYYY-MM-DD HH:mm:ss`。

不能只传日期，例如 `2026-08-10`；前端日期选择和时间选择后应拼成完整日期时间再提交。

## 7. `poster_wechat` 处理

`poster_wechat` 可以继续读取历史活动数据，但后端当前不会用它生成社群二维码、分享海报或消息卡片。

前端可以移除活动发布页中的“活动微信社群”图片上传入口；不影响活动发布、购票、核销和分享。

## 8. 当前不接入的筛选/指标

以下数据源尚未提供，相关前端入口保持隐藏：

- 活动浏览量、独立访客数和转化率。
- 销售渠道筛选（微信/抖音等）。
- 订单或活动维度的提现状态筛选。
