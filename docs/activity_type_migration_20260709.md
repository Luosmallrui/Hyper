# 场地/派对类型迁移说明

更新时间：2026-07-09

## 背景

场地/派对的区分从“商家入驻申请”迁移到“活动创建”。入驻只表示用户申请成为主办方/商家，不再决定该账号是场地还是派对。

## 核心规则

- 入驻申请 `/api/v1/organizer/apply` 不再要求 `type`。
- `organizers.type` 仅作为历史兼容字段保留，后端默认写入 `merchant`。
- 活动创建 `/api/v1/activity/create` 新增 `type`：
  - `party`：派对，默认值。
  - `venue`：场地。
- C 端场地列表、场地详情、场地订阅聚合不再按 `organizers.type=venue` 判断，而是要求主办方名下存在已上架的 `activities.type=venue` 活动。

## 入驻申请

```http
POST /api/v1/organizer/apply
Authorization: Bearer <access_token>
Content-Type: application/json
```

请求：

```json
{
  "name": "Hyper Club",
  "logo": "https://cdn.xxx/logo.png",
  "province": "四川省",
  "city": "成都市",
  "district": "武侯区"
}
```

兼容说明：

- 旧前端如果仍传 `type: "venue"` 或 `type: "merchant"`，后端会兼容接收。
- 后端不会再根据入驻 `type` 判断场地/派对。

## 活动创建

```http
POST /api/v1/activity/create
Authorization: Bearer <access_token>
Content-Type: application/json
```

派对活动：

```json
{
  "step": 1,
  "type": "party",
  "name": "周末电音派对",
  "start_time": "2026-07-18 20:00:00",
  "end_time": "2026-07-18 23:00:00"
}
```

场地活动：

```json
{
  "step": 1,
  "type": "venue",
  "name": "SWING 鸡尾酒吧",
  "start_time": "2026-07-18 19:00:00",
  "end_time": "2026-07-19 02:00:00"
}
```

字段说明：

| 字段 | 必填 | 说明 |
|---|---|---|
| type | 否 | `party` / `venue`，不传默认 `party` |

错误：

```json
{
  "code": 500,
  "msg": "活动类型无效，仅支持 party 或 venue"
}
```

## 地图接口

```http
GET /api/v1/map/markers?source=all
GET /api/v1/map/markers?source=party
GET /api/v1/map/markers?source=venue
```

规则：

| source | 返回 |
|---|---|
| all | 所有已上架新活动 |
| activity | 所有已上架新活动 |
| party | `activities.type=party` 的已上架活动 |
| venue | `activities.type=venue` 的已上架活动 |
| merchant | 兼容保留，返回空列表 |

返回项中的 `type` 会返回 `party` 或 `venue`。

## C 端场地接口

以下接口仍保留原路径：

```http
GET /api/v1/venues
GET /api/v1/venues/:id
GET /api/v1/venues/:id/notes
POST /api/v1/venues/:id/follow
DELETE /api/v1/venues/:id/follow
POST /api/v1/venues/:id/subscribe
DELETE /api/v1/venues/:id/subscribe
GET /api/v1/subscriptions?type=venue
```

新的可见条件：

- `organizers.status=2`
- `organizers.enabled=1`
- 存在至少一个 `activities.organizer_id=organizers.id AND activities.type='venue' AND activities.status=3`

## 数据库变更

已有库需要执行：

```sql
ALTER TABLE `activities`
  ADD COLUMN `type` varchar(20) NOT NULL DEFAULT 'party' COMMENT '活动类型: party派对 venue场地' AFTER `organizer_id`;

ALTER TABLE `activities`
  ADD INDEX `idx_activity_type` (`type`);

ALTER TABLE `organizers`
  MODIFY COLUMN `type` varchar(20) NOT NULL DEFAULT 'merchant' COMMENT '兼容字段: 入驻不再区分场地/派对，场地/派对由 activities.type 决定';
```

历史数据处理建议：

```sql
UPDATE `activities`
SET `type` = 'party'
WHERE `type` = '' OR `type` IS NULL;
```

如果已有某些活动需要作为场地展示，按活动 ID 单独更新：

```sql
UPDATE `activities`
SET `type` = 'venue'
WHERE `id` IN (10, 11);
```
