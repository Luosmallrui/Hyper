# 帖子详情点赞状态修复

更新时间：2026-08-14

## 修复内容

`GET /api/v1/note/:note_id` 的 `is_liked` 现在会在 Redis 未命中时回查 `note_likes` 数据库记录，不会再把冷缓存误判为未点赞。

同时修复以下历史数据场景：同一用户曾取消点赞，`note_likes` 保留 `status = 0` 记录时，再次点赞会将该记录恢复为 `status = 1`，不会因唯一键冲突提示“已经点赞过了”。

## 前端影响

接口与字段均不变，无需改动请求：

```http
GET /api/v1/note/:note_id
Authorization: Bearer <access_token>
```

```json
{
  "code": 200,
  "data": {
    "id": 2087463119642169344,
    "is_liked": true
  }
}
```

注意：详情接口允许匿名访问。未携带 `Authorization` 时，`is_liked` 必然为 `false`；小程序请求详情时应继续携带当前用户的 Bearer Token。

点赞接口保持不变，重复调用按幂等成功处理：

```http
POST /api/v1/note/:note_id/like
Authorization: Bearer <access_token>
```

```json
{
  "code": 200,
  "data": {
    "liked": true
  }
}
```
