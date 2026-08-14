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

### 入驻兼容字段 `organizer.type`

> 2026-07-09 更新：入驻申请不再区分场地/派对。`organizer.type` 仅保留兼容旧后台展示，不再作为 C 端场地/派对判断依据。

| 值 | 说明 |
|---|---|
| merchant | 默认兼容值 |
| venue | 历史兼容值 |

### 活动类型 `activity.type`

| 值 | 说明 |
|---|---|
| party | 派对，默认值 |
| venue | 场地 |

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

## 2.1 首页地图点位

### 获取首页地图点位

该接口用于小程序首页地图。当前已迁移为只返回新票务活动点位，不再混入旧 `parties` 派对/场地数据。

```http
GET /api/v1/map/markers?source=all&limit=200
```

Query:

| 字段 | 必填 | 说明 |
|---|---|---|
| source | 否 | `all` / `activity` / `party` / `venue` / `merchant`，默认 `all`；`party`、`venue` 按 `activities.type` 过滤，`merchant` 暂保留但返回空列表 |
| limit | 否 | 最多返回数量，默认 `200`，最大 `500` |
| keyword | 否 | 关键词，匹配标题、地址、描述 |
| category_id | 否 | 兼容分类筛选：`1` 映射场地 `venue`，`2` 映射派对 `party`；也可直接使用 `source` 或 `type` |
| district_id | 否 | 行政区 ID，会转为行政区名称后按 `activities.district` 匹配 |
| district | 否 | 行政区名称或 ID；如 `武侯区`，票务活动按 `activities.district` 匹配 |
| area_id | 否 | 预留字段 |
| area | 否 | 商圈/区域名称或 ID，活动按地址模糊匹配 |
| business_area | 否 | 商圈名称，按地址/位置模糊匹配 |
| tags/tag_ids | 否 | 优惠标签硬过滤：传 `GET /api/v1/tags` 返回的动态标签 ID；多值逗号分隔且需同时满足 |
| lat/lng | 否 | 用户经纬度，配合 `distance` 使用 |
| distance | 否 | 距离，单位 km；传 `lat/lng` 时生效 |

示例：

```http
GET /api/v1/map/markers?source=all&limit=200&category_id=1&district=武侯区
```

注意：新票务活动使用 `activities.type` 区分场地与派对；`category_id=1/2` 是为现有分类选择器保留的兼容映射。

数据来源：

| source | 来源表 | 条件 | 说明 |
|---|---|---|---|
| all | `activities` | `status=3` | 新票务活动，管理员审核通过后可见 |
| activity | `activities` | `status=3` | 新票务活动，管理员审核通过后可见 |
| party | `activities` | `status=3 AND type='party'` | 派对型活动 |
| venue | `activities` | `status=3 AND type='venue'` | 场地型活动 |
| merchant | - | - | 老商家/派对数据已停止作为首页数据源，返回空列表 |

响应：

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": "activity-1",
        "source": "activity",
        "source_id": 1,
		"activity_id": 1,
        "detail_type": "activity",
        "detail_url": "/api/v1/activity/1",
        "user_id": 1,
        "user": "Hyper Club",
        "username": "Hyper Club",
        "user_avatar": "https://cdn.xxx/logo.jpg",
        "userAvatar": "https://cdn.xxx/logo.jpg",
        "title": "周末电音派对",
        "type": "activity",
        "location": "朝阳区 xx 路",
        "address": "朝阳区 xx 路",
        "lat": 39.9333,
        "lng": 116.4533,
        "cover_image": "https://cdn.xxx/list.jpg",
        "created_at": "2026-06-03 12:00:00",
        "avg_price": 8800,
        "current_count": 0,
        "post_count": 0,
        "icon": "https://cdn.hypercn.cn/icon/party.png",
        "is_subscribe": false,
        "is_follow": false,
        "start_time": "2026-06-12 20:00:00",
        "end_time": "2026-06-13 02:00:00",
        "status": 3
      }
    ],
    "total": 1
  }
}
```

前端跳转建议：

优先使用 `detail_url`；当前地图点位只会跳转到新票务活动详情：

```http
GET /api/v1/activity/:id
```

说明：为兼容旧首页地图卡片，`activity` 类型也会返回 `user`、`username`、`user_avatar`、`userAvatar`、`avg_price`、`current_count`、`post_count`、`is_subscribe`、`is_follow` 等字段。`avg_price` 使用该活动票种最低价，单位为分。

### 订阅/取消订阅活动

> 新活动不要再调用旧 `/api/v1/merchant/subscribe`。正式接口如下：

`type=venue` 的场地活动同样可以直接传活动 ID 调用；后端会自动写入场地订阅关系。完整约定见 [activity_subscribe_venue_compat_api_20260814.md](activity_subscribe_venue_compat_api_20260814.md)。

```http
POST /api/v1/activity/{id}/subscribe
Authorization: Bearer <token>
```

响应：

```json
{
  "code": 200,
  "data": {
    "success": true
  }
}
```

取消订阅：

```http
POST /api/v1/activity/{id}/unsubscribe
Authorization: Bearer <token>
```

获取当前用户已订阅活动：

```http
GET /api/v1/activity/subscriptions?page=1&pageSize=20
Authorization: Bearer <token>
```

该接口同时返回已订阅的派对和场地：派对来自 `activity_subscriptions`，场地来自
`venue_subscriptions`。场地项的 `type` 为 `venue`，前端据此展示营业时间样式，点击详情仍使用活动 ID。

响应：

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": 10,
        "name": "jjjj",
        "poster_list": "https://cdn.hypercn.cn/ticketing/poster_list/xxx.png",
        "start_time": "2026-06-13T16:51:00+08:00",
        "end_time": "2026-06-19T16:51:00+08:00",
        "status": 3,
        "is_subscribe": true
      }
    ],
    "total": 1
  }
}
```

活动详情会返回当前用户订阅状态：

