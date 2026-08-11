# C 端商家主页接口

更新时间：2026-08-11

Base URL：`/api/v1`

已审核通过且启用的商家，可从活动详情或场地详情进入独立的公开商家主页。主页展示商家资料，并分别展示其已上架活动和场地；不暴露商家后台、银行卡或审核资料。

## 1. 获取商家主页

```http
GET /api/v1/organizers/:id?activity_page=1&activity_size=10&venue_page=1&venue_size=10
Authorization: Bearer <access_token>   # 可选
```

| 参数 | 必填 | 说明 |
|---|---|---|
| `id` | 是 | 商家/主办方 ID，即活动详情中的 `organizer.id`。 |
| `activity_page` | 否 | 活动区页码，默认 `1`。 |
| `activity_size` | 否 | 活动区每页数量，默认 `10`，最大 `100`。 |
| `venue_page` | 否 | 场地区页码，默认 `1`。 |
| `venue_size` | 否 | 场地区每页数量，默认 `10`，最大 `100`。 |

接口匿名可访问。带登录 Token 时，响应会按当前用户返回商家、活动及场地的关注/订阅状态。

成功响应示例：

```json
{
  "code": 200,
  "data": {
    "id": 7,
    "user_id": 51,
    "name": "Hyper Club",
    "logo": "https://cdn.hypercn.cn/organizer/logo.png",
    "owner_nickname": "Hyper 官方",
    "owner_avatar": "https://cdn.hypercn.cn/avatar.png",
    "cover_image": "https://cdn.hypercn.cn/organizer/cover.png",
    "gallery": ["https://cdn.hypercn.cn/organizer/1.png"],
    "description": "周末派对与演出活动主办方",
    "business_hours": "10:00-22:00",
    "service_phone": "13800000000",
    "province": "四川省",
    "city": "成都市",
    "district": "武侯区",
    "address": "天府三街 88 号",
    "latitude": 30.657,
    "longitude": 104.066,
    "average_spend": 7600,

    "follow_count": 28,
    "is_follow": false,
    "follow_target_type": "organizer",
    "follow_target_id": 7,

    "activity_count": 2,
    "venue_count": 1,
    "activities": {
      "list": [
        {
          "id": 15,
          "type": "party",
          "name": "周末电音派对",
          "poster_list": "https://cdn.hypercn.cn/ticketing/poster_list.png",
          "start_time": "2026-08-15T20:00:00+08:00",
          "end_time": "2026-08-15T23:00:00+08:00",
          "status": 3,
          "is_follow": false,
          "is_subscribe": false,
          "follow_target_type": "activity",
          "follow_target_id": 15
        }
      ],
      "total": 2
    },
    "venues": {
      "list": [
        {
          "id": 7,
          "name": "Hyper Club",
          "cover_image": "https://cdn.hypercn.cn/organizer/cover.png",
          "description": "场地介绍",
          "business_hours": "10:00-22:00",
          "address": "天府三街 88 号",
          "is_follow": false,
          "is_subscribe": false,
          "follow_target_type": "venue",
          "follow_target_id": 7
        }
      ],
      "total": 1
    }
  }
}
```

## 2. 展示与跳转约定

1. 活动详情中已有 `organizer.id`，点击主办方区域后跳转商家主页，并请求 `GET /organizers/:id`。
2. `activities.list` 只包含该商家 `type=party` 且 `status=3` 的活动。派对使用 `start_time`、`end_time` 表示单次活动时间；点击后使用既有 `GET /api/v1/activity/:id`。
3. `venues.list` 只包含已上架场地。场地是持续经营实体，使用 `business_hours` 表示营业时间，不展示或使用活动的 `start_time`、`end_time`；点击后使用既有 `GET /api/v1/venues/:id`。
4. 现有新架构中，一个商家对应一份 `organizer_profiles` 场地资料，因此当前 `venues.total` 为 `0` 或 `1`。门店信息仍在场地详情的 `stores` 字段展示。
5. 未审核、已驳回或被停用的商家主页不可见，前端按不存在处理。

### 场地与派对的时间填写

场地创建或编辑时，使用既有 `POST /api/v1/activity/create`，但传入 `type: "venue"` 与 `business_hours`；不需要也不会使用 `start_time`、`end_time`：

```json
{
  "step": 1,
  "type": "venue",
  "name": "Hyper Club 场地",
  "business_hours": "每天 10:00-02:00"
}
```

`business_hours` 会写入该商家的 `organizer_profiles.business_hours`，用于场地详情和商家主页展示。传给场地的 `start_time`、`end_time` 会被忽略，兼容旧前端残留字段；后端会自动将底层兼容有效期写为创建/切换时刻至 20 年后。该有效期不对 C 端展示，也不代表营业时间。

派对使用 `type: "party"`，并继续传具体日期时间的 `start_time`、`end_time`；不能用 `business_hours` 代替。

## 3. 商家主页关注

商家主页本身有独立关注关系，不等同于关注某个活动、场地或商家账号。

关注：

```http
POST /api/v1/follow/follow
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "target_type": "organizer",
  "target_id": 7
}
```

取消关注：

```http
POST /api/v1/follow/unfollow
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "target_type": "organizer",
  "target_id": 7
}
```

商家主页响应中的 `follow_count`、`is_follow` 仅对应以上 `organizer` 关注关系；活动和场地卡片继续使用各自返回的 `follow_target_type`、`follow_target_id`。

## 4. SQL

不需要新增数据表。商家主页关注复用既有 `content_follows` 表；其 `target_type` 为字符串，新增 `organizer` 枚举不涉及数据库迁移。
