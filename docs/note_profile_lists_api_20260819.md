# 我的笔记、点赞与收藏列表字段补齐

更新时间：2026-08-19

以下三个登录态接口的每个笔记卡片均返回作者和统计字段，前端无需再以当前登录用户作为作者兜底：

```http
GET /api/v1/user/my-notes?cursor=&pageSize=
GET /api/v1/note/my/likes?page=1&pageSize=20
GET /api/v1/note/my/collects?page=1&pageSize=20
Authorization: Bearer <access_token>
```

`/user/my-notes` 返回 `data.list`；后两个接口返回 `data.notes`。三者的单项字段一致：

```json
{
  "id": 2080000000000000000,
  "user_id": 35,
  "nickname": "笔记作者昵称",
  "avatar": "https://cdn.hypercn.cn/avatars/example.jpg",
  "title": "周末活动记录",
  "content": "正文内容",
  "media_data": [],
  "like_count": 12,
  "coll_count": 3,
  "comment_count": 5,
  "share_count": 1,
  "is_liked": true,
  "created_at": "2026-08-19T12:00:00+08:00"
}
```

字段说明：

- `nickname` / `avatar`：笔记实际作者资料，不是当前浏览者资料。
- `like_count`：笔记累计获赞数。
- `coll_count`、`comment_count`、`share_count`：对应累计统计。
- `is_liked`：当前登录用户是否已点赞该笔记。
- `id` 为雪花 ID；前端应按字符串处理，避免 JavaScript 精度丢失。

无数据时分别返回 `data.list: []` 或 `data.notes: []`，并保持 `code: 200`。
