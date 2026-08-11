# 管理端活动下架与恢复接口

更新时间：2026-08-11

## 设计原则

管理端对新活动表 `activities` 中的派对（`type=party`）和场地（`type=venue`）执行“删除”时，默认改为下架隐藏，不物理删除数据。

- 已有订单、退款、核销记录、财务流水和统计数据保留。
- 客户端地图、搜索、公开详情、订阅列表、场地列表、主办方公开主页不再展示隐藏数据。
- 已隐藏活动不可继续购票、订阅或关注。
- 管理端和所属主办方仍可查询历史数据；管理端可恢复展示。
- 隐藏状态独立于审核状态。恢复后只有审核状态为 `3` 的数据才会重新出现在客户端。

旧 `parties` 表也提供兼容删除接口：

```http
DELETE /api/v1/admin/parties/:id
Authorization: Bearer <admin_access_token>
```

该接口将既有 `parties.status` 更新为 `offline`，同样不物理删除。恢复时继续调用已有的 `PUT /api/v1/admin/parties/:id/status` 并提交 `{ "status": "active" }`。本文件主体接口用于当前活动体系的 `activities` 表。

## 数据库迁移

部署代码前执行以下 SQL：

```sql
ALTER TABLE `activities`
  ADD COLUMN `is_hidden` tinyint NOT NULL DEFAULT 0 COMMENT '0公开 1平台下架隐藏' AFTER `reject_reason`,
  ADD COLUMN `hidden_at` datetime NULL COMMENT '平台下架时间' AFTER `is_hidden`,
  ADD COLUMN `hidden_reason` varchar(500) NOT NULL DEFAULT '' COMMENT '平台下架原因' AFTER `hidden_at`,
  ADD INDEX `idx_activity_public` (`status`, `is_hidden`);
```

如某列或索引已存在，请跳过对应语句，避免重复执行报错。

## 下架隐藏（管理端默认删除）

```http
DELETE /api/v1/admin/activities/:id?reason=内容整改
Authorization: Bearer <admin_access_token>
```

`reason` 可选，建议在违规下架、内容整改等场景传入，便于主办方和管理员日志追溯。

成功响应：

```json
{
  "code": 200,
  "data": {
    "id": 10,
    "is_hidden": true,
    "hidden_at": "2026-08-11T16:00:00+08:00",
    "hidden_reason": "内容整改"
  }
}
```

此接口不会删除活动、票种、订单、退款、核销或财务数据。

## 恢复或再次隐藏

```http
PATCH /api/v1/admin/activities/:id/visibility
Authorization: Bearer <admin_access_token>
Content-Type: application/json
```

恢复展示：

```json
{
  "visible": true
}
```

再次隐藏：

```json
{
  "visible": false,
  "reason": "临时下架"
}
```

成功响应：

```json
{
  "code": 200,
  "data": {
    "id": 10,
    "is_hidden": false,
    "hidden_at": null,
    "hidden_reason": ""
  }
}
```

## 管理端列表字段与筛选

```http
GET /api/v1/admin/activities?page=1&pageSize=20&is_hidden=1
Authorization: Bearer <admin_access_token>
```

`is_hidden` 可选：

- `0`：仅公开展示的数据。
- `1`：仅已下架隐藏的数据。
- 不传：全部非草稿活动，便于管理端统一管理和恢复。

活动列表新增字段：

```json
{
  "is_hidden": 1,
  "hidden_at": "2026-08-11 16:00:00",
  "hidden_reason": "内容整改"
}
```

## 前端接入约定

1. 管理端活动删除按钮调用 `DELETE /api/v1/admin/activities/:id`，按钮文案建议为“下架/隐藏”，不要提示为永久删除。
2. 已下架记录显示“恢复展示”按钮，调用 `PATCH /api/v1/admin/activities/:id/visibility` 并提交 `{ "visible": true }`。
3. 下架前如需填写原因，作为 query 参数 `reason` 传入删除接口，或使用 visibility 接口提交 `reason`。
4. 对审核未通过、待审核、草稿的数据，“恢复展示”不等于审核通过；仍必须先完成审核并使 `status=3`。
