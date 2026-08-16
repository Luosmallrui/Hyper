# 主办方地图图标接口

更新时间：2026-08-15

## 约定

地图图标由前端维护图标 key、中文名和 CDN URL 的映射。后端仅保存前端提交的 `marker_icon` URL，不解析拼音 key，也不维护图标字典。

仅接受 `https://cdn.hypercn.cn/...` 地址，长度不超过 255；不合法地址会返回参数错误。

## 生产 SQL

已有生产库执行一次：

```sql
ALTER TABLE `organizers`
  ADD COLUMN `marker_icon` varchar(255) NOT NULL DEFAULT '' COMMENT '地图标记图标 URL' AFTER `logo`;
```

旧主办方字段为空时，地图继续使用既有默认图标：场地 `jiuba.png`、派对 `party.png`。

## 入驻申请

```http
POST /api/v1/organizer/apply
Authorization: Bearer <access_token>
```

新增可选字段：

```json
{
  "name": "Hyper Club",
  "logo": "https://cdn.hypercn.cn/logo.png",
  "marker_icon": "https://cdn.hypercn.cn/marker-icons/qiche.png",
  "province": "四川省",
  "city": "成都市",
  "district": "武侯区"
}
```

## 审核状态

```http
GET /api/v1/organizer/audit-status
Authorization: Bearer <access_token>
```

响应新增：

```json
{
  "code": 200,
  "data": {
    "organizer_id": 9,
    "status": 2,
    "marker_icon": "https://cdn.hypercn.cn/marker-icons/qiche.png"
  }
}
```

## 主办方资料更新

```http
PUT /api/v1/organizer/profile
Authorization: Bearer <access_token>
```

请求增加：

```json
{
  "marker_icon": "https://cdn.hypercn.cn/marker-icons/qiche.png"
}
```

`GET /api/v1/organizer/profile`、`GET /api/v1/organizer/info` 均返回 `marker_icon`，用于编辑回填。

## 地图 markers

```http
GET /api/v1/map/markers?source=all
```

新活动和固定场地的 `icon` 优先返回其主办方已保存的 `marker_icon`：

```json
{
  "id": "venue-9",
  "source": "venue",
  "source_id": 9,
  "icon": "https://cdn.hypercn.cn/marker-icons/qiche.png"
}
```

固定场地的 `source_id` 永远是 `organizer_id`，不再依赖或返回虚构的
`activity_id`。前端继续直接使用 `item.icon` 渲染地图，无需额外按 key 查找。
