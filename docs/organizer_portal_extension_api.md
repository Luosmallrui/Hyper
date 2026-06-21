# 商家端补齐接口文档

本文档记录本次新增的商家端接口，统一前缀：

```http
/api/v1/organizer
Authorization: Bearer <access_token>
```

## 1. 活动合集

### 合集列表

```http
GET /api/v1/organizer/collections?page=1&size=10&keyword=周末
```

响应 `data.list[]`：

```json
{
  "id": 1,
  "organizer_id": 3,
  "organizer_name": "测试主办方",
  "title": "周末活动合集",
  "share_title": "这个周末去哪里",
  "description": "合集简介",
  "share_image": "https://cdn.xxx/share.png",
  "status": 1,
  "activity_count": 3,
  "created_at": "2026-06-21T12:00:00+08:00",
  "updated_at": "2026-06-21T12:00:00+08:00"
}
```

### 新增/编辑合集

```http
POST /api/v1/organizer/collections
PUT /api/v1/organizer/collections/:id
```

请求：

```json
{
  "title": "周末活动合集",
  "share_title": "这个周末去哪里",
  "description": "合集简介",
  "share_image": "https://cdn.xxx/share.png",
  "status": 1,
  "activity_ids": [8, 9, 10]
}
```

说明：`activity_ids` 必须属于当前登录商家。

### 删除合集

```http
DELETE /api/v1/organizer/collections/:id
```

## 2. 商家消息

### 消息列表

```http
GET /api/v1/organizer/messages?page=1&size=10&unread_only=0
```

说明：返回平台发布给商家/全体对象的消息，`unread_only=1` 只看未读。

响应 `data.list[]`：

```json
{
  "id": 1,
  "title": "入驻成功",
  "content": "您的入驻申请已通过",
  "type": "notice",
  "target": "organizer",
  "is_read": false,
  "read_at": null,
  "created_at": "2026-06-21T12:00:00+08:00"
}
```

### 标记已读

```http
POST /api/v1/organizer/messages/:id/read
```

## 3. 商家财务

### 财务汇总

```http
GET /api/v1/organizer/finance/summary
```

响应：

```json
{
  "gross_amount": 10000,
  "refund_amount": 1000,
  "settle_amount": 8000,
  "withdraw_amount": 1000,
  "order_count": 12
}
```

字段单位均为分。

### 财务流水

```http
GET /api/v1/organizer/finance/flows?page=1&size=10&type=order
```

`type` 可选：

- `order`：订单收入
- `refund`：退款支出
- `withdraw`：提现支出
- 不传：全部

响应 `data.list[]`：

```json
{
  "id": "order-T202606210001",
  "type": "order",
  "amount": 1000,
  "order_no": "T202606210001",
  "activity_id": 10,
  "description": "票务订单收入",
  "created_at": "2026-06-21T12:00:00+08:00"
}
```

## 4. 商家等级规则

### 等级规则列表

```http
GET /api/v1/organizer/level-rules
```

首次调用会自动补默认规则：

- LV1：手续费 5%，0 场
- LV2：手续费 3%，5 场
- LV3：手续费 0%，10 场

### 新增/编辑等级规则

```http
POST /api/v1/organizer/level-rules
PUT /api/v1/organizer/level-rules/:id
```

请求：

```json
{
  "level": 2,
  "name": "LV2",
  "fee_rate": 0.03,
  "required_activity_count": 5,
  "description": "办理5场活动升级",
  "benefits": "手续费降至3%",
  "status": 1
}
```

### 删除等级规则

```http
DELETE /api/v1/organizer/level-rules/:id
```

## 5. 角色与子账号

### 角色列表

```http
GET /api/v1/organizer/roles?page=1&size=10
```

### 新增/编辑角色

```http
POST /api/v1/organizer/roles
PUT /api/v1/organizer/roles/:id
```

请求：

```json
{
  "name": "运营",
  "description": "可管理活动和订单",
  "permissions": ["activity.read", "activity.write", "order.read"],
  "status": 1
}
```

### 删除角色

```http
DELETE /api/v1/organizer/roles/:id
```

删除角色后，已绑定该角色的子账号 `role_id` 会置为 `0`。

### 子账号列表

```http
GET /api/v1/organizer/staff?page=1&size=10
```

### 新增/编辑子账号

```http
POST /api/v1/organizer/staff
PUT /api/v1/organizer/staff/:id
```

请求：

```json
{
  "user_id": 52,
  "role_id": 1,
  "name": "张三",
  "phone": "13800138000",
  "status": 1
}
```

### 删除子账号

```http
DELETE /api/v1/organizer/staff/:id
```

## 6. 商家操作日志

### 日志列表

```http
GET /api/v1/organizer/operation-logs?page=1&size=10&keyword=role
```

响应 `data.list[]`：

```json
{
  "id": 1,
  "organizer_id": 3,
  "operator_id": 52,
  "operator": "",
  "action": "save_role",
  "resource": "role",
  "method": "",
  "path": "",
  "ip": "",
  "remark": "role_id=1",
  "created_at": "2026-06-21T12:00:00+08:00"
}
```

当前已记录合集、等级规则、角色、子账号的新增/编辑/删除操作。
