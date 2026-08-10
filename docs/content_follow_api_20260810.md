# 活动、场地与派对对象关注接口

更新时间：2026-08-10

## 1. 目标

“关注主办方用户”和“关注某个活动、场地或派对”是两套独立关系：

| 场景 | 存储 | 说明 |
|---|---|---|
| 关注用户/主办方 | `user_follow` | 原有社交关注、粉丝列表、私信通知逻辑，保持不变。 |
| 关注活动、场地、派对 | `content_follows` | 用于内容卡片的 `is_follow` 和 `follow_count`。 |

接口路径没有变。前端在内容卡片操作时，向原有请求体补充 `target_type`、`target_id` 即可。

## 2. 关注与取消关注

### 关注内容

```http
POST /api/v1/follow/follow
Authorization: Bearer <access_token>
Content-Type: application/json
```

```json
{
  "user_id": "35",
  "target_type": "activity",
  "target_id": 15
}
```

### 取消关注内容

```http
POST /api/v1/follow/unfollow
Authorization: Bearer <access_token>
Content-Type: application/json
```

```json
{
  "user_id": "35",
  "target_type": "activity",
  "target_id": 15
}
```

约定：

- 登录用户以 Token 为准，`user_id` 保留是为了兼容现有前端请求结构；对象关注时后端不使用该字段确定操作者。
- `target_type` 与 `target_id` 必须同时传入。
- 不传这两个字段时，接口完全保留原行为：关注/取消关注 `user_id` 对应的用户或主办方。
- 重复关注、重复取消关注均幂等返回成功。
- 不能关注自己的内容；不存在、未上架或不可见的目标返回 `404`。

成功响应：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "followed": true,
    "is_follow": true,
    "target_type": "activity",
    "target_id": 15
  }
}
```

取消关注时 `followed`、`is_follow` 为 `false`。

## 3. 目标类型与 ID

前端必须优先使用接口响应中的 `follow_target_type`、`follow_target_id`，不要根据 `user_id`、`source_id` 或活动类型自行推导。

| `target_type` | 适用内容 | `target_id` |
|---|---|---|
| `activity` | 新票务派对/普通活动 | `activities.id` |
| `venue` | 新场地、`type=venue` 的活动 | `organizers.id` |
| `party` | 旧 `parties` 表派对 | `parties.id` |

特别注意：地图中场地型活动的 `source_id` 是 `activities.id`，但它的 `follow_target_id` 是场地主办方 `organizers.id`。因此应直接传服务端返回的 `follow_target_*`。

## 4. 返回字段

以下接口已按对象维度返回关注状态和粉丝数：

```http
GET /api/v1/map/markers
GET /api/v1/venues
GET /api/v1/venues/:id
GET /api/v1/activity/:id
GET /api/v1/activity/search
GET /api/v1/activity/subscriptions
GET /api/v1/merchant/list
GET /api/v1/merchant/:id
GET /api/v1/merchant/:id/follower/count
```

通用字段：

```json
{
  "is_follow": true,
  "follow_count": 28,
  "follow_target_type": "venue",
  "follow_target_id": 7
}
```

- `is_follow`：当前登录用户是否关注该对象。游客或未登录时为 `false`。
- `follow_count`：关注该对象的人数，不再是主办方账号的用户粉丝数。
- `follow_target_type`、`follow_target_id`：用于调用第 2 节接口的唯一目标标识。
- 地图原有 `current_count` 不表示粉丝数，前端展示粉丝人数必须使用 `follow_count`。

## 5. 场地专用接口

已有接口路径和响应保持不变，但其内部已改为对象级场地关注：

```http
POST   /api/v1/venues/:id/follow
DELETE /api/v1/venues/:id/follow
```

其中 `:id` 为 `organizers.id`。已接入这两个接口的场地详情页不需要改调用路径；列表刷新后将得到正确的场地 `follow_count`。

## 6. 仍属于用户关注的接口

下面接口不带内容目标，继续只处理用户关注关系：

```http
GET /api/v1/follow/:user_id
GET /api/v1/follow/:user_id/followers/count
GET /api/v1/follow/:user_id/following/count
GET /api/v1/follow/list
```

用户主页、个人粉丝列表、社交动态作者卡片应继续使用这一组接口。内容卡片应使用第 2 节的目标参数。

## 7. 前端接入步骤

1. 从地图、场地、活动或派对列表/详情读取 `is_follow`、`follow_count`、`follow_target_type`、`follow_target_id`。
2. 点击关注时，保留原 `user_id` 字段，并补充响应中的目标字段。
3. 点击取消关注时使用同一组字段调用 `/follow/unfollow`。
4. 成功后乐观更新 `is_follow` 和 `follow_count`；回到列表或详情页后以接口最新数据为准。
5. 新派对型活动使用 `activity`，旧派对才使用 `party`。

## 8. 数据库迁移

部署后端代码前，生产库执行：

```sql
CREATE TABLE IF NOT EXISTS `content_follows` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '关注用户ID',
  `target_type` VARCHAR(20) NOT NULL COMMENT 'activity/venue/party',
  `target_id` BIGINT UNSIGNED NOT NULL COMMENT '活动ID、场地主办方ID或旧派对ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_content_follow` (`user_id`, `target_type`, `target_id`),
  KEY `idx_content_follow_target` (`target_type`, `target_id`),
  KEY `idx_content_follow_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
```

旧 `user_follow` 无法准确判断用户当时点击的是主办方名下哪一张内容卡片，因此不自动迁移为对象关注。历史用户关注仍保留在用户社交关系中；对象关注从前端接入目标字段后开始准确记录。