```http
GET /api/v1/activity/{id}
Authorization: Bearer <token>
```

```json
{
  "code": 200,
  "data": {
    "id": 10,
    "name": "jjjj",
    "is_subscribe": true
  }
}
```

兼容说明：

- 旧 `/api/v1/merchant/subscribe` 若收到的 `party_id` 实际存在于新 `activities.id`，后端会兼容写入新活动订阅。
- 前端仍应尽快切到 `/api/v1/activity/{id}/subscribe`。

数据库：

```sql
CREATE TABLE IF NOT EXISTS `activity_subscriptions`
(
    `id`          bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '订阅ID',
    `activity_id` bigint unsigned NOT NULL COMMENT '活动ID',
    `user_id`     bigint unsigned NOT NULL COMMENT '用户ID',
    `created_at`  datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_activity_user` (`activity_id`, `user_id`) USING BTREE,
    KEY `idx_activity_subscription_activity` (`activity_id`) USING BTREE,
    KEY `idx_activity_subscription_user` (`user_id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='活动订阅表';
```

## 2.2 新场地 C 端展示接口

新场地不再使用旧 `/api/v1/merchant/*`。C 端展示数据来自：

- `organizers`：只返回 `status=2`、`enabled=1` 且名下存在已上架 `activities.type=venue` 的主办方。
- `organizer_profiles`：介绍、图册、地址、营业时间、定位等资料。
- `organizer_stores`：门店/场地位置。
- `notes.store_id`：场地相关动态。
- `content_follows`：关注场地对象。
- `venue_subscriptions`：订阅场地。

### 场地列表

```http
GET /api/v1/venues?page=1&size=10&keyword=酒吧
Authorization: Bearer <access_token>
```

响应：

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": 10,
        "user_id": 2,
        "name": "SWING 鸡尾酒吧",
        "logo": "https://cdn.xxx/logo.jpg",
        "cover_image": "https://cdn.xxx/cover.jpg",
        "description": "场地介绍",
        "business_hours": "19:30-02:30",
        "service_phone": "13800000000",
        "province": "四川省",
        "city": "成都市",
        "district": "武侯区",
        "address": "天府三街",
        "latitude": 30.657,
        "longitude": 104.066,
        "average_spend": 7600,
        "is_follow": false,
        "is_subscribe": false,
        "follow_count": 12,
        "subscribe_count": 8,
        "post_count": 3,
        "created_at": "2026-06-28T12:00:00+08:00"
      }
    ],
    "total": 1
  }
}
```

### 场地详情

```http
GET /api/v1/venues/:id
Authorization: Bearer <access_token>
```

详情在列表字段基础上额外返回：

```json
{
  "gallery": ["https://cdn.xxx/1.jpg"],
  "stores": [
    {
      "id": 1,
      "organizer_id": 10,
      "name": "主店",
      "logo": "https://cdn.xxx/store.jpg",
      "address": "天府三街",
      "latitude": 30.657,
      "longitude": 104.066,
      "phone": "13800000000"
    }
  ]
}
```

### 场地相关动态

```http
GET /api/v1/venues/:id/notes?cursor=0&pageSize=20
Authorization: Bearer <access_token>
```

说明：

- 后端会查询该场地下所有 `organizer_stores.id`，再按 `notes.store_id` 返回公开动态。
- 只返回 `status != -1` 且 `visible_conf = 1` 的动态。
- `cursor` 使用上一页最后一条 `created_at` 的纳秒时间戳，即响应里的 `next_cursor`。

### 关注/取消关注场地

关注的是场地对象，落表 `content_follows`。完整对象关注参数和响应字段见 [content_follow_api_20260810.md](content_follow_api_20260810.md)。

```http
POST /api/v1/venues/:id/follow
DELETE /api/v1/venues/:id/follow
Authorization: Bearer <access_token>
```

响应：

```json
{
  "code": 200,
  "data": {
    "is_follow": true
  }
}
```

### 订阅/取消订阅场地

```http
POST /api/v1/venues/:id/subscribe
DELETE /api/v1/venues/:id/subscribe
Authorization: Bearer <access_token>
```

响应：

```json
{
  "code": 200,
  "data": {
    "is_subscribe": true
  }
}
```

### 我的订阅聚合列表

只聚合新活动和新场地，不返回旧派对/旧场地 `party_likes` 数据。

```http
GET /api/v1/subscriptions?page=1&size=20&type=all
Authorization: Bearer <access_token>
```

Query:

| 字段 | 必填 | 说明 |
|---|---|---|
| page | 否 | 默认 1 |
| size | 否 | 默认 10 |
| type | 否 | `all` / `activity` / `venue`，默认 `all` |

响应：

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": "activity-10",
        "source": "activity",
        "source_id": 10,
        "title": "周末电音派对",
        "cover_image": "https://cdn.xxx/poster.png",
        "description": "活动介绍",
        "start_time": "2026-06-28T20:00:00+08:00",
        "end_time": "2026-06-28T23:00:00+08:00",
        "status": 3,
        "address": "天府三街",
        "latitude": 30.657,
        "longitude": 104.066,
        "subscribed_at": "2026-06-28T12:00:00+08:00"
      },
      {
        "id": "venue-3",
        "source": "venue",
        "source_id": 3,
        "title": "SWING 鸡尾酒吧",
        "cover_image": "https://cdn.xxx/cover.jpg",
        "description": "场地介绍",
        "address": "天府三街",
        "latitude": 30.657,
        "longitude": 104.066,
        "extra": {
          "business_hours": "19:30-02:30",
          "service_phone": "13800000000",
          "province": "四川省",
          "city": "成都市",
          "district": "武侯区",
          "average_spend": 7600
        },
        "subscribed_at": "2026-06-28T11:00:00+08:00"
      }
    ],
    "total": 2
  }
}
```

数据库：

