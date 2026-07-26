# 游客模式匿名浏览接口说明

更新时间：2026-07-24

## 规则

- 未携带 `Authorization` 的浏览请求按游客处理，返回正常公开数据。
- 浏览接口携带有效 Token 时，会解析当前用户并返回对应互动状态。
- 游客的 `is_follow`、`is_subscribe`、`is_liked`、`isFollowed` 等当前用户态字段统一为 `false`。
- 关注、订阅、下单、发帖、评论写入、消息和订单接口仍保持强制登录，返回 `401`。

## 已支持匿名访问

### 首屏与核心链路

```http
GET /api/v1/map/markers
GET /api/v1/category/list
GET /api/v1/districts/tree
GET /api/v1/merchant/tags
GET /api/v1/activity/:id
GET /api/v1/note/related
GET /api/v1/note/list
GET /api/v1/note/:note_id
GET /api/v1/comments/list/:note_id
GET /api/v1/comments/replies/:rootId
GET /api/v1/channel
GET /api/v1/channel/list
GET /api/v1/search/
```

### 场地与他人主页

```http
GET /api/v1/merchant/list
GET /api/v1/merchant/:id
GET /api/v1/merchant/:id/follower/count
GET /api/v1/merchant/:id/goods
GET /api/v1/venues
GET /api/v1/venues/:id
GET /api/v1/venues/:id/notes
GET /api/v1/user/info?user_id=:id
GET /api/v1/user/note?user_id=:id
```

## 隐私约束

`GET /api/v1/user/info?user_id=:id` 可匿名查看他人主页，但仅返回公开资料和统计信息。`phone_number` 在游客访问和非本人访问时为空；不会返回手机号、邮箱、`openid`、实名资料等敏感字段。

## 频道与搜索

- 游客调用 `GET /api/v1/channel` 时，所有频道均放在未订阅列表，不创建 `user:channel:0` Redis 数据。
- 游客搜索不会写搜索历史。

## 自验

不带 Token 请求以下接口应返回 `code: 200`：

```bash
curl -i 'https://www.hypercn.cn/api/v1/map/markers?source=all&limit=200'
curl -i 'https://www.hypercn.cn/api/v1/activity/1'
curl -i 'https://www.hypercn.cn/api/v1/note/list?pageSize=20'
curl -i 'https://www.hypercn.cn/api/v1/search/?type=0&keyword=%E9%9F%B3%E4%B9%90'
```

下列写操作不带 Token 仍应返回 `401`：

```http
POST /api/v1/follow/follow
POST /api/v1/comments/create
POST /api/v1/order/create
POST /api/v1/note/create
```
