# PC 前端接口对接更新清单

更新时间：2026-07-16

本文汇总本次已实现或修正的商家端 PC、管理端 PC 接口。所有金额字段均以分为单位。

## 1. 商家端

统一前缀：`/api/v1/organizer`，请求头：

```http
Authorization: Bearer <organizer_access_token>
```

### 1.1 商家资料

```http
GET /profile
PUT /profile
```

资料字段：

```json
{
  "name": "Hyper Club",
  "logo": "https://cdn.xxx/logo.png",
  "cover_image": "https://cdn.xxx/cover.png",
  "gallery": ["https://cdn.xxx/1.png"],
  "description": "商家简介",
  "business_hours": "19:30-02:30",
  "contact_name": "张三",
  "service_phone": "13800138000",
  "province": "四川省",
  "city": "成都市",
  "district": "武侯区",
  "address": "天府三街",
  "latitude": 30.657,
  "longitude": 104.066,
  "average_spend": 8800
}
```

- `average_spend` 为人均消费，单位分。
- 经度、纬度非 `0` 时后端校验中国大陆常用范围。

### 1.2 商家等级与手续费

```http
GET /info
```

重点字段：

```json
{
  "level": "LV2",
  "level_value": 2,
  "service_fee_rate": 0.03,
  "fee_rate": 0.03,
  "completed_activity_count": 5,
  "next_level_required_count": 10
}
```

规则由管理端等级配置决定。默认规则为：

| 已上线且已结束活动数 | 等级 | 手续费 |
| --- | --- | --- |
| `0-4` | LV1 | 5% |
| `5-9` | LV2 | 3% |
| `>=10` | LV3 | 0% |

草稿、审核中和驳回活动不计入活动数。停用商家访问 `/info` 会被拒绝，PC 应退出商家控制台并提示“商家账号已停用”。

### 1.3 核销员启停

```http
PATCH /verifier/:id/status
```

```json
{
  "status": 1
}
```

- `1` 启用，`0` 停用。
- 停用后 `POST /api/v1/verifier/confirm` 无法完成核销。
- 扫码预览可继续显示订单信息，但不会改变订单状态。

### 1.4 活动合集

```http
GET    /collections?page=1&size=10&keyword=
GET    /collections/:id
POST   /collections
PUT    /collections/:id
DELETE /collections/:id
```

新增或编辑请求：

```json
{
  "title": "周末活动合集",
  "share_title": "这个周末去哪里",
  "description": "合集简介",
  "share_image": "https://cdn.xxx/share.png",
  "status": 1,
  "activity_ids": [8, 9, 10]
}
```

列表返回 `activity_count`、`created_at`、`updated_at`；详情额外返回 `activity_ids` 与 `activities`。`activity_ids` 必须全部属于当前主办方。

### 1.5 主办方退款详情

```http
GET /refunds/:refund_no
```

返回：

```text
refund                 当前退款单及微信退款状态
order                  关联订单、买家、活动、票种和支付方式
viewers                脱敏观演人
refund_logs            当前退款单的处理时间线
pay_records            安全展示字段的支付流水
verification_records   核销记录
```

- 跨主办方或不存在的退款单返回 `404`。
- 详情中的买家和核销员手机号、观演人手机号和证件号均已脱敏。
- 审核或退款状态变化后重新请求本接口刷新详情。

### 1.6 商家动态管理

```http
GET    /posts?page=1&size=10&keyword=&status=&activity_id=&store_id=
POST   /posts
PUT    /posts/:id
PATCH  /posts/:id/visibility
DELETE /posts/:id
```

显示或隐藏：

```json
{
  "visible": false
}
```

也可以传：

```json
{
  "status": 2
}
```

动态发布和编辑支持标题、正文、图片/媒体、活动关联、场地关联、可见范围和状态。主办方只能操作自己的动态。

### 1.7 平台消息与图文详情

```http
GET  /messages?page=1&size=10&unread_only=0
GET  /messages/:id
POST /messages/:id/read
POST /messages/read-all
```

详情响应重点字段：

```json
{
  "id": 1001,
  "title": "商家服务费规则调整通知",
  "content": "<p>消息正文</p>",
  "content_type": "rich_text",
  "cover_image": "https://cdn.xxx/cover.png",
  "media_data": ["https://cdn.xxx/1.png"],
  "type": "announcement",
  "is_read": true,
  "read_at": "2026-07-16T12:00:00+08:00",
  "created_at": "2026-07-16T10:00:00+08:00"
}
```