```sql
CREATE TABLE IF NOT EXISTS `venue_subscriptions`
(
    `id`           bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '场地订阅ID',
    `organizer_id` bigint unsigned NOT NULL COMMENT '场地主办方ID',
    `user_id`      bigint unsigned NOT NULL COMMENT '用户ID',
    `created_at`   datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_venue_user` (`organizer_id`, `user_id`) USING BTREE,
    KEY `idx_venue_subscription_organizer` (`organizer_id`) USING BTREE,
    KEY `idx_venue_subscription_user` (`user_id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='场地订阅表';
```

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
    "type": "venue",
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

字段说明：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| type | string | 否 | 兼容旧字段；入驻不再区分场地/派对，后端忽略该字段并默认按主办方入驻处理 |
| name | string | 是 | 入驻名称 |
| logo | string | 否 | Logo URL，可先用 `/api/v1/upload` 上传 |
| province | string | 否 | 省份 |
| city | string | 否 | 城市 |
| district | string | 否 | 区县 |

响应：

```json
{
  "code": 200,
  "data": {
    "application_id": 123,
    "status": 1,
    "submitted_at": "2026-06-20T13:21:00+08:00"
  }
}
```

重复提交：

```json
{
  "code": 409,
  "msg": "入驻申请正在审核中，请勿重复提交"
}
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
    "type": "venue",
    "status": 1,
    "reject_reason": "",
    "submitted_at": "2026-06-20T13:21:00+08:00",
    "reviewed_at": null
  }
}
```

未提交时：

```json
{
  "code": 200,
  "data": {
    "status": 0,
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
Authorization: Bearer <organizer_token>
```

响应：

```json
{
  "code": 200,
  "data": {
    "bank_account_name": "张三",
    "bank_account_no": "6222...",
    "bank_name": "招商银行",
    "can_withdraw": true,
    "gross_amount": 50000,
    "refund_amount": 10000,
    "withdraw_amount": 20000,
    "pending_withdraw_amount": 5000,
    "available_amount": 15000,
    "arrival_cycle": "T+1 到 T+3 个工作日",
    "pending_audit": null,
    "latest_audit": {
      "id": 12,
      "bank_account_name": "张三",
      "bank_account_no": "6222000000000000000",
      "bank_name": "招商银行",
      "status": 1,
      "reject_reason": "",
      "reviewed_at": "2026-06-15T16:30:00+08:00",
      "created_at": "2026-06-15T16:20:00+08:00",
      "updated_at": "2026-06-15T16:30:00+08:00"
    }
  }
}
```

字段说明：

| 字段 | 说明 |
|---|---|
| bank_account_name/bank_account_no/bank_name | 当前已审核通过、可用于提现的正式收款账户 |
| can_withdraw | 是否已有正式收款账户且可提现金额大于 0 |
| gross_amount | 可结算订单实收总额，单位分 |
| refund_amount | 已成功退款金额，单位分 |
| withdraw_amount | 已审核通过提现金额，单位分 |
| pending_withdraw_amount | 待审核提现冻结金额，单位分 |
| available_amount | 当前可提现金额，单位分，计算公式为 `gross_amount - refund_amount - withdraw_amount - pending_withdraw_amount` |
| arrival_cycle | 提现到账周期展示文案，读取平台配置 `withdraw_arrival_cycle`，默认 `T+1 到 T+3 个工作日` |
| pending_audit | 当前待审核申请；没有则为 `null` 或不返回 |
| latest_audit.status | `0` 待审核；`1` 通过；`2` 拒绝 |

### 提交收款账户审核

```http
PUT /api/v1/organizer/withdraw-info
Authorization: Bearer <organizer_token>
```

请求：

```json
{
  "bank_account_name": "张三",
  "bank_account_no": "6222000000000000000",
  "bank_name": "招商银行"
}
```

说明：

- 商家入驻审核通过后才能提交收款账户审核。
- 提交后生成一条 `status=0` 的审核申请，不会立即覆盖正式收款账户。
- 同一商家存在待审核申请时，不允许重复提交。
- 管理员审核通过后，申请中的银行卡信息才会同步到 `organizers.bank_*`，此时 `can_withdraw=true`。
- 管理员拒绝后，商家可修改信息再次提交。

成功响应：

```json
{
  "code": 200,
  "data": {
    "success": true
  }
}
```

### 商家提现列表

```http
GET /api/v1/organizer/withdraws?page=1&size=10&status=0
Authorization: Bearer <organizer_token>
```

查询参数：

| 参数 | 必填 | 说明 |
|---|---|---|
| page | 否 | 默认 1 |
| size | 否 | 默认 10 |
| status | 否 | `0` 待审核；`1` 通过；`2` 拒绝 |

响应：

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": 1,
        "organizer_id": 10,
        "amount": 10000,
        "bank_account_name": "张三",
        "bank_account_no": "6222000000000000000",
        "bank_name": "招商银行",
        "status": 0,
        "remark": "",
        "created_at": "2026-06-15T16:40:00+08:00",
        "updated_at": "2026-06-15T16:40:00+08:00"
      }
    ],
    "total": 1
  }
}
```

### 发起商家提现

```http
POST /api/v1/organizer/withdraws
Authorization: Bearer <organizer_token>
```

请求：

```json
{
  "amount": 10000,
  "remark": "6月活动结算提现"
}
```

说明：

- `amount` 单位为分。
- 只有收款账户审核通过后才能发起提现。
- `amount` 必须大于 0，且不能超过 `GET /api/v1/organizer/withdraw-info` 返回的 `available_amount`。
- 提交提现时会快照当前正式收款账户到提现记录中，避免后续改卡影响历史打款记录。
- 同一商家存在待审核提现申请时，不允许重复提交。
- 提交成功后会生成 `status=0` 的提现记录，商家端通过 `GET /api/v1/organizer/withdraws` 查看申请记录。
- 待审核提现金额会计入 `pending_withdraw_amount`，从 `available_amount` 中冻结。

成功响应：

```json
{
  "code": 200,
  "data": {
    "id": 1
  }
}
```

### 门店列表

```http
GET /api/v1/organizer/stores?page=1&size=10&keyword=三里屯
```

### 创建门店

```http
POST /api/v1/organizer/stores
```

请求：

```json
{
  "name": "Hyper 三里屯店",
  "logo": "https://cdn.xxx/store.jpg",
  "address": "北京市朝阳区三里屯",
  "latitude": 39.9333,
  "longitude": 116.4533,
  "phone": "13800138000"
}
```

### 编辑门店

```http
PUT /api/v1/organizer/stores/:id
```

请求字段同创建门店。

### 删除门店

```http
DELETE /api/v1/organizer/stores/:id
```

---

## 4. 管理后台入驻审核

以下接口需要管理员 Token：

```http
Authorization: Bearer <admin_access_token>
```

### 入驻申请列表

```http
GET /api/v1/admin/organizers?page=1&pageSize=20&status=1&type=venue
```

Query:

| 字段 | 必填 | 说明 |
|---|---|---|
| page | 否 | 页码，默认 `1` |
| pageSize | 否 | 每页数量，默认 `20` |
| status | 否 | 主办方状态，不传返回全部 |
| type | 否 | `venue` / `merchant`，不传返回全部 |

### 入驻申请详情

```http
GET /api/v1/admin/organizers/:id
```

### 审核入驻申请

```http
PUT /api/v1/admin/organizers/:id/audit
```

通过：

```json
{
  "status": 2
}
```

拒绝：

```json
{
  "status": 3,
  "reject_reason": "资料不完整"
}
```

说明：

- `status=2` 表示审核通过，用户入驻成功。
- `status=3` 表示审核拒绝，必须填写 `reject_reason`。
- 普通用户申请 `venue` 或 `merchant` 任意一个通过后，即可视为入驻成功。

### 活动审核列表

```http
GET /api/v1/admin/activities?page=1&pageSize=20&status=1&keyword=周末&organizer_id=1
```

Query:

| 字段 | 必填 | 说明 |
|---|---|---|
| page | 否 | 页码，默认 `1` |
| pageSize | 否 | 每页数量，默认 `20` |
| status | 否 | 活动状态；不传返回已提交过审核的活动，不包含草稿；`1` 待审核，`2` 审核中，`3` 已上架，`4` 审核未通过 |
| keyword | 否 | 按活动名称或地址模糊搜索 |
| organizer_id | 否 | 按主办方 ID 筛选 |

响应：

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": 1,
        "organizer_id": 1,
        "organizer_name": "Hyper Club",
        "organizer_type": "venue",
        "name": "周末电音派对",
        "share_title": "周末见",
        "start_time": "2026-06-12 20:00:00",
        "end_time": "2026-06-13 02:00:00",
        "real_name_mode": 1,
        "minor_check": 1,
        "description": "活动概要",
        "province": "北京市",
        "city": "北京市",
        "district": "朝阳区",
        "address": "朝阳区 xx 路",
        "poster_list": "https://cdn.xxx/list.jpg",
        "qualification_doc": "https://cdn.xxx/doc.jpg",
        "status": 1,
        "reject_reason": "",
        "ticket_spec_count": 2,
        "created_at": "2026-06-03 12:00:00",
        "updated_at": "2026-06-03 12:00:00"
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 20
  }
}
```

