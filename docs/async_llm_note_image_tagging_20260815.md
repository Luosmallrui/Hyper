# 笔记图片异步 LLM 标签接口

更新时间：2026-08-15

## 目的

图片上传接口不再同步等待视觉 LLM。客户端先得到可用图片 URL；后台完成标签生成后通过现有 WebSocket 推送结果。Socket 离线或推送失败时，客户端可查询结果兜底。

本方案已撤销腾讯云 `DetectLabelPro` 依赖，统一使用现有 `llm.tag_model` 的视觉模型配置。

## 上线前 SQL

在生产库执行一次：

```sql
ALTER TABLE `image`
  ADD COLUMN `tags` JSON NULL COMMENT '异步 LLM 生成的话题标签',
  ADD COLUMN `tag_status` TINYINT NOT NULL DEFAULT 0 COMMENT '0处理中 1完成 2失败',
  ADD COLUMN `tag_error` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '内部失败原因',
  ADD KEY `idx_tag_status` (`tag_status`);
```

RocketMQ 需要存在 Topic：

```text
HYPER_NOTE_IMAGE_TAG
```

`conn-server` 的 SimpleConsumer 已订阅该 Topic；部署后需要重启 `conn-server`，旧进程不会热更新订阅表达式。

## 1. 上传图片

```http
POST /api/v1/note/upload
Authorization: Bearer <access_token>
Content-Type: multipart/form-data
```

表单字段：`image`。

成功后立即返回，不等待标签：

```json
{
  "code": 200,
  "data": {
    "image_id": "2089000000000000000",
    "url": "https://cdn.hypercn.cn/note/2026/08/15/2089000000000000000.jpg",
    "width": 1080,
    "height": 1440,
    "tags": [],
    "tag_status": "pending"
  }
}
```

约定：

- `image_id` 固定为字符串，避免雪花 ID 在 JavaScript 中丢精度。
- `pending`：后台任务已等待处理。
- `failed`：图片上传成功，但标签任务投递失败；图片仍可正常发帖。
- 前端应立即展示图片；标签区域可显示加载状态，不应阻塞发布流程。

## 2. WebSocket 标签完成事件

客户端保持现有 IM WebSocket 连接。服务端完成任务后推送：

```json
{
  "event": "note.image_tags",
  "payload": {
    "image_id": "2089000000000000000",
    "url": "https://cdn.hypercn.cn/note/2026/08/15/2089000000000000000.jpg",
    "status": "completed",
    "tags": [
      { "id": 12, "name": "音乐现场", "view_count": 0 },
      { "id": 29, "name": "夜生活", "view_count": 0 }
    ]
  },
  "is_self": false
}
```

失败事件：

```json
{
  "event": "note.image_tags",
  "payload": {
    "image_id": "2089000000000000000",
    "url": "https://cdn.hypercn.cn/note/2026/08/15/2089000000000000000.jpg",
    "status": "failed",
    "tags": [],
    "error": "标签生成失败"
  }
}
```

前端按 `image_id` 合并到对应上传图片。拿到 `completed` 后，将 `tags` 直接用于笔记编辑页的话题建议；失败时可隐藏建议或提供手动选题入口。

## 3. 查询标签结果（断线兜底）

```http
GET /api/v1/note/upload/:image_id/tags
Authorization: Bearer <access_token>
```

仅图片所属用户可查询。

```json
{
  "code": 200,
  "data": {
    "image_id": "2089000000000000000",
    "url": "https://cdn.hypercn.cn/note/2026/08/15/2089000000000000000.jpg",
    "status": "completed",
    "tags": [
      { "id": 12, "name": "音乐现场", "view_count": 0 }
    ]
  }
}
```

`status` 枚举：

```text
pending    仍在生成
completed  已完成
failed     生成失败
```

建议前端在上传成功后订阅 Socket；Socket 未连接、页面恢复前台或超过 5 秒仍为 `pending` 时调用本接口补拉一次。

## 后端处理规则

1. `POST /note/upload` 上传 OSS、写入 `image` 后投递 `HYPER_NOTE_IMAGE_TAG`，立即响应。
2. `conn-server` 消费消息，调用 `llm.GenNoteTag`，再把标签写为既有话题并持久化到 `image.tags`。
3. 任务成功、失败都会更新 `tag_status`，不会影响图片上传或笔记发布。
4. 数据库或话题写入失败时 RocketMQ 不确认消息，由 Broker 重投；视觉模型无标签返回时标记失败，不做无限计费重试。
5. 重复消息会复用已持久化标签，只补发 Socket 事件，不重复创建话题。
