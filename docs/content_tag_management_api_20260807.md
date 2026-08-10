# 活动、场地与派对标签管理接口

更新时间：2026-08-07

## 1. 设计与权限

标签定义由平台管理端统一维护，类型固定为 `coupon_tag`。标签可绑定到：

| 目标 | 关联目标 | 说明 |
|---|---|---|
| 活动 | `activity` | 新票务活动，`type=party` 时使用活动标签 |
| 场地 | `venue` | 新场地对应主办方 `organizer_id`，所有场地资料、场地详情及场地型活动共享标签 |
| 旧派对 | `party` | 兼容旧 `parties` 表数据 |

- 只有管理端管理员能新建、修改、停用、删除或分配标签。
- 主办方和小程序客户端只读取标签与筛选结果，不能自行标记优惠。
- 标签 ID 是数据库 `admin_categories.id`，前端不得写死 `1/2/4` 等数值。
- 停用标签不会删除历史绑定，但不会出现在客户端标签列表、详情或筛选结果中；删除标签会同时删除其全部绑定关系。

## 2. 客户端读取标签

```http
GET /api/v1/tags
GET /api/v1/merchant/tags
```

两个接口等价，均匿名可调用，只返回已启用的 `coupon_tag` 标签，按 `sort ASC, id ASC` 排序。

成功响应：

```json
{
  "code": 200,
  "msg": "ok",
  "data": [
    {
      "id": 101,
      "name": "积分立减",
      "image": "https://cdn.hypercn.cn/tags/points.png",
      "value": "",
      "sort": 1
    },
    {
      "id": 102,
      "name": "新人优惠",
      "image": "",
      "value": "",
      "sort": 2
    }
  ]
}
```

## 3. 用户侧筛选

标签参数支持 `tag_ids` 或兼容别名 `tags`，使用英文逗号分隔标签 ID。多标签是“同时满足”关系。

```http
GET /api/v1/map/markers?source=all&tag_ids=101,102
GET /api/v1/venues?tag_ids=101,102&page=1&size=20
GET /api/v1/merchant/list?tag_ids=101,102&page=1&pageSize=20
```

- 地图：派对型活动按活动标签筛选；场地型活动按场地标签筛选。
- 场地列表：按场地标签筛选。
- 旧派对列表：按旧派对标签筛选。
- 不传标签参数时，不影响既有关键字、区域、距离及分类筛选。
- 格式错误（如 `tag_ids=abc` 或 `0`）返回 `code: 400`；已删除或停用的标签 ID 会得到空列表。

## 4. 客户端响应字段

以下接口返回统一字段：

```http
GET /api/v1/map/markers
GET /api/v1/venues
GET /api/v1/venues/:id
GET /api/v1/merchant/list
GET /api/v1/merchant/:id
GET /api/v1/activity/:id
GET /api/v1/activity/my-list
GET /api/v1/activity/search
```

示例：

```json
{
  "tag_ids": [101, 102],
  "tags": [
    {
      "id": 101,
      "name": "积分立减",
      "image": "https://cdn.hypercn.cn/tags/points.png",
      "value": "",
      "sort": 1
    },
    {
      "id": 102,
      "name": "新人优惠",
      "image": "",
      "value": "",
      "sort": 2
    }
  ]
}
```

地图和旧派对列表还保留 `discount_tags: ["积分立减", "新人优惠"]` 作为旧客户端兼容字段；新前端应优先使用 `tags`。

## 5. 管理端标签配置

现有管理端分类接口直接用于维护标签，管理员 Token 必填。

```http
GET    /api/v1/admin/categories?type=coupon_tag&page=1&pageSize=20
POST   /api/v1/admin/categories
PUT    /api/v1/admin/categories/:id
DELETE /api/v1/admin/categories/:id
```

创建请求：

```json
{
  "type": "coupon_tag",
  "name": "积分立减",
  "image": "https://cdn.hypercn.cn/tags/points.png",
  "value": "",
  "sort": 1,
  "status": 1
}
```

约定：

- 同一 `coupon_tag` 下的标签名称不能重复。
- `status=1` 启用，`status=0` 停用。
- `image`、`value` 均可为空；`value` 可用于管理端自定义备注或外部配置编码，不参与筛选。

## 6. 管理端分配标签

```http
PUT /api/v1/admin/activities/:id/tags
PUT /api/v1/admin/venues/:id/tags
PUT /api/v1/admin/organizers/:id/tags
PUT /api/v1/admin/parties/:id/tags
Authorization: Bearer <admin_access_token>
Content-Type: application/json
```

`venues/:id/tags` 与 `organizers/:id/tags` 等价，`:id` 都是场地主办方 ID。

请求：

```json
{
  "tag_ids": [101, 102]
}
```

成功响应：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "success": true,
    "tag_ids": [101, 102]
  }
}
```

规则：

- 该请求为全量覆盖；传 `[]` 表示清空目标全部标签。
- 只能绑定存在且启用的 `coupon_tag`。
- `PUT /admin/activities/:id/tags` 遇到 `type=venue` 的活动时，会自动绑定到其主办方场地，确保活动详情、场地详情和地图展示一致。
- 每次分配都会记录管理员操作日志。

管理端详情已返回 `tag_ids` 与 `tags`，用于编辑时回显：

```http
GET /api/v1/admin/activities/:id
GET /api/v1/admin/organizers/:id
GET /api/v1/admin/parties/:id
```