### 活动审核详情

```http
GET /api/v1/admin/activities/:id
```

响应包含活动详情、主办方信息和票种列表：

```json
{
  "code": 200,
  "data": {
    "id": 1,
    "organizer_id": 1,
    "name": "周末电音派对",
    "status": 1,
    "reject_reason": "",
    "organizer": {
      "id": 1,
      "type": "venue",
      "name": "Hyper Club"
    },
    "ticket_specs": [
      {
        "id": 1,
        "activity_id": 1,
        "name": "早鸟票",
        "price": 8800,
        "stock": 100
      }
    ]
  }
}
```

### 审核活动

```http
PUT /api/v1/admin/activities/:id/audit
```

通过并上架：

```json
{
  "status": 3
}
```

拒绝：

```json
{
  "status": 4,
  "reject_reason": "活动资质不完整"
}
```

说明：

- `status=3` 表示审核通过并上架。
- `status=4` 表示审核拒绝，必须填写 `reject_reason`。

### 后台票务订单列表

```http
GET /api/v1/admin/orders?page=1&pageSize=20&activity_id=1&status=1&keyword=罗小瑞
```

Query:

| 字段 | 必填 | 说明 |
|---|---|---|
| page | 否 | 页码，默认 `1` |
| pageSize | 否 | 每页数量，默认 `20` |
| activity_id | 否 | 按活动筛选 |
| status | 否 | 按票务订单状态筛选 |
| keyword | 否 | 订单号、购票人姓名、身份证号模糊搜索 |

### 后台票务订单详情

```http
GET /api/v1/admin/orders/:order_no
```

响应包含订单、退款单和退款进度：

```json
{
  "code": 200,
  "data": {
    "order": {
      "order_no": "T2026053114300012ab34cd",
      "status": 1,
      "actual_price": 8800,
      "activity_name": "周末电音派对",
      "ticket_spec_name": "早鸟票",
      "buyer_name": "罗小瑞"
    },
    "refunds": [],
    "refund_logs": []
  }
}
```

### 后台退款审批

```http
POST /api/v1/admin/orders/:order_no/refund/approve
POST /api/v1/admin/orders/:order_no/refund/reject
```

拒绝请求：

```json
{
  "reject_reason": "不符合退款规则"
}
```

说明：后台订单退款审批会流转退款单状态；实际微信退款仍建议使用 `/api/v1/pay/refund/:refund_no/approve` 发起。

### 后台财务结算

```http
GET /api/v1/admin/finance/summary
GET /api/v1/admin/finance/settlements?page=1&pageSize=20
GET /api/v1/admin/finance/settlements/:organizer_id
GET /api/v1/admin/finance/settlements/:organizer_id/export
```

