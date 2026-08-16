# 主办方、固定场地与活动模型接口

更新时间：2026-08-15

## 目标模型

```text
主办方 organizer（个人、门店或活动组织者）
├── type = venue：一个固定场地资料，固定地址、经纬度、营业时间，常驻地图
└── 可发布多个临时活动 activity（type = party）
```

`venue` 是主办方资料，不是活动。场地主办方和普通活动组织者都可以发布活动；场地主办方发布的活动固定使用该场地已审核通过的地址与坐标。

历史 `activities.type = venue` 仍可查询、订阅和显示，作为只读兼容数据；新前端不得再创建这类记录。

## 生产 SQL

先执行本文件中的场地二审与活动图标字段迁移。若尚未执行地图图标迁移，同时执行 [organizer_marker_icon_api_20260815.md](./organizer_marker_icon_api_20260815.md) 的 `marker_icon` SQL。

```sql
ALTER TABLE `organizers`
  ADD COLUMN `pending_profile_revision` mediumtext NOT NULL COMMENT '场地资料待审核修订 JSON' AFTER `marker_icon`,
  ADD COLUMN `pending_profile_status` tinyint NOT NULL DEFAULT 0 COMMENT '场地资料修订状态: 0无 1审核中 3驳回' AFTER `pending_profile_revision`,
  ADD COLUMN `pending_profile_reason` varchar(500) NOT NULL DEFAULT '' COMMENT '场地资料修订驳回原因' AFTER `pending_profile_status`,
  ADD KEY `idx_organizer_pending_profile_status` (`pending_profile_status`);

ALTER TABLE `activities`
  ADD COLUMN `marker_icon` varchar(255) NOT NULL DEFAULT '' COMMENT '活动地图标记图标 URL' AFTER `name`;
```

不要把历史全部 `activities.type=venue` 自动迁移成 `organizers.type=venue`：一个旧主办方可能存在多条旧场地活动，必须由运营选择哪一条才是其唯一固定场地。历史数据在迁移前不会消失。

## 1. 场地入驻

```http
POST /api/v1/organizer/apply
Authorization: Bearer <access_token>
Content-Type: application/json
```

场地入驻使用 `type: "venue"`，必须提交 `venue_profile`。普通活动组织者使用 `type: "merchant"`（也兼容旧的 `party`），不需要 `venue_profile`。

```json
{
  "name": "Hyper Club",
  "type": "venue",
  "logo": "https://cdn.hypercn.cn/logo.png",
  "marker_icon": "https://cdn.hypercn.cn/marker-icons/jiuba.png",
  "province": "四川省",
  "city": "成都市",
  "district": "武侯区",
  "venue_profile": {
    "cover_image": "https://cdn.hypercn.cn/venue-cover.jpg",
    "gallery": ["https://cdn.hypercn.cn/venue-1.jpg"],
    "description": "场地介绍",
    "business_hours": "19:30-次日02:30",
    "contact_name": "小李",
    "service_phone": "13800000000",
    "address": "成都市武侯区天府三街 1 号",
    "latitude": 30.657,
    "longitude": 104.066,
    "average_spend": 7600
  }
}
```

场地必须提供地址、营业时间以及中国范围内的经纬度。首次审核通过后，资料才会出现在场地列表和地图。

## 2. 场地资料修改与二次审核

```http
PUT /api/v1/organizer/profile
Authorization: Bearer <access_token>
```

请求体沿用现有主办方资料字段。对于已通过的 `type=venue` 主办方，前端应提交完整资料（名称、Logo、省市区、封面/图册、介绍、营业时间、联系人、电话、地址、经纬度、人均及 `marker_icon`）。

- 提交后，公开场地仍展示旧资料，不会下架。
- 管理端通过后，新资料一次性替换旧资料并立即同步地图。
- 管理端驳回后，旧资料继续公开；主办方得到驳回原因后可重新提交。
- 普通 `merchant` 主办方继续沿用原来的直接资料更新行为。

`GET /api/v1/organizer/profile` 现在返回顶层 `type`。当 `type=venue` 时，同时返回 `venue_profile`；保留原有扁平资料字段，兼容旧客户端：

```json
{
	  "type": "venue",
	  "address": "成都市武侯区天府三街 1 号",
	  "latitude": 30.657,
	  "longitude": 104.066,
	  "venue_profile": {
	    "address": "成都市武侯区天府三街 1 号",
	    "latitude": 30.657,
	    "longitude": 104.066,
	    "business_hours": "19:30-次日02:30"
	  },
	  "has_pending_profile_revision": true,
  "pending_profile_reason": "地址证明不清晰",
  "pending_profile_revision": {
    "name": "Hyper Club 新店",
    "address": "成都市武侯区天府三街 2 号",
    "latitude": 30.657,
    "longitude": 104.066,
    "business_hours": "19:30-次日02:30"
  }
}
```

