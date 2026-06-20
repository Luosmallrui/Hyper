# 场地相关动态接口说明

## 1. 创建动态时关联场地

```http
POST /api/v1/note/create
Authorization: Bearer <token>
```

请求体新增字段：

```json
{
  "title": "场地实拍",
  "content": "这里氛围不错",
  "type": 1,
  "visible_conf": 1,
  "store_id": 123,
  "activity_id": 0,
  "media_data": []
}
```

说明：

- `store_id`：关联门店/场地 ID。
- `activity_id`：可选，关联活动 ID。
- `visible_conf=1` 且动态审核通过 `status=1` 后，才会出现在相关动态列表。

## 2. 修改已有动态的场地/活动关联

```http
PATCH /api/v1/note/{note_id}/relation
Authorization: Bearer <token>
```

或：

```http
PUT /api/v1/note/{note_id}/relation
Authorization: Bearer <token>
```

请求体：

```json
{
  "store_id": 123,
  "activity_id": 7
}
```

清空关联时传 `0`：

```json
{
  "store_id": 0,
  "activity_id": 0
}
```

响应：

```json
{
  "code": 200,
  "data": {
    "success": true
  }
}
```

权限：

- 只能修改自己的动态。
- 已删除动态不可修改。

## 3. 获取场地相关动态

```http
GET /api/v1/note/related?store_id=123&pageSize=20
Authorization: Bearer <token>
```

也支持按活动查询：

```http
GET /api/v1/note/related?activity_id=7&pageSize=20
Authorization: Bearer <token>
```

响应沿用动态列表结构：

```json
{
  "code": 200,
  "data": {
    "notes": [],
    "next_cursor": 0, 
    "has_more": false
  }
}
```