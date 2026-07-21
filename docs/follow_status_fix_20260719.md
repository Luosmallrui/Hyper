# 关注状态不持久问题修复说明（2026-07-19）

对应反馈：`follow_status_bug_report_20260718.md`（小程序端地图卡片关注后返回消失）。

## 结论

**接口契约不变，前端无需改动。** 关注/取关写库逻辑本身没有问题，问题出在两个只读接口没有把 `is_follow` 算出来，本次修复的是这两处读取逻辑。

## 根因

1. `GET /api/v1/map/markers`：返回的每个 marker 里 `is_follow` 字段之前从未被赋值，恒为 `false`，与当前用户是否真实关注了对方无关。
2. `GET /api/v1/activity/:id`：响应里之前**没有 `is_follow` 字段**（只有 `is_subscribe`）。

`POST /api/v1/follow/follow`、`POST /api/v1/follow/unfollow` 本身写库正常，`GET /api/v1/follow/list` 一直是对的。

## 修复内容

- `GET /api/v1/map/markers`：`is_follow` 现在按当前登录用户（token 解析出的 user_id）与 marker 归属的主办方账号（`organizer.user_id`）之间的真实关注关系计算，批量查询，不影响接口响应时间。
- `GET /api/v1/activity/:id`：响应新增 `is_follow` 字段，语义与 `is_subscribe` 一致，按当前用户与该活动主办方账号的关注关系计算。未登录（无 token）时两者均为 `false`。

## 验证方式

1. 小程序端登录用户对某活动/场地对应的商家账号执行关注。
2. 重新拉取 `GET /api/v1/map/markers`，对应 marker 的 `is_follow` 应为 `true`。
3. 打开该活动详情 `GET /api/v1/activity/:id`，`is_follow` 应为 `true`。
4. 取消关注后，两个接口的 `is_follow` 应恢复为 `false`。

## 备注

`getPartyMarkers`（基于旧 `organizers.type` 商家表的地图数据源）目前是不可达的遗留代码，`source=merchant` 按文档约定返回空列表属于既定行为（见 `activity_type_migration_20260709.md`），本次未改动，与本问题无关。
