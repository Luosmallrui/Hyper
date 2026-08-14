# 内容关注允许关注自己

更新时间：2026-08-14

## 规则

活动、场地、主办方和旧派对的内容关注允许内容拥有者本人操作。接口保持不变：

```http
POST /api/v1/follow/follow
Content-Type: application/json

{
  "target_type": "venue",
  "target_id": 9
}
```

或使用场地专用接口：

```http
POST /api/v1/venues/9/follow
```

重复关注幂等成功。关注成功后内容的 `is_follow` 为 `true`，`follow_count` 包含本人这一次关注。

## 边界

- 仅放开对象内容关注：`activity`、`venue`、`organizer`、`party`。
- 用户关系关注仍禁止关注自己，旧接口 `POST /api/v1/follow/follow` 仅传 `user_id` 时维持原有保护。
- 目标仍必须存在并处于可公开关注状态；下架、隐藏、未审核或停用内容不能关注。
