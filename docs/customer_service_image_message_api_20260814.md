# 站内 IM 与客服图片消息接口

更新时间：2026-08-14

## 能力范围

普通单聊、群聊和管理端客服工作台均支持图片消息。消息沿用既有 `IM_CHAT_MSGS`、消息落库与 Socket 推送链路，不新增独立聊天表或独立图片消息接口。

## 1. 上传图片

```http
POST /api/v1/upload
Authorization: Bearer <access_token>
Content-Type: multipart/form-data
```

| 字段 | 必填 | 说明 |
|---|---:|---|
| `file` | 是 | 图片文件 |
| `type` | 否 | 建议传 `misc` |

从响应 `data.url` 取得公网 `https` 图片地址。

## 2. 发送单聊或群聊图片

```http
POST /api/v1/message/send
Authorization: Bearer <access_token>
Content-Type: application/json
```

```json
{
  "target_id": "35",
  "session_type": 1,
  "msg_type": 2,
  "content": "https://cdn.hypercn.cn/ticketing/misc/2026/08/14/image.jpg",
  "ext": {
    "image_url": "https://cdn.hypercn.cn/ticketing/misc/2026/08/14/image.jpg",
    "thumbnail_url": "https://cdn.hypercn.cn/ticketing/misc/2026/08/14/image_thumb.jpg",
    "width": 1080,
    "height": 1440
  }
}
```

群聊时将 `session_type` 改为 `2`，`target_id` 传群 ID。

## 3. 客服工作台发送图片

```http
POST /api/v1/admin/customer-service/sessions/:user_id/messages
Authorization: Bearer <admin_access_token>
Content-Type: application/json
```

请求体与普通消息相同，仅不需要传 `target_id` 和 `session_type`：

```json
{
  "msg_type": 2,
  "ext": {
    "image_url": "https://cdn.hypercn.cn/ticketing/misc/2026/08/14/image.jpg",
    "thumbnail_url": "https://cdn.hypercn.cn/ticketing/misc/2026/08/14/image_thumb.jpg",
    "width": 1080,
    "height": 1440
  }
}
```

## 字段约定

- 图片消息固定使用 `msg_type: 2`。
- 图片地址可传 `content` 或 `ext.image_url`，至少提供一个。
- 服务端仅接受 `http` 或 `https` 的图片地址，并将最终地址同时返回在 `content` 与 `ext.image_url`。
- `thumbnail_url`、`width`、`height` 可选，前端可用于缩略图、尺寸占位和预览；未传时直接使用 `content` 展示原图。
- 消息历史接口和 Socket 消息都使用既有的 `msg_type`、`content`、`ext` 字段，无需新增解析通道。

## 错误响应

| 情况 | 响应说明 |
|---|---|
| 缺少图片地址 | `图片消息缺少 image_url` |
| 地址不是 HTTP(S) | `图片消息 image_url 必须是有效的 http 或 https 地址` |
| 文本消息为空 | `消息内容不能为空` |
