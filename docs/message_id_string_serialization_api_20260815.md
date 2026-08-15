# IM 消息 ID 字符串化修复

更新时间：2026-08-15

## 修复内容

`GET /api/v1/message/list` 的消息 `id` 已统一改为 JSON 字符串，不再以 JSON number 返回。Snowflake ID 超过 JavaScript 安全整数范围，前端无需也不得使用 `Number()` 转换。

同时，消息扩展字段中已知的雪花 ID 统一按字符串返回：

- `ext.note_id`
- `ext.activity_id`
- `ext.note.id`
- `ext.party.id`
- `ext.activity.id`

## 返回示例

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": "2085200666392793000",
        "sender_id": 35,
        "msg_type": 8,
        "ext": {
          "note_id": "2087463119642169344",
          "note": {
            "id": "2087463119642169344"
          }
        }
      }
    ]
  }
}
```

## 前端影响

`MessageItem.id` 保持字符串类型即可，无需改接口路径。删除消息时直接拼接原值：

```text
DELETE /api/v1/message/2085200666392793000
```

不要对消息 ID、帖子 ID、活动卡片 ID 执行 `Number(id)`、`parseInt(id)` 或任何数值运算。
