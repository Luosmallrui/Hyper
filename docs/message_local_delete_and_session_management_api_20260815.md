# IM 本地删除、清空聊天与移除会话接口

更新时间：2026-08-15

## 语义

三个动作彼此独立，均只影响当前登录用户：

| 动作 | 效果 | 对方是否受影响 |
|---|---|---|
| 删除单条消息 | 当前用户的聊天记录不再返回该消息 | 否 |
| 清空聊天 | 当前用户不再看到当前会话的历史消息；后续新消息仍可见 | 否 |
| 移除会话 | 当前用户的消息列表不再显示该会话；历史消息不删除 | 否 |

删除和清空均为用户维度的软隐藏，不会删除 IM 落库记录，因此不会影响客服、群聊管理和必要的运营审计。移除会话后，只要任一方再发消息，当前用户的会话会重新出现在消息列表。

基础前缀：`/api/v1`。以下接口均需：

```http
Authorization: Bearer <access_token>
```

`session_type`：`1` 单聊，`2` 群聊。单聊 `peer_id` 是对方用户 ID，群聊 `peer_id` 是群 ID。

## 1. 删除单条消息

```http
DELETE /api/v1/message/:message_id
Content-Type: application/json
```

```json
{
  "session_type": 1,
  "peer_id": "35"
}
```

成功响应：

```json
{
  "code": 200,
  "data": {
    "message_id": "2087463119642169344",
    "deleted": true
  }
}
```

- 仅能删除属于指定会话的消息。
- 重复删除幂等成功。
- 删除最新消息时，当前用户会话列表的消息预览会回退到上一条可见消息。

## 2. 清空与某人的聊天

```http
POST /api/v1/message/clear
Content-Type: application/json
```

```json
{
  "session_type": 1,
  "peer_id": "35"
}
```

成功响应：

```json
{
  "code": 200,
  "data": {
    "cleared": true
  }
}
```

清空后重新调用 `GET /api/v1/message/list` 不再返回清空时刻之前的历史；清空后收到或发送的新消息正常显示。会话仍保留在 `GET /api/v1/session/` 列表中，预览可按空消息或“暂无消息”展示。

## 3. 从消息列表移除会话

```http
DELETE /api/v1/session
Content-Type: application/json
```

```json
{
  "session_type": 1,
  "peer_id": "35"
}
```

成功响应：

```json
{
  "code": 200,
  "data": {
    "removed": true
  }
}
```

该操作只移除当前用户的 `GET /api/v1/session/` 列表项，不清空聊天历史。前端若希望“从列表消失且历史也不再显示”，应依次调用“清空聊天”和“移除会话”。

## 前端处理建议

- 删除单条成功后直接从当前消息数组移除该 `message_id`。
- 清空成功后清空当前聊天窗口的消息数组；保留会话卡片。
- 移除成功后从会话列表移除对应的 `session_type + peer_id` 项。
- 群聊删除消息时，如果用户已退群或群已解散，接口分别返回 `403`、`404`。

## 生产迁移 SQL

先执行下列 SQL，再部署包含本功能的后端：

```sql
ALTER TABLE `im_session`
  ADD COLUMN `cleared_at` BIGINT NOT NULL DEFAULT 0 COMMENT '当前用户清空会话历史的时间戳(毫秒)' AFTER `is_mute`,
  ADD COLUMN `last_msg_hidden` TINYINT NOT NULL DEFAULT 0 COMMENT '当前用户是否隐藏最新消息预览' AFTER `cleared_at`;

CREATE TABLE IF NOT EXISTS `im_message_user_deletions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `message_id` BIGINT NOT NULL COMMENT 'IM 消息 ID',
  `session_type` TINYINT NOT NULL COMMENT '1单聊 2群聊',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '执行本地删除的用户 ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_message_user_delete` (`message_id`, `session_type`, `user_id`),
  KEY `idx_message_user_delete_user` (`user_id`, `session_type`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='IM 用户本地删除记录';
```
