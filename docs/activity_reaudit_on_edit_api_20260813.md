# 已审核活动/场地修改后的二次审核

更新时间：2026-08-13

## 目的

主办方已审核通过并上架的活动或场地，后续修改名称、海报、地址、介绍、时间、标签、场地经营时间或票券信息后，必须重新进入平台审核。未经二次审核的最新内容不能继续在用户侧公开展示。

本次复用既有活动状态和管理员审核接口；为了区分首次审核和二次审核，需要新增一个轻量字段。

## 数据库迁移

生产部署前执行一次：

```sql
ALTER TABLE `activities`
  ADD COLUMN `audit_type` varchar(20) NOT NULL DEFAULT 'initial'
  COMMENT '审核类型: initial首次审核 re_audit修改后二审'
  AFTER `status`;

ALTER TABLE `activities`
  ADD INDEX `idx_activity_audit_type` (`status`, `audit_type`);
```

已有活动会自动以默认值 `initial` 补齐，不需要单独更新历史数据。

## 状态约定

| 值 | 状态 | 说明 |
| --- | --- | --- |
| `0` | 草稿 | 未提交审核。 |
| `1` | 待审核 | 首次提交，或已上架内容被主办方修改后的二次审核。 |
| `2` | 审核中 | 预留状态。 |
| `3` | 已上架 | 管理员审核通过，允许在地图、搜索、详情、商家主页等公开位置展示。 |
| `4` | 审核未通过 | 返回 `reject_reason`，商家修改后重新提交。 |

`audit_type`：

| 值 | 含义 |
| --- | --- |
| `initial` | 首次提交审核；被驳回后再次提交仍保持该值。 |
| `re_audit` | 已上架内容被主办方修改后产生的二次审核；被驳回后再次提交仍保持该值。 |

## 商家端编辑

继续使用现有接口，无新增端点：

```http
POST /api/v1/activity/create
Authorization: Bearer <organizer_access_token>
Content-Type: application/json
```

编辑已有内容时必须传原活动 ID：

```json
{
  "activity_id": 22,
  "step": 3,
  "name": "A WEIGHT OF SOUND 2026",
  "description": "更新后的活动介绍",
  "poster_list": "https://cdn.example.com/activity-list.jpg",
  "start_time": "2026-09-01T20:00:00+08:00",
  "end_time": "2026-09-02T02:00:00+08:00",
  "tag_ids": [1, 2]
}
```

活动票券编辑使用专用接口：

```http
POST /api/v1/activity/:id/ticket-specs
DELETE /api/v1/ticket-spec/:id
```

`POST /api/v1/activity/:id/ticket-specs` 的请求体为：

```json
{
  "specs": [
    {
      "id": "101",
      "name": "早鸟票",
      "description": "限量",
      "is_enabled": 1,
      "sale_start": "2026-09-01T10:00:00+08:00",
      "sale_end": "2026-09-10T18:00:00+08:00",
      "price": 9900,
      "stock": 100,
      "purchase_limit": 2,
      "max_attendees": 1
    },
    {
      "name": "正价票",
      "price": 12900,
      "stock": 200,
      "is_enabled": 1
    }
  ]
}
```

- 带 `id`：更新该票券。
- 不带 `id`：新增票券。
- 该接口是**批量 upsert，不是全量替换**；请求数组中缺失的旧票券不会自动删除。
- 删除票券必须逐条调用 `DELETE /api/v1/ticket-spec/:id`。
- 以上任一成功写操作均会触发已上架活动的二次审核。

### 票券更新严格契约

- `id` 在 HTTP JSON 中统一为字符串，使用 `GET /api/v1/activity/:id` 返回的 `ticket_specs[].id` 原样传回；后端内部按 `int64` 处理。不得自行生成、转为 JS number 或截断。
- 更新项必须带完整票券字段，至少 `id`、`name`、`price`、`stock`、`is_enabled`、`sale_start`、`sale_end`、`purchase_limit`、`max_attendees`；后端按完整对象覆盖可编辑字段，不支持票券字段级 PATCH。
- `name` 和 `price` 必填；`stock` 不能小于 `0`。`purchase_limit <= 0`、`max_attendees <= 0` 时后端各回退为 `1`。
- `sale_start` / `sale_end` 可为空字符串，或使用 RFC3339 / `YYYY-MM-DDTHH:mm` / `YYYY-MM-DD HH:mm:ss` / `YYYY-MM-DD HH:mm`。
- `is_enabled` 传 `0` 可以禁用票券，`price: 0` 和 `stock: 0` 会按零值实际保存，不会因 GORM 零值规则被忽略。
- `id` 不属于当前活动、活动不属于当前主办方或票券不存在时，接口返回失败；不得把其他活动票券 ID 用作新增 ID。