### 后台用户管理

```http
GET /api/v1/admin/users?page=1&pageSize=20&keyword=13800138000
PUT /api/v1/admin/users/:id/status
```

状态请求：

```json
{
  "status": 0
}
```

说明：`status=1` 正常，`status=0` 封禁。

### 后台 Banner 管理

```http
GET /api/v1/admin/banners
POST /api/v1/admin/banners
PUT /api/v1/admin/banners/:id
DELETE /api/v1/admin/banners/:id
PUT /api/v1/admin/banners/sort
```

创建/更新请求：

```json
{
  "title": "首页活动",
  "image": "https://cdn.xxx/banner.jpg",
  "link": "/pages/activity/detail?id=1",
  "position": "home",
  "sort": 1,
  "status": 1
}
```

排序请求：

```json
{
  "list": [
    { "id": 1, "sort": 1 },
    { "id": 2, "sort": 2 }
  ]
}
```

### 后台平台设置

```http
GET /api/v1/admin/settings
PUT /api/v1/admin/settings
```

更新请求：

```json
{
  "settings": [
    {
      "key": "service_fee_rate",
      "value": "0.05",
      "remark": "平台服务费率"
    }
  ]
}
```

### 绑定管理员微信通知

管理员需要先在小程序端授权订阅消息，再把 `wx.login` 得到的 `code` 传给后端绑定 openid。

前端流程：

```js
await wx.requestSubscribeMessage({
  tmplIds: ['6qWdBuY6p3EJ-rCRq7OWonTVAVlj-b6io1cDtNDlYTw']
})
const login = await wx.login()
await request({
  url: '/api/v1/admin/wechat-subscribe',
  method: 'POST',
  data: { code: login.code }
})
```

后端接口：

```http
POST /api/v1/admin/wechat-subscribe
```

请求：

```json
{
  "code": "wx.login 返回的 code"
}
```

说明：

- 绑定成功后，普通用户提交 `/api/v1/organizer/apply` 会自动给已绑定管理员发送订阅消息。
- 模板 ID：`6qWdBuY6p3EJ-rCRq7OWonTVAVlj-b6io1cDtNDlYTw`
- 模板字段：`thing1` 商家名称、`thing2` 申请人、`time3` 申请时间、`phrase4` 状态、`thing5` 备注。

---

## 5. 活动模块

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
  "type": "party",
  "name": "周末电音派对",
  "share_title": "一起蹦到凌晨",
  "start_time": "2026-06-12T20:00",
  "end_time": "2026-06-13T02:00",
  "real_name_mode": 1,
  "minor_check": 1,
  "description": "<p>活动介绍</p>"
}
```

字段说明：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| type | string | 否 | `party` 派对 / `venue` 场地；不传默认 `party`。场地/派对区分从入驻申请迁移到这里 |

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
    "type": "party",
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

### 活动统计总览

```http
GET /api/v1/activity/:id/statistics
```

响应：

```json
{
  "code": 200,
  "data": {
    "verify_rate": 0.5,
    "ticket_count": 20,
    "buyer_count": 18,
    "gross_amount": 176000,
    "refund_amount": 8800,
    "net_amount": 167200,
    "average_ticket_price": 8800,
    "verified_count": 10
  }
}
```

### 活动每日统计

```http
GET /api/v1/activity/:id/statistics/daily?days=7
```

响应：

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "date": "2026-06-11",
        "amount": 8800,
        "ticket_count": 1,
        "order_count": 1
      }
    ],
    "total": 1
  }
}
```

---

## 6. 票券配置模块

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

## 7. 订单模块

### 观演人列表

```http
GET /api/v1/viewers
```

响应：

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": 1,
        "real_name": "罗小瑞",
        "id_card": "5001**********0817",
        "phone": "138****8000",
        "type": 2,
        "created_at": "2026-06-02T12:00:00+08:00",
        "updated_at": "2026-06-02T12:00:00+08:00"
      }
    ],
    "total": 1
  }
}
```

### 创建观演人

```http
POST /api/v1/viewers
```

请求：

```json
{
  "real_name": "罗小瑞",
  "id_card": "500101199811040817",
  "phone": "13800138000"
}
```

响应：

```json
{
  "code": 200,
  "data": {
    "success": true,
    "id": 1
  }
}
```

说明：

- 单个用户最多 5 个常用观演人。
- 身份证号必须通过校验码校验。
- 返回列表时身份证号、手机号会脱敏。

### 更新观演人

```http
PUT /api/v1/viewers/:id
```

请求：

```json
{
  "real_name": "罗小瑞",
  "phone": "13800138000"
}
```

### 删除观演人

```http
DELETE /api/v1/viewers/:id
```

说明：如果观演人已关联未完成票务订单，暂不可删除。

### 创建票务订单

```http
POST /api/v1/order/create
```

请求：

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

响应：

```json
{
  "code": 200,
  "data": {
    "success": true,
    "order_no": "T2026053114300012ab34cd",
    "total_price": 17600,
    "actual_price": 17600,
    "qr_code": "TICKET:T2026053114300012ab34cd:xxxx",
    "qr_code_url": "https://cdn.hypercn.cn/ticket/qrcode/2026/06/12/T2026053114300012ab34cd.png",
    "viewers": [
      {
        "viewer_id": 12,
        "real_name": "罗小瑞",
        "id_card_masked": "5001**********0817",
        "phone_masked": "138****8000",
        "type": 1
      },
      {
        "viewer_id": 13,
        "real_name": "农子健",
        "id_card_masked": "5001**********0818",
        "phone_masked": "138****8001",
        "type": 1
      }
    ]
  }
}
```

说明：

- 活动必须为 `status=3` 已上架
- 票券必须启用并且在销售时间内
- `quantity` 可以大于 1，但不能超过票档 `purchase_limit`，且不能超过剩余库存
- 开启实名模式时，`viewer_ids`/`viewers` 解析后的观演人数必须等于 `quantity`
- `viewer_ids` 用于提交当前账号已保存的观演人；`viewers` 可提交观演人快照，若传 `id` 或 `viewer_id` 则以后端保存的观演人为准
- `buyer_name` 和 `buyer_id_card` 仍兼容旧客户端；实名模式下只买 1 张且未传 `viewer_ids/viewers` 时可继续使用
- 后端会把首位观演人同步到 `buyer_name`/`buyer_id_card`，并把全部观演人写入 `ticket_order_viewers`
- 开启未成年人校验时，每位观演人 18 岁以下均不可购票

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
- 仅 `status=0` 待支付订单可发起预支付
- 若订单已超过 `expire_time`，后端会自动取消订单、关闭待支付流水、返还抵扣积分，并返回错误
- 同一 `order_no` 作为微信 `out_trade_no`，微信侧不会产生两笔同订单支付；后端支付回调也按订单状态幂等处理
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
    "viewers": [
      {
        "viewer_id": 12,
        "real_name": "罗小瑞",
        "id_card": "500101199811040817",
        "id_card_masked": "5001**********0817",
        "phone": "13800138000",
        "phone_masked": "138****8000",
        "type": 1
      },
      {
        "viewer_id": 13,
        "real_name": "农子健",
        "id_card": "500101199901010818",
        "id_card_masked": "5001**********0818",
        "phone": "13800138001",
        "phone_masked": "138****8001",
        "type": 1
      }
    ],
    "pay_method": "",
    "pay_time": null,
    "created_at": "2026-05-31T14:30:00+08:00",
    "qr_code": "TICKET:T2026053114300012ab34cd:xxxx",
    "qr_code_url": "https://cdn.hypercn.cn/ticket/qrcode/2026/06/12/T2026053114300012ab34cd.png",
    "expire_time": "2026-05-31T14:45:00+08:00"
  }
}
```