`GET /api/v1/organizer/audit-status` 同步返回 `has_pending_profile_revision` 与 `pending_profile_reason`。

`PUT /api/v1/organizer/profile` 既接受历史扁平字段，也接受下列嵌套格式；场地主办方推荐使用嵌套格式：

```json
{
  "name": "Hyper Club",
  "province": "四川省",
  "city": "成都市",
  "district": "武侯区",
  "venue_profile": {
    "address": "成都市武侯区天府三街 1 号",
    "latitude": 30.657,
    "longitude": 104.066,
    "business_hours": "19:30-次日02:30"
  }
}
```

## 3. 管理端审核

沿用原接口，不新增审核入口：

```http
GET /api/v1/admin/organizers?status=1&page=1&pageSize=20
GET /api/v1/admin/organizers/:id
PUT /api/v1/admin/organizers/:id/audit
```

审核列表新增字段：

```json
{
  "id": 9,
  "status": 2,
  "audit_kind": "profile_revision",
  "has_pending_profile_revision": true,
  "pending_profile_reason": ""
}
```

- `audit_kind = initial`：首次入驻审核，`status=1`。
- `audit_kind = profile_revision`：已通过场地的资料二审，主办方 `status` 保持 `2`。
- 详情接口返回 `pending_profile_revision` 供管理员比对新旧资料。

审核请求不变：

```json
{ "status": 2 }
```

通过使用 `status=2`；驳回使用 `status=3` 并传 `reject_reason`。

## 4. 发布活动

```http
POST /api/v1/activity/create
Authorization: Bearer <access_token>
```

新活动只支持：

```json
{
  "type": "party",
  "step": 1,
  "name": "今晚电音派对",
  "marker_icon": "https://cdn.hypercn.cn/marker-icons/dianyin.png"
}
```

`marker_icon` 为活动自身的地图标记图标，使用与入驻相同的前端图标选择器；仅接受 `https://cdn.hypercn.cn/...`。
它不会修改主办方 Logo 或固定场地图标。地图优先级为：活动 `marker_icon` > 主办方 `marker_icon` > 默认派对图标。已上架活动修改该字段会走活动二次审核。

新请求传 `type: "venue"` 会返回：

```text
新场地请在入驻申请中选择 venue 并填写固定资料，活动发布仅支持 party
```

对 `type=venue` 主办方，后端会强制以当前已审核的场地省市区、地址、经纬度保存活动，前端无需再次让用户选择场地位置；活动自己的开始/结束时间、票券和内容仍按原活动流程填写。

## 5. 地图与场地详情

```http
GET /api/v1/map/markers?source=all
GET /api/v1/map/markers?source=venue
GET /api/v1/map/markers?source=activity
GET /api/v1/venues
GET /api/v1/venues/:organizer_id
```

固定场地 marker：

```json
{
  "id": "venue-9",
  "source": "venue",
  "source_id": 9,
  "detail_type": "venue",
  "detail_url": "/api/v1/venues/9",
  "title": "Hyper Club",
  "type": "venue",
  "address": "成都市武侯区天府三街 1 号",
  "lat": 30.657,
  "lng": 104.066,
  "icon": "https://cdn.hypercn.cn/marker-icons/jiuba.png",
  "follow_target_type": "venue",
  "follow_target_id": 9
}
```

临时活动 marker 保持：`id = activity-{activity_id}`、`source_id = activity_id`、详情跳转 `/api/v1/activity/:id`。

前端必须按 `source` 路由，不能把 `venue` 的 `source_id` 当作 `activity_id`。

## 6. 主办方主页

```http
GET /api/v1/organizers/:id
```

返回中的 `type` 为主办方类型。`type=venue` 时：

- 顶部资料使用主办方 Logo、名称和场地资料，而不是申请人的用户头像/昵称。
- `venues.list` 包含该主办方自己的固定场地（`id=organizer_id`）。
- `activities.list` 包含该主办方发布的全部已上架临时活动。

## 前端改动清单

1. 入驻向导保留“场地 / 活动组织者”选择；选场地时展示固定地址、经纬度、营业时间和场地资料，不展示票券。
2. 发布向导移除“创建场地”分支，只发布活动；场地主办方不展示地址选择控件，地址由后端固定。
3. 地图卡片按 `source=venue|activity` 打开相应详情；不依赖 `venue.activity_id`。
4. 场地资料提交后读取 `has_pending_profile_revision`，展示“资料审核中”；驳回时展示 `pending_profile_reason`，公开页继续用当前资料。
5. 管理端审核列表对 `audit_kind=profile_revision` 显示“场地资料修改审核”，详情页同时展示 `pending_profile_revision` 和当前公开资料。
6. 发布活动页增加“地图图标”选择器，提交 `marker_icon`；编辑活动时用 `GET /api/v1/activity/:id` 返回的 `marker_icon` 回填。