`POST /api/v1/activity/create` 的 `step=4` / `ticket_specs` 是旧向导兼容入口，同样只做新增和更新，不能用于删除。编辑页不得把它当作票券全量保存接口。

### 自动二审规则

当活动当前 `status=3` 时，下列任一商家修改成功，后端自动将活动改为 `status=1`，并清空历史 `reject_reason`：

- 活动或场地基础资料：名称、分享标题、介绍、实名/未成年人配置、行政区、地址、坐标、各类海报、资质。
- 派对活动时间。
- 场地经营时间。
- 活动或场地标签，包括清空全部标签。
- 派对票券的新增、编辑、删除。

返回仍沿用创建/保存接口的响应：

```json
{
  "code": 200,
  "data": {
    "activity_id": 22
  }
}
```

保存后前端重新请求下列接口即可获得 `status=1`：

```http
GET /api/v1/activity/my-list?page=1&size=50
GET /api/v1/activity/:id
```

`GET /api/v1/activity/my-list` 的列表项和 `GET /api/v1/activity/:id` 的详情都会返回：

```json
{
  "status": 1,
  "audit_type": "re_audit"
}
```

商家端文案规则：

- `status=1 && audit_type=re_audit`：`修改审核中`
- `status=1 && audit_type=initial`：`审核中`
- `status=4`：`审核未通过`
- `status=3`：`已上架`

活动在二审期间仍应出现在商家自己的活动列表中，并禁用重复提交按钮。

## 编辑回填字段

商家编辑页可直接调用：

```http
GET /api/v1/activity/:id
Authorization: Bearer <organizer_access_token>
```

该接口返回可回填的完整活动字段，包括：

- 基础字段：`type`、`name`、`share_title`、`description`、`province`、`city`、`district`、`address`、`latitude`、`longitude`。
- 活动配置：`start_time`、`end_time`、`real_name_mode`、`minor_check`、`qualification_doc`。
- 场地经营时间：`business_hours`（仅 `type=venue`，来自该商家资料）。
- 图片字段：`poster_detail`、`poster_long`、`poster_list`、`poster_wechat`。
- 标签：`tag_ids`、`tags`。
- 票券：`ticket_specs`（场地固定返回空数组）。
- 审核：`status`、`audit_type`、`reject_reason`。

前端应以详情原值初始化编辑表单，并以字段级 diff 只提交用户确实修改的字段；后端目前以“字段是否出现在请求体”判断修改，不比较新旧值。空字符串、空数组和 `null` 不应作为“未修改”的替代值。若用户主动清空标签，应提交 `tag_ids: []`，后端会清空标签并触发二次审核。

## 用户侧可见性

二审中的内容不再是公开内容：

- 不出现在首页地图、搜索、公开活动列表、场地列表和公开商家主页。
- 非主办方访问 `GET /api/v1/activity/:id` 返回 `404`。
- 主办方本人可继续访问详情和编辑。
- 平台主动隐藏、且用户已有历史订单的活动，订单详情仍可读取历史活动信息；该兼容逻辑不受影响。

## 管理端审核

管理员继续使用原接口处理首次审核和二次审核：

```http
PUT /api/v1/admin/activities/:id/audit
Authorization: Bearer <admin_access_token>
Content-Type: application/json
```

通过：

```json
{
  "status": 3
}
```

驳回：

```json
{
  "status": 4,
  "reject_reason": "请补充场地封面与准确地址"
}
```

管理员活动列表使用既有筛选查看待复审内容：

```http
GET /api/v1/admin/activities?page=1&pageSize=20&status=1
```

只筛选二次审核：

```http
GET /api/v1/admin/activities?page=1&pageSize=20&status=1&audit_type=re_audit
```

审核通过后活动恢复 `status=3` 并重新出现在公开位置；驳回后商家读取 `reject_reason`，修改资料后调用：

```http
POST /api/v1/activity/:id/submit-audit
```

再次进入待审核状态。
