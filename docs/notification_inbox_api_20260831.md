# 用户通知收件箱 API

更新时间：2026-08-31

## 背景

系统消息 / 互动通知（点赞、评论、收藏、关注）/ 支付消息统一落库到 `user_notifications` 表，WebSocket 只负责在线实时提醒，离线消息通过收件箱接口补拉。

业务方调用 `service.INotificationService.Notify(ctx, userID, type, title, content, payload)` 写入通知；失败只记日志，不影响主流程。

## 通知类型 type

| 取值 | 说明 |
| --- | --- |
| `system` | 系统消息 |
| `interaction` | 互动通知（点赞/评论/收藏/关注） |
| `payment` | 支付消息（支付/退款/核销） |

## 接口

以下接口均需登录：`Authorization: Bearer <access_token>`，前缀 `/api`。

### 1. 通知列表

```http
GET /api/v1/notifications?type=&page=1&size=20
```

- `type`：可选，`system` / `interaction` / `payment`，缺省返回全部类型
- `page`：页码，默认 1
- `size`：每页条数，默认 20，上限 50

按 id 倒序返回：

```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": 101,
        "type": "interaction",
        "title": "有人赞了你的笔记",
        "content": "xxx 赞了你的笔记",
        "payload": "{\"note_id\":123}",
        "is_read": false,
        "created_at": "2026-08-31 12:00:00"
      }
    ],
    "total": 35
  }
}
```

`payload` 为 JSON 字符串，携带跳转参数，前端按 `type` 解析。

### 2. 未读数

```http
GET /api/v1/notifications/unread-count
```

```json
{
  "code": 0,
  "data": {
    "total": 8,
    "system": 2,
    "interaction": 5,
    "payment": 1
  }
}
```

### 3. 标记已读

```http
POST /api/v1/notifications/read
Content-Type: application/json

{"ids": [101, 102]}
```

仅标记当前用户本人的通知，重复标记幂等。

### 4. 全部已读

```http
POST /api/v1/notifications/read-all
Content-Type: application/json

{"type": "interaction"}
```

`type` 可选，缺省（或空 body）表示全部类型标记已读。

## WebSocket 实时事件

通知写入后通过 MQ（topic `HYPER_SYSTEM_MSGS`，`SystemMessage.type = "user_notification"`）向在线连接推送：

```
event: notice.new
```

载荷：

```json
{
  "user_id": 6,
  "type": "interaction",
  "title": "有人赞了你的笔记",
  "content": "xxx 赞了你的笔记",
  "payload": "{\"note_id\":123}",
  "created_at": "2026-08-31 12:00:00"
}
```

客户端收到 `notice.new` 后可刷新未读数角标；离线期间的通知通过收件箱接口补拉。
