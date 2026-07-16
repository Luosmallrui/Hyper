# 商家端平台消息图文详情接口

更新时间：2026-07-16

## 数据库迁移

部署包含本接口的后端前，生产库需要执行：

```sql
ALTER TABLE `platform_messages`
  ADD COLUMN `content_type` varchar(20) NOT NULL DEFAULT 'text' COMMENT '内容类型：text/rich_text' AFTER `content`,
  ADD COLUMN `cover_image` varchar(255) NOT NULL DEFAULT '' COMMENT '消息封面图' AFTER `content_type`,
  ADD COLUMN `media_data` text NULL COMMENT '消息图集JSON数组' AFTER `cover_image`;
```

已存在的历史消息会以 `content_type: "text"`、空封面和空图集兼容展示。

## 管理端发布消息

现有发布接口扩展为支持图文内容：

```http
POST /api/v1/admin/messages
Authorization: Bearer <admin_access_token>
Content-Type: application/json
```

```json
{
  "title": "商家服务费规则调整通知",
  "content": "<p>7 月起将按商家等级结算服务费。</p>",
  "content_type": "rich_text",
  "cover_image": "https://cdn.hypercn.cn/message/cover.png",
  "media_data": [
    "https://cdn.hypercn.cn/message/1.png",
    "https://cdn.hypercn.cn/message/2.png"
  ],
  "type": "announcement",
  "target": "organizer",
  "status": 1
}
```

字段说明：

| 字段 | 说明 |
| --- | --- |
| `content_type` | `text` 或 `rich_text`，默认 `text` |
| `cover_image` | 消息封面图片 URL，可为空 |
| `media_data` | 消息正文图集 URL 数组，可为空 |
| `type` | 建议使用 `system`、`announcement` 等业务类型 |
| `target` | 商家使用 `organizer`、`merchant`、`business` 或 `all` |

## 商家端消息列表

```http
GET /api/v1/organizer/messages?page=1&size=10&unread_only=0
Authorization: Bearer <organizer_access_token>
```

列表返回 `id`、`title`、`content`、`content_type`、`cover_image`、`type`、`is_read`、`read_at`、`created_at`，可用于展示消息名称、发布时间和未读状态。

## 商家端消息详情

```http
GET /api/v1/organizer/messages/:id
Authorization: Bearer <organizer_access_token>
```

成功响应：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "id": 1001,
    "title": "商家服务费规则调整通知",
    "content": "<p>7 月起将按商家等级结算服务费。</p>",
    "content_type": "rich_text",
    "cover_image": "https://cdn.hypercn.cn/message/cover.png",
    "media_data": [
      "https://cdn.hypercn.cn/message/1.png"
    ],
    "type": "announcement",
    "target": "organizer",
    "is_read": true,
    "read_at": "2026-07-16T12:00:00+08:00",
    "created_at": "2026-07-16T10:00:00+08:00"
  }
}
```

规则：

- 仅返回已发布且投递给当前商家的消息。
- 首次打开详情会自动标记为已读，并同步投递记录的阅读状态。
- 消息不存在、未发布或不属于当前商家时返回 `404`。
- 前端仅在 `content_type=rich_text` 时按富文本渲染 `content`；渲染器必须按小程序或网页端安全策略过滤危险标签和链接。

## 阅读接口

以下既有接口保持不变：

```http
POST /api/v1/organizer/messages/:id/read
POST /api/v1/organizer/messages/read-all
```
