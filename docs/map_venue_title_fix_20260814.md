# 地图场地名称取值修复

更新时间：2026-08-14

## 问题

新场地已迁移为 `activities.type = venue`。但地图 marker 曾沿用旧“商家即场地”逻辑，将场地 marker 的 `title` 覆盖为 `organizers.name`，导致场地活动名与地图名称不一致。

例如：

| 字段 | 值 |
| --- | --- |
| `activities.id` | `15` |
| `activities.name` | `一个场地` |
| `organizers.id` | `9` |
| `organizers.name` | `吗哈哈哈` |

旧响应错误地把 `title` 返回为“吗哈哈哈”。

## 修复后的契约

`GET /api/v1/map/markers` 中，所有新活动和新场地统一约定：

| 字段 | 场地 `source=venue` | 派对 `source=activity` |
| --- | --- | --- |
| `title` | `activities.name`，场地名称 | `activities.name`，活动名称 |
| `user` / `username` | `organizers.name`，商家名称 | `organizers.name`，商家名称 |
| `activity_id` | 承载场地资料的活动 ID | 活动 ID |
| `source_id` | 主办方 ID | 活动 ID |

修复后示例：

```json
{
  "id": "venue-9",
  "source": "venue",
  "source_id": 9,
  "activity_id": 15,
  "title": "一个场地",
  "user": "吗哈哈哈",
  "username": "吗哈哈哈"
}
```

前端展示场地卡片标题使用 `title`；商家署名使用 `user` 或 `username`。不要再用商家名覆盖场地名称。
