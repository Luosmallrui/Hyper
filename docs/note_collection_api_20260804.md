# 帖子收藏接口

更新时间：2026-08-04

## 通用约定

- 接口前缀：`/api/v1`
- 除“收藏数量”外，均需要登录态：

```http
Authorization: Bearer <access_token>
```

- 帖子 ID 使用 URL 路径参数 `note_id`，必须为正整数。
- 收藏关系以 `note_collections` 中的有效记录（`status = 1`）为准。帖子详情返回的 `is_collected` 与下述状态接口使用同一判断逻辑。
- 业务成功统一返回 HTTP `200`，响应体中的 `code` 为 `200`。

## 1. 收藏帖子

```http
POST /api/v1/note/:note_id/collect
Authorization: Bearer <access_token>
```

请求体为空。

成功响应：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "collected": true
  }
}
```

说明：

- 重复收藏是幂等操作，已收藏时仍返回成功。
- 成功后帖子收藏数 `coll_count` 加一；重复请求不会重复增加。

## 2. 取消收藏

```http
DELETE /api/v1/note/:note_id/collect
Authorization: Bearer <access_token>
```

请求体为空。

成功响应：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "collected": false
  }
}
```

说明：

- 未收藏时取消收藏同样是幂等操作，返回成功。
- 成功取消后收藏数 `coll_count` 减一，最小为 `0`。

## 3. 查询当前用户收藏状态

```http
GET /api/v1/note/:note_id/collect
Authorization: Bearer <access_token>
```

成功响应：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "collected": true
  }
}
```

`collected` 为 `true` 表示当前登录用户已收藏该帖子。

> 帖子详情 `GET /api/v1/note/:note_id` 在携带登录 Token 时也会直接返回 `data.is_collected`；详情页无需额外请求本接口。

## 4. 查询收藏数量

```http
GET /api/v1/note/:note_id/collections/count
```

该接口可匿名调用。

成功响应：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "collect_count": 12
  }
}
```

`collect_count` 为帖子累计有效收藏数量。

## 5. 我的收藏列表

```http
GET /api/v1/note/my/collects?page=1&pagesize=20
Authorization: Bearer <access_token>
```

查询参数：

| 参数 | 必填 | 说明 |
|---|---|---|
| `page` | 否 | 页码，从 `1` 开始，默认 `1` |
| `pagesize` | 否 | 每页数量，默认 `20`，范围 `1-100` |

注意：当前接口参数名为全小写 `pagesize`，不要传 `pageSize`。

成功响应：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "notes": [
      {
        "id": 10001,
        "user_id": 35,
        "title": "周末骑行",
        "content": "成都周边路线分享",
        "topic_ids": [1, 2],
        "location": {
          "lat": 30.657,
          "lng": 104.066,
          "name": "天府三街",
          "address": "成都市高新区"
        },
        "media_data": [
          {
            "url": "https://cdn.hypercn.cn/note/example.png",
            "thumbnail_url": "",
            "width": 1080,
            "height": 1440,
            "duration": 0
          }
        ],
        "type": 1,
        "status": 1,
        "visible_conf": 1,
        "created_at": "2026-08-04T10:00:00+08:00",
        "updated_at": "2026-08-04T10:00:00+08:00"
      }
    ],
    "total": 1
  }
}
```

列表按收藏时间倒序返回；已删除、违规或当前用户无权查看的帖子会自动排除。

## 错误响应

业务错误的 HTTP 状态目前保持 `200`，以响应体 `code` 和 `msg` 判断结果。

```json
{
  "code": 401,
  "msg": "未登录"
}
```

常见错误：

| `code` | 场景 |
|---|---|
| `400` | `note_id` 为空或不是正整数；列表分页参数不合法 |
| `401` | 收藏、取消收藏、状态或我的收藏列表未携带有效 Token |
| `500` | 帖子不存在，或数据库/统计更新失败 |

## 前端接入建议

1. 详情页初始状态直接使用 `GET /api/v1/note/:note_id` 的 `is_collected` 与 `coll_count`。
2. 点击收藏后调用 `POST /note/:note_id/collect`，成功后将本地 `is_collected` 设为 `true`；点击取消后调用对应 `DELETE` 并设为 `false`。
3. 页面重新进入或下拉刷新时，以详情接口返回的 `is_collected` 为最终状态，不要只依赖本地缓存。
4. 未登录点击收藏时先引导登录，不要以匿名 Token 调用写接口。
