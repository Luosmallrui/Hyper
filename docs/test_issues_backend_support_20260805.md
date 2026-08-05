# 测试问题后端支持更新

更新时间：2026-08-05

本文对应 `test_issues_backend_support_20260804.md`。以下代码已完成，部署活动优惠标签功能前须先执行文末 SQL。

## 1. 地图 markers 优惠标签筛选

接口：

```http
GET /api/v1/map/markers?source=all&tag_ids=1,4
```

`tags` 与 `tag_ids` 均可使用，标签值约定：

| 值 | 名称 |
|---|---|
| `1` | 积分立减 |
| `2` | 买单立减 |
| `4` | 新人优惠 |

多个值按“同时满足”过滤。例如 `tag_ids=1,4` 只返回同时支持积分立减和新人优惠的活动。未传标签参数时不做标签过滤。

地图 marker 新增/恢复真实字段：

```json
{
  "id": "activity-15",
  "tag_ids": [1, 4],
  "discount_tags": ["积分立减", "新人优惠"]
}
```

活动创建或编辑时可通过既有接口写入标签：

```http
POST /api/v1/activity/create
Authorization: Bearer <access_token>
Content-Type: application/json
```

```json
{
  "activity_id": 15,
  "step": 1,
  "tag_ids": [1, 4]
}
```

- `tag_ids` 不传：不修改已有标签。
- `tag_ids: []`：清空标签。
- 传入非 `1`、`2`、`4` 的值会返回参数错误。

## 2. 旧商家详情与活动详情的 owner user_id

### 旧兼容详情

```http
GET /api/v1/merchant/:id
```

旧 `parties` 表没有对应记录时，接口会回退查询已认证、启用的 `organizers` 场地资料，不再返回 `user_id: 0` 的空对象。响应中的：

```json
{
  "id": 15,
  "user_id": 35,
  "is_follow": false,
  "is_subscribe": false
}
```

`user_id` 为主办方所属用户 ID，可直接用于原有关注接口。新客户端仍推荐优先使用 `GET /api/v1/venues/:id` 与 `POST/DELETE /api/v1/venues/:id/follow`。

### 活动详情

```http
GET /api/v1/activity/:id
```

响应顶层已新增 `user_id`，值为活动主办方用户 ID；原有 `organizer.user_id` 同时保留：

```json
{
  "id": 15,
  "user_id": 35,
  "tag_ids": [1],
  "discount_tags": ["积分立减"],
  "organizer": {
    "id": 7,
    "user_id": 35
  }
}
```

## 3. 商家密码登录 token 续期

已确认 `POST /api/v1/auth/login-password` 与验证码登录使用相同 token 响应结构，包含：

```json
{
  "access_token": "...",
  "refresh_token": "...",
  "access_expire": 1780000000,
  "refresh_expire": 1780600000
}
```

前端在收到 `401` 时可以携带 `refresh_token` 调用：

```http
POST /api/v1/auth/refresh
Authorization: Bearer <refresh_token>
```

## 4. 小程序 Token 调用账户接口

以下接口只校验通用用户 JWT，不限制 PC 来源；小程序用户 Token 可以直接调用：

```http
PUT /api/v1/organizer/basic
GET /api/v1/organizer/profile
PUT /api/v1/organizer/profile
GET /api/v1/auth/profile
POST /api/v1/auth/send-code
POST /api/v1/auth/reset-password
```

`send-code` 和 `reset-password` 当前按手机号统一校验，前端请求体保持既有 `{ "phone": "..." }` 即可，不需要额外传 `scene` 或 `type`。

## 5. 帖子收藏接口

帖子收藏状态已修复为以 `note_collections.status` 为准，避免 Redis 未同步时详情页始终显示未收藏。完整前端契约见：[note_collection_api_20260804.md](note_collection_api_20260804.md)。

部署包含本次改动的后端版本后，以下接口即可在测试和生产环境联调：

```http
POST   /api/v1/note/:note_id/collect
DELETE /api/v1/note/:note_id/collect
GET    /api/v1/note/:note_id/collect
GET    /api/v1/note/:note_id/collections/count
GET    /api/v1/note/my/collects?page=1&pagesize=20
```

## 部署 SQL

生产库已有 `activities` 表时，先检查字段，再执行一次迁移：

```sql
SHOW COLUMNS FROM activities LIKE 'discount_tags';

ALTER TABLE activities
  ADD COLUMN discount_tags INT NOT NULL DEFAULT 0
  COMMENT '优惠标签位: 1积分立减 2买单立减 4新人优惠'
  AFTER type;
```

可按实际需要为已有活动补充测试标签：

```sql
-- 同时支持积分立减和新人优惠：1 | 4 = 5
UPDATE activities SET discount_tags = 5 WHERE id = 15;
```

旧活动默认值为 `0`，不会命中任何带 `tag_ids` 的筛选条件；这符合“未配置优惠标签即不展示对应优惠”的口径。