说明：`qr_code_url` 是用户购票后的线下取票/核销二维码图片；图片内容为 `qr_code` 字符串，核销端扫码后将扫码内容传给 `/api/v1/verifier/scan`。

### 我的票务订单列表

```http
GET /api/v1/order/list?page=1&size=10&status=1
```

查询参数：

| 参数 | 必填 | 说明 |
|---|---|---|
| page | 否 | 页码，默认 1 |
| size | 否 | 每页数量，默认 10 |
| status | 否 | 订单状态；不传返回全部 |
| legacy | 否 | 传 `1` 时返回旧商品订单列表；不传时返回票务订单列表 |

响应：

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "order_no": "T2026053114300012ab34cd",
        "status": 1,
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
          "id": 1,
          "name": "早鸟票"
        },
        "buyer_name": "罗小瑞",
        "buyer_id_card": "5001**********0817",
        "created_at": "2026-05-31T14:30:00+08:00",
        "expire_time": "2026-05-31T14:45:00+08:00",
        "pay_time": "2026-05-31T14:32:00+08:00"
      }
    ],
    "total": 1
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

说明：

- 仅 `status=0` 待支付订单允许用户主动取消。
- 取消后订单状态更新为 `3` 已取消。
- 后端会回滚票券 `sold_count`。
- 如果下单时使用了积分抵扣，后端会把 `points_amount` 返还到积分余额，并写入积分流水。
- 同一订单积分返还按 `order_no + TypeOrderRefund` 做幂等，避免重复返还。

### 待支付订单自动取消

规则：

- 创建票务订单时 `expire_time = created_at + 15 分钟`。
- API 服务启动后每分钟扫描过期票务订单，自动取消 `status=0 AND actual_price>0 AND expire_time<=now` 的订单。
- 用户查询订单列表、查询订单详情、继续支付时，也会兜底触发当前用户过期待支付订单取消。
- 自动取消同样会回滚票券销量、关闭待支付流水、返还抵扣积分。
- 过期后调用 `POST /api/v1/pay/prepay` 会返回错误，不再创建新的微信预支付。

### 删除订单

```http
DELETE /api/v1/order/:order_no
```

说明：

- 这是用户侧“我的订单”删除，本质为软删除/隐藏订单。
- 删除后该订单不再出现在 `GET /api/v1/order/list`，再次请求 `GET /api/v1/order/:order_no` 会按不存在处理。
- 后台订单、退款记录、核销记录、财务统计仍保留，不做物理删除。
- 仅允许删除终态订单：
  - `2`：已使用
  - `3`：已取消
  - `5`：已退款
- 以下状态不允许删除：待支付、待使用、退款中、退款驳回。

响应：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "success": true
  }
}
```

---

### 主办方订单列表

```http
GET /api/v1/organizer/orders?page=1&size=10&activity_id=1&status=1&keyword=罗小瑞&start_date=2026-06-01&end_date=2026-06-30
Authorization: Bearer <access_token>
```

说明：该接口用于商家端“订单与售后”的订单列表，只返回当前登录主办方名下活动的票务订单；用户侧“我的订单”仍使用 `GET /api/v1/order/list`。

参数：

| 参数 | 必填 | 说明 |
|---|---|---|
| page | 否 | 页码，默认 1 |
| size | 否 | 每页数量，默认 10 |
| activity_id | 否 | 按活动 ID 筛选 |
| status | 否 | 按订单状态筛选 |
| keyword | 否 | 匹配订单号、买家昵称/手机号、购票人姓名/身份证、观演人姓名/身份证/手机号 |
| start_date | 否 | 下单开始日期，格式 `YYYY-MM-DD` |
| end_date | 否 | 下单结束日期，格式 `YYYY-MM-DD`，包含当天 |

响应：

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "order_no": "T2026053114300012ab34cd",
        "status": 1,
        "total_price": 17600,
        "actual_price": 17600,
        "points_amount": 0,
        "points_discount": 0,
        "quantity": 2,
        "user_id": 1001,
        "user_name": "邪修的马路路",
        "user_mobile": "138****8000",
        "user_avatar": "https://cdn.xxx/avatar.png",
        "buyer_name": "罗小瑞",
        "buyer_id_card": "5001**********0817",
        "viewers": [
          {
            "viewer_id": 12,
            "real_name": "罗小瑞",
            "id_card_masked": "5001**********0817",
            "phone_masked": "138****8000",
            "type": 1
          }
        ],
        "activity_id": 1,
        "activity_name": "周末电音派对",
        "ticket_spec_id": 1,
        "ticket_spec_name": "早鸟票",
        "pay_method": "JSAPI",
        "pay_time": "2026-05-31T14:32:00+08:00",
        "created_at": "2026-05-31T14:30:00+08:00",
        "expire_time": "2026-05-31T14:45:00+08:00"
      }
    ],
    "total": 1
  }
}
```

