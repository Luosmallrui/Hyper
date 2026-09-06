# 地图封面字段（2026-09-03）

## 新增字段

| 表 | 字段 | 说明 |
|---|---|---|
| `activities` | `poster_map` varchar(255) | 活动地图封面，在现有 `poster_list/poster_detail/poster_wechat/poster_long` 基础上增量新增 |
| `organizer_profiles` | `map_cover` varchar(255) | 场地地图封面，在现有 `cover_image` 基础上增量新增 |

DDL 参考（`config/table.sql` 末尾注释）：

```sql
ALTER TABLE `activities` ADD COLUMN `poster_map` varchar(255) NOT NULL DEFAULT '' COMMENT '地图封面' AFTER `poster_wechat`;
ALTER TABLE `organizer_profiles` ADD COLUMN `map_cover` varchar(255) NOT NULL DEFAULT '' COMMENT '地图封面' AFTER `cover_image`;
```

## 受影响的接口（均为增量加字段，现有字段不变）

- 活动创建/更新：`ActivityCreateRequest` 新增可选 `poster_map`，写入 `activities.poster_map`。
- 活动详情：响应增量返回 `poster_map`（内嵌 `models.Activity` 自动透出）。
- 场地资料读取/保存（商家端 `OrganizerProfileRequest`/`OrganizerProfileResponse` 及管理端 `getAdminVenueProfile`/`upsertOrganizerVenueProfile`）：增量读写 `map_cover`；历史场地活动回填（`fillAdminVenueProfileFromLegacy`）时 `map_cover` 为空则取旧活动的 `poster_map`。
- `/v1/map/markers` 的 `cover_image` 取值优先级：
  - 场地 marker：`map_cover → cover_image → logo`
  - 活动 marker：`poster_map → poster_list`
  - 未配置地图封面时行为与旧数据完全一致。

## 上传 type 白名单

`buildUploadKey` 新增两个允许的上传类型：`poster_map`（活动地图封面）、`venue_map_cover`（场地地图封面）。
