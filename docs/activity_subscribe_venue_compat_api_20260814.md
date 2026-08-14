# 活动订阅接口兼容场地活动

更新时间：2026-08-14

## 背景

新场地保存于 `activities.type = venue`，但订阅关系按场地主办方存储在 `venue_subscriptions`。此前调用：

```http
POST /api/v1/activity/15/subscribe
```

会因场地被活动订阅接口排除而返回“活动不存在”。

## 兼容后的接口

继续使用活动 ID 调用：

```http
POST /api/v1/activity/:id/subscribe
POST /api/v1/activity/:id/unsubscribe
Authorization: Bearer <access_token>
```

后端按活动类型自动选择订阅关系：

| `activities.type` | 订阅表 | 订阅目标 |
| --- | --- | --- |
| `party` | `activity_subscriptions` | `activities.id` |
| `venue` | `venue_subscriptions` | `activities.organizer_id` |

所以场地活动 `id=15, organizer_id=9` 调用 `/activity/15/subscribe` 会写入：

```text
venue_subscriptions(organizer_id=9, user_id=当前用户)
```

主办方订阅自己发布的场地或派对是允许的；重复订阅幂等成功。

## 订阅状态

以下接口会根据活动类型读取正确的订阅表并返回 `is_subscribe`：

```http
GET /api/v1/activity/:id
GET /api/v1/map/markers
```

前端不需要因场地增加单独的订阅按钮分支；直接使用活动 ID 调用订阅接口，并以接口返回的 `is_subscribe` 为准。