字段说明：

- `user_name/user_mobile/user_avatar`：真正的下单账号信息，商家端“买家”应展示这些字段。
- `buyer_name/buyer_id_card`：实名购票兼容字段，当前表示首位观演人，不等同于下单账号。
- `viewers`：本订单全部观演人脱敏信息。

### 主办方订单详情

```http
GET /api/v1/organizer/orders/:order_no
Authorization: Bearer <access_token>
```

说明：

- 只允许当前登录主办方查看自己名下活动的订单。
- 如果订单号存在但不属于当前主办方，返回不存在/无权限错误。
- 返回结构复用用户侧订单详情，并额外返回下单账号信息。

响应：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "order_no": "T2026053114300012ab34cd",
    "status": 1,
    "refund_status": "",
    "total_price": 17600,
    "actual_price": 17600,
    "points_amount": 0,
    "points_discount": 0,
    "quantity": 2,
    "user_id": 1001,
    "user_name": "邪修的马路路",
    "user_mobile": "138****8000",
    "user_avatar": "https://cdn.xxx/avatar.png",
    "activity": {
      "id": 1,
      "name": "周末电音派对",
      "start_time": "2026-06-28T20:00:00+08:00",
      "end_time": "2026-06-28T23:00:00+08:00",
      "poster_list": "https://cdn.xxx/poster.png"
    },
    "ticket_spec": {
      "name": "早鸟票"
    },
    "buyer_name": "罗小瑞",
    "buyer_id_card": "500101199001010817",
    "viewers": [
      {
        "viewer_id": 12,
        "real_name": "罗小瑞",
        "id_card": "500101199001010817",
        "id_card_masked": "5001**********0817",
        "phone": "13800008000",
        "phone_masked": "138****8000",
        "type": 1
      }
    ],
    "pay_method": "JSAPI",
    "pay_time": "2026-05-31T14:32:00+08:00",
    "created_at": "2026-05-31T14:30:00+08:00",
    "qr_code": "TICKET:T2026053114300012ab34cd:xxxx",
    "qr_code_url": "https://cdn.hypercn.cn/ticket/qrcode/2026/06/12/T2026053114300012ab34cd.png",
    "expire_time": "2026-05-31T14:45:00+08:00",
    "refund_info": null
  }
}
```

### 主办方取消未支付订单

```http
POST /api/v1/organizer/orders/:order_no/cancel
Authorization: Bearer <access_token>
Content-Type: application/json
```

请求体中的 `reason_id` 使用 `GET /api/v1/order/cancel-reasons` 返回的取消原因：

```json
{
  "reason_id": 1
}
```

规则：

- 仅当前登录主办方名下活动的订单可操作；其他主办方订单返回 `404`。
- 仅 `status=0` 的待支付订单可取消。已支付、已核销或售后订单返回 `409`，提示“已支付订单请通过退款售后流程处理”。
- 取消会回滚票种销量、返还本订单已抵扣积分，并关闭未支付流水。
- 已取消订单重复提交是幂等操作，不会再次回滚库存或积分。
- 已支付订单请通过退款售后流程处理，不能通过本接口取消。

响应：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "order_no": "T2026053114300012ab34cd",
    "status": 3,
    "cancel_reason_id": 1,
    "cancelled_at": "2026-07-15T15:30:00+08:00"
  }
}
```

### 主办方退款详情

```http
GET /api/v1/organizer/refunds/:refund_no
Authorization: Bearer <access_token>
```

仅返回当前主办方名下活动的退款单。响应包含退款单、关联订单、脱敏观演人、退款日志、支付流水和核销记录。详情见 [主办方退款详情接口](./organizer_refund_detail_api_20260716.md)。

---

## 8. 退款模块

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

### 主办方查看退款审核列表

```http
GET /api/v1/organizer/refunds?page=1&size=10&status=0
Authorization: Bearer <access_token>
```

参数：

| 参数 | 必填 | 说明 |
|---|---|---|
| page | 否 | 页码，默认 1 |
| size | 否 | 每页数量，默认 10 |
| status | 否 | 退款状态，不传返回全部；`0` 表示待审核 |

