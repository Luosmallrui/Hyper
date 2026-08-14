# 地图 Marker 补充活动 ID

更新时间：2026-08-14

## 背景

新场地 marker 的 `source_id` 是场地主办方/商家 ID，用于场地详情和关注：

```http
GET /api/v1/venues/:source_id
```

它不是承载场地资料的 `activities.id`，前端不能从 `source_id` 推导活动 ID。现补充 `activity_id`，支持前端跳转统一活动页或活动编辑页。

## 接口

```http
GET /api/v1/map/markers?source=all&limit=200
```

所有来自新 `activities` 表的地图 marker 都返回 `activity_id`。

### 场地示例

```json
{
  "id": "venue-9",
  "source": "venue",
  "source_id": 9,
  "activity_id": 15,
  "detail_type": "venue",
  "detail_url": "/api/v1/venues/9",
  "title": "一个场地",
  "status": 3
}
```

字段含义：

| 字段 | 含义 | 本例 |
| --- | --- | --- |
| `source_id` | 场地主办方 ID；用于场地详情、场地关注/订阅。 | `9` |
| `activity_id` | 承载场地展示资料的活动记录 ID；用于活动页、商家活动编辑。 | `15` |
| `detail_url` | 用户侧首选详情地址。 | `/api/v1/venues/9` |

前端若需跳转活动页面，使用：

```text
/pages/activity/index?id=${marker.activity_id}
```

用户查看场地资料时仍优先使用：

```text
${marker.detail_url}
```

### 派对/活动示例

```json
{
  "id": "activity-22",
  "source": "activity",
  "source_id": 22,
  "activity_id": 22,
  "detail_type": "activity",
  "detail_url": "/api/v1/activity/22"
}
```

旧 `parties` 表返回的兼容 marker 没有对应新活动记录，`activity_id` 不返回。

## 关注规则不变

不要使用 `activity_id` 推导关注目标。仍严格使用响应内：

```json
{
  "follow_target_type": "venue",
  "follow_target_id": 9
}
```

场地关注目标是主办方 ID；活动页跳转 ID 与关注目标 ID 是两个不同概念。
