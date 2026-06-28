# 新场地与订阅聚合接口文档

本文档只说明新场地 C 端展示、场地关注/订阅、我的订阅聚合列表接口。

默认请求头：

```http
Authorization: Bearer <access_token>
```

## 1. 数据范围

新场地不再使用旧 `/api/v1/merchant/*`。

场地展示数据来源：

- `organizers`：只返回 `type=venue`、`status=2`、`enabled=1` 的场地。
- `organizer_profiles`：介绍、图册、地址、营业时间、定位等资料。
- `organizer_stores`：门店/场地位置。
- `notes.store_id`：场地相关动态。
- `user_follow`：关注场地主办方用户。
- `venue_subscriptions`：订阅场地。

我的订阅聚合列表只返回：

- 新活动：`activity_subscriptions`
- 新场地：`venue_subscriptions`

不返回旧派对/旧场地 `party_likes` 数据。

## 2. 场地列表

```http
GET /api/v1/venues?page=1&size=10&keyword=酒吧
```

Query:

| 字段 | 必填 | 说明 |
|---|---|---|
| page | 否 | 默认 1 |
| size | 否 | 默认 10 |
| keyword | 否 | 关键词，匹配场地名称、地址、介绍、行政区 |

响应：

```json
{
  "code": 200,
  "msg": "ok",
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

字段说明：

| 字段 | 说明 |
|---|---|
| id | 场地 ID，即 `organizers.id` |
| user_id | 场地主办方用户 ID |
| average_spend | 人均消费，单位分 |
| is_follow | 当前用户是否关注该场地主办方 |
| is_subscribe | 当前用户是否订阅该场地 |
| follow_count | 场地主办方粉丝数 |
| subscribe_count | 场地订阅数 |
| post_count | 场地相关公开动态数 |

## 3. 场地详情

```http
GET /api/v1/venues/:id
```

响应在场地列表字段基础上额外返回：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "id": 10,
    "name": "SWING 鸡尾酒吧",
    "gallery": [
      "https://cdn.xxx/1.jpg",
      "https://cdn.xxx/2.jpg"
    ],
    "stores": [
      {
        "id": 1,
        "organizer_id": 10,
        "name": "主店",
        "logo": "https://cdn.xxx/store.jpg",
        "address": "天府三街",
        "latitude": 30.657,
        "longitude": 104.066,
        "phone": "13800000000",
        "created_at": "2026-06-28T12:00:00+08:00",
        "updated_at": "2026-06-28T12:00:00+08:00"
      }
    ]
  }
}
```

说明：

- `gallery` 来自 `organizer_profiles.gallery`。
- `stores` 来自 `organizer_stores`。
- 若场地不存在、未审核通过、被禁用，返回错误。

## 4. 场地相关动态

```http
GET /api/v1/venues/:id/notes?cursor=0&pageSize=20
```

Query:

| 字段 | 必填 | 说明 |
|---|---|---|
| cursor | 否 | 上一页 `next_cursor`，第一页传 0 或不传 |
| pageSize | 否 | 默认使用通用分页大小，最大 100 |

说明：

- 后端会查询该场地下所有 `organizer_stores.id`，再按 `notes.store_id` 返回公开动态。
- 只返回 `status != -1` 且 `visible_conf = 1` 的动态。
- `cursor` 使用上一页最后一条 `created_at` 的纳秒时间戳，即响应里的 `next_cursor`。

响应：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "notes": [
      {
        "id": "2060000000000000000",
        "user_id": 2,
        "title": "今晚现场",
        "content": "动态内容",
        "type": 1,
        "media_data": [],
        "activity_id": 0,
        "store_id": 1,
        "avatar": "https://cdn.xxx/avatar.jpg",
        "nickname": "树懒",
        "created_at": "2026-06-28T12:00:00+08:00",
        "updated_at": "2026-06-28T12:00:00+08:00",
        "time_stamp": 1782628800000000000
      }
    ],
    "next_cursor": 1782628800000000000,
    "has_more": false
  }
}
```

## 5. 关注/取消关注场地

关注的是场地所属主办方用户，落表 `user_follow`。

### 关注

```http
POST /api/v1/venues/:id/follow
```

响应：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "is_follow": true
  }
}
```

### 取消关注

```http
DELETE /api/v1/venues/:id/follow
```

响应：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "is_follow": false
  }
}
```

## 6. 订阅/取消订阅场地

订阅场地落表 `venue_subscriptions`。

### 订阅

```http
POST /api/v1/venues/:id/subscribe
```

响应：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "is_subscribe": true
  }
}
```

### 取消订阅

```http
DELETE /api/v1/venues/:id/subscribe
```

响应：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "is_subscribe": false
  }
}
```

## 7. 我的订阅聚合列表

```http
GET /api/v1/subscriptions?page=1&size=20&type=all
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
  "msg": "ok",
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
          "user_id": 2,
          "logo": "https://cdn.xxx/logo.jpg",
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

字段说明：

| 字段 | 说明 |
|---|---|
| id | 前端展示用唯一 ID，格式为 `activity-{id}` 或 `venue-{id}` |
| source | `activity` 新活动；`venue` 新场地 |
| source_id | 对应业务表 ID |
| title | 活动名称或场地名称 |
| cover_image | 活动海报或场地封面 |
| subscribed_at | 订阅时间 |
| extra | 场地额外展示字段；活动一般为空 |