响应：

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "refund_no": "R20260531150000abcd",
        "status": 0,
        "refund_amount": 8800,
        "deduct_amount": 0,
        "reason": "行程冲突",
        "reject_reason": "",
        "expect_arrive_date": "2026-06-03",
        "wechat_refund_id": "",
        "wechat_status": "",
        "order_no": "T2026053114300012ab34cd",
        "user_id": 1,
        "buyer_name": "罗小瑞",
        "buyer_id_card": "500101199811040817",
        "activity_id": 1,
        "activity_name": "周末电音派对",
        "ticket_spec_id": 1,
        "ticket_spec_name": "早鸟票",
        "quantity": 1,
        "created_at": "2026-05-31T15:00:00+08:00"
      }
    ],
    "total": 1
  }
}
```

### 审核通过并发起微信退款

后台审核退款通过后调用：

```http
POST /api/v1/pay/refund/:refund_no/approve
Authorization: Bearer <access_token>
```

响应：

```json
{
  "code": 200,
  "data": {
    "success": true
  }
}
```

后端处理逻辑：

- 校验退款单必须是 `status=0` 审核中
- 校验对应票务订单必须是 `status=1` 待使用或 `status=4` 退款中
- 审核通过发起微信退款时，后端才会把订单状态更新为 `4` 退款中
- 校验支付流水 `pay_records.pay_status=2`
- 调用微信支付退款 API：`/v3/refund/domestic/refunds`
- 微信受理后把 `refunds.status` 更新为 `1` 退款中
- 微信退款成功回调后把 `refunds.status` 更新为 `2`，订单状态更新为 `5`

### 退款状态展示口径

- 用户提交退款申请后，主订单 `status` 暂不改成 `4`；列表/详情通过 `refund_status=pending_review` 和 `refund_status_text=待审核` 展示售后状态。
- 商家/管理员审核通过并发起退款后，主订单 `status` 才进入 `4` 退款中，同时 `refund_status=refunding`。
- 核销时如果订单存在 `待审核` 或 `退款中` 退款单，后端会拒绝核销。

### 同步微信退款状态

```http
POST /api/v1/pay/refund/:refund_no/sync
Authorization: Bearer <access_token>
```

说明：

- 用于微信退款回调未成功送达、订单长时间停留在“退款中”的兜底处理。
- 后端会调用微信支付“查询单笔退款”接口：`/v3/refund/domestic/refunds/{out_refund_no}`。
- 如果微信返回 `SUCCESS`，后端会同步：
  - `refunds.status = 2`
  - `refunds.wechat_status = SUCCESS`
  - `ticket_orders.status = 5`
  - 追加退款日志
- 如果微信仍返回处理中，则保持 `refunds.status = 1`、`ticket_orders.status = 4`。

响应：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "success": true,
    "refund_no": "R20260620120000abcd",
    "status": 2,
    "wechat_status": "SUCCESS"
  }
}
```

### 审核拒绝退款

```http
POST /api/v1/refund/:refund_no/reject
Authorization: Bearer <access_token>
```

请求：

```json
{
  "reject_reason": "活动开始前24小时内不可退款"
}
```

响应：

```json
{
  "code": 200,
  "data": {
    "success": true
  }
}
```

后端处理逻辑：

- 只允许主办方拒绝自己活动的退款
- 退款单必须是 `status=0` 审核中
- 拒绝后 `refunds.status=3`
- 拒绝后 `ticket_orders.status=6`
- 写入一条 `refund_logs`

### 微信退款回调

微信商户平台退款通知地址配置为：

```http
POST /api/v1/pay/refund-notify
```

服务端会验签、解密并处理退款结果。前端不需要调用。

### 获取退款详情

```http
GET /api/v1/refund/:refund_no
```

### 取消退款

```http
POST /api/v1/refund/:refund_no/cancel
```

---

## 9. 核销模块

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
    "wechat_mini_program_code_url": "https://cdn.hypercn.cn/verifier/qrcode/2026/06/11/1.png",
    "wechat_qr_url": "https://cdn.hypercn.cn/verifier/qrcode/2026/06/11/1.png",
    "wechat_qr_page": "pages/user-sub/verifier-bind/index",
    "wechat_scene": "v=1",
    "wechat_qr": "https://cdn.hypercn.cn/verifier/qrcode/2026/06/11/1.png",
    "douyin_qr": "hyper://verifier/activate?verifier_id=1&channel=douyin"
  }
}
```

### 获取核销员绑定确认信息

```http
GET /api/v1/verifier/activation-info?v=1
```

说明：微信小程序码只携带短参数 `scene=v%3D1`，不会直接带 `organizerName` 这种中文 query。小程序绑定页解析 `v` 后调用本接口获取主办方名称。

响应：

```json
{
  "code": 200,
  "data": {
    "verifier_id": 1,
    "name": "核销员A",
    "phone": "13800138000",
    "status": 0,
    "is_bound": false,
    "organizer_id": 10,
    "organizer_name": "测试主办方"
  }
}
```

### 核销员激活

```http
POST /api/v1/verifier/activate
Authorization: Bearer <access_token>
```

说明：核销员激活本质是“绑定邀请记录到小程序用户”。小程序扫码进入后，应先完成登录/注册；未注册用户先走现有手机号登录/注册流程，拿到 `access_token` 后再调用本接口。后端会校验登录用户手机号和后台添加的核销员手机号一致。

请求：

```json
{
  "phone": "13800138000"
}
```

响应：

```json
{
  "code": 200,
  "data": {
    "success": true,
    "verifier_id": 1,
    "user_id": 10001,
    "status": 1
  }
}
```

### 扫码识别订单

```http
POST /api/v1/verifier/scan
```

请求：

```json
{
  "qr_code": "TICKET:T2026053114300012ab34cd:xxxx"
}
```

说明：`activity_id` 为可选字段。普通扫码核销不需要传；如果前端处于某个活动的专场核销页面，可以传 `activity_id`，后端会校验票是否属于该活动，不属于则返回 `WRONG_ACTIVITY`。

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

## 10. 前端推荐流程

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
2. 主办方获取激活码：`GET /api/v1/organizer/verifier/:id/activation-qr`
3. 核销员扫码进入小程序，未注册则先登录/注册
4. 核销员绑定邀请：`POST /api/v1/verifier/activate`
5. 扫码识别：`POST /api/v1/verifier/scan`
6. 确认核销：`POST /api/v1/verifier/confirm`

---

## 11. 当前实现边界

- 新票务订单使用 `ticket_orders`，旧商品订单仍使用 `orders/products/pay`。
- 微信支付已支持通过 `/api/v1/pay/prepay` + `order_no` 支付 `ticket_orders`。
- 微信退款已支持通过 `/api/v1/pay/refund/:refund_no/approve` 发起，回调地址为 `/api/v1/pay/refund-notify`。
- 核销员当前临时使用 `X-Verifier-Id`，后续可升级为独立核销员 token。