- `content_type`：`text` 或 `rich_text`。
- 打开 `GET /messages/:id` 会自动标记已读。
- 富文本渲染必须使用前端安全白名单，不可直接信任危险标签或链接。

## 2. 管理端

统一前缀：`/api/v1/admin`，请求头：

```http
Authorization: Bearer <admin_access_token>
```

### 2.1 商家启停

```http
GET   /organizers?page=1&pageSize=20&status=&type=
PATCH /organizers/:id/status
```

```json
{
  "enabled": 0
}
```

- `enabled=0` 停用，`enabled=1` 启用。
- 仅审核通过的商家可启用。
- 停用后商家端控制台入口和管理能力会被拒绝。

### 2.2 分类、优惠标签与动态频道

```http
GET    /categories?type=activity&page=1&pageSize=20
POST   /categories
PUT    /categories/:id
DELETE /categories/:id
```

```json
{
  "type": "coupon_tag",
  "name": "积分立减",
  "image": "https://cdn.xxx/tag.png",
  "value": "points_discount",
  "sort": 1,
  "status": 1
}
```

分类类型：

| `type` | 用途 |
| --- | --- |
| `activity` | 活动类型，支持图片、名称、排序、启停 |
| `note_channel` | 动态频道，如演出、骑行、活动 |
| `coupon_tag` | 优惠标签，如积分立减、买单立减、新人优惠 |
| `distance` | 首页距离筛选 |
| `other` | 其它分类 |

`status=1` 为启用，`status=0` 为停用。该值现在可真实保存，前端不要把 `0` 省略。

### 2.3 客户端用户与互动记录

```http
GET /users?page=1&pageSize=20&keyword=
PUT /users/:id/status
```

```http
GET /users/:id/records/likes?page=1&pageSize=20
GET /users/:id/records/collections?page=1&pageSize=20
GET /users/:id/records/following?page=1&pageSize=20
GET /users/:id/records/followers?page=1&pageSize=20
GET /users/:id/records/attends?page=1&pageSize=20
GET /users/:id/records/subscribes?page=1&pageSize=20
```

记录语义：

- `likes`：该用户发布动态收到的点赞，返回点赞用户、动态标题和时间。
- `collections`：该用户自己的收藏记录。
- `following`：该用户关注的人。
- `followers`：关注该用户的人。

### 2.4 管理端动态管理

```http
GET /notes?page=1&pageSize=20&status=&keyword=
PUT /notes/:id/status
```

状态更新：

```json
{
  "status": 0
}
```

状态约定：

| 值 | 含义 |
| --- | --- |
| `1` | 公开 |
| `0` | 隐藏或关闭 |
| `-1` | 删除 |

客户端广场、关注流、场地和活动相关动态只会展示 `status=1` 的公开动态。管理端可查询隐藏动态后用 `status=1` 恢复公开。

动态互动记录：

```http
GET /notes/:id/records/likes
GET /notes/:id/records/collections
GET /notes/:id/records/comments
GET /notes/:id/records/shares
```

### 2.5 平台消息发布

```http
POST /messages
```

```json
{
  "title": "商家服务费规则调整通知",
  "content": "<p>消息正文</p>",
  "content_type": "rich_text",
  "cover_image": "https://cdn.xxx/cover.png",
  "media_data": ["https://cdn.xxx/1.png"],
  "type": "announcement",
  "target": "organizer",
  "status": 1
}
```

面向商家发布时，`target` 使用 `organizer`、`merchant`、`business` 或 `all`。`status=1` 会立即创建投递记录并发送站内消息。

## 3. 部署前 SQL

图文平台消息上线前，生产数据库必须执行：

```sql
ALTER TABLE `platform_messages`
  ADD COLUMN `content_type` varchar(20) NOT NULL DEFAULT 'text' COMMENT '内容类型：text/rich_text' AFTER `content`,
  ADD COLUMN `cover_image` varchar(255) NOT NULL DEFAULT '' COMMENT '消息封面图' AFTER `content_type`,
  ADD COLUMN `media_data` text NULL COMMENT '消息图集JSON数组' AFTER `cover_image`;
```

历史消息自动按纯文本和空图集兼容。
