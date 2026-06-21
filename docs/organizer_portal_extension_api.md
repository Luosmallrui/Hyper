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

### 合集详情

```http
GET /api/v1/organizer/collections/:id
```

响应：

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
  "activity_ids": [8, 9, 10],
  "activities": [
    {
      "id": 8,
      "name": "活动名",
      "poster_list": "https://cdn.xxx/poster.png",
      "start_time": "2026-06-21T12:00:00+08:00",
      "end_time": "2026-06-21T14:00:00+08:00",
      "status": 3
    }
  ],
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

### 一键已读

```http
POST /api/v1/organizer/messages/read-all
```

响应：

```json
{
  "updated_count": 12
}
```

## 3. 商家资料与订阅统计

### 商家资料详情

```http
GET /api/v1/organizer/profile
```

响应：

```json
{
  "id": 3,
  "name": "测试主办方",
  "logo": "https://cdn.xxx/logo.png",
  "cover_image": "https://cdn.xxx/cover.png",
  "gallery": ["https://cdn.xxx/1.png"],
  "description": "商家简介",
  "business_hours": "18:00-02:00",
  "contact_name": "张三",
  "service_phone": "13800138000",
  "province": "四川省",
  "city": "成都市",
  "district": "郫都区",
  "address": "电子科技大学清水河校区",
  "latitude": 30.753,
  "longitude": 103.936,
  "average_spend": 8800
}
```

### 更新商家资料

```http
PUT /api/v1/organizer/profile
```

请求字段同详情响应。说明：

- `name/logo/province/city/district` 同步更新 `organizers`。
- `cover_image/gallery/description/business_hours/contact_name/service_phone/address/latitude/longitude/average_spend` 写入 `organizer_profiles`。
- 经纬度非 0 时后端会校验在中国大陆常用范围内。

### 订阅统计

```http
GET /api/v1/organizer/subscription/summary
```

响应：

```json
{
  "total_subscriptions": 128,
  "today_subscriptions": 6
}
```

### 手机号查询用户

用于商家端添加子账号前，把手机号解析为 `user_id`。

```http
GET /api/v1/organizer/users/lookup?phone=13800138000
```

响应：

```json
{
  "user_id": 52,
  "phone": "13800138000",
  "nickname": "张三",
  "avatar": "https://cdn.xxx/avatar.png",
  "status": 1
}
```

## 4. 商家财务

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

## 5. 商家等级规则

### 等级规则列表

```http
GET /api/v1/organizer/level-rules
```

首次调用会自动补默认规则：

- LV1：手续费 5%，0 场
- LV2：手续费 3%，5 场
- LV3：手续费 0%，10 场

商家端只读展示等级规则，不能新增、编辑、删除。

### 管理端等级规则配置

```http
GET    /api/v1/admin/organizer-level-rules
POST   /api/v1/admin/organizer-level-rules
PUT    /api/v1/admin/organizer-level-rules/:id
DELETE /api/v1/admin/organizer-level-rules/:id
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

## 6. 角色与子账号

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

## 7. 核销员状态

### 启用/停用核销员

```http
PATCH /api/v1/organizer/verifier/:id/status
PATCH /api/v1/admin/verifiers/:id/status
```

请求：

```json
{
  "status": 1
}
```

状态：

- `0`：停用
- `1`：启用

## 8. 商家动态

说明：商家动态使用现有 `notes` 数据表，但通过 organizer 专用接口管理。创建/编辑时，`activity_id` 必须属于当前商家；`store_id` 必须属于当前商家的门店。

状态约定：

- `1`：公开
- `2`：隐藏/私密
- `3`：违规/不可展示
- `-1`：已删除

### 动态列表

```http
GET /api/v1/organizer/posts?page=1&size=10&keyword=&status=&activity_id=&store_id=
```

响应 `data.list[]`：

```json
{
  "id": "2069000000000000000",
  "user_id": 52,
  "title": "今晚 9 点预热视频",
  "content": "动态正文",
  "media_data": [
    {
      "url": "https://cdn.xxx/post.png",
      "thumbnail_url": "",
      "width": 0,
      "height": 0,
      "duration": 0
    }
  ],
  "location": {
    "lat": 30.753,
    "lng": 103.936,
    "name": "电子科技大学清水河校区",
    "address": "成都市郫都区"
  },
  "type": 1,
  "status": 1,
  "visible_conf": 1,
  "activity_id": 10,
  "activity_name": "测试活动",
  "store_id": 3,
  "store_name": "测试门店",
  "like_count": 0,
  "coll_count": 0,
  "share_count": 0,
  "comment_count": 0,
  "created_at": "2026-06-21T12:00:00+08:00",
  "updated_at": "2026-06-21T12:00:00+08:00"
}
```

### 新增/编辑动态

```http
POST /api/v1/organizer/posts
PUT /api/v1/organizer/posts/:id
```

请求：

```json
{
  "title": "今晚 9 点预热视频",
  "content": "动态正文",
  "images": ["https://cdn.xxx/post.png"],
  "media_data": [],
  "location": {
    "lat": 30.753,
    "lng": 103.936,
    "name": "电子科技大学清水河校区",
    "address": "成都市郫都区"
  },
  "type": 1,
  "status": 1,
  "visible_conf": 1,
  "activity_id": 10,
  "store_id": 3
}
```

说明：

- `images` 是前端简化字段；如果 `media_data` 为空，后端会把 `images[]` 转成 `media_data[]`。
- `type` 默认 `1`，支持 `1` 图文、`2` 视频。
- `status` 默认 `1` 公开。

### 显示/隐藏动态

```http
PATCH /api/v1/organizer/posts/:id/visibility
```

请求二选一：

```json
{
  "visible": false
}
```

或：

```json
{
  "status": 2
}
```

### 删除动态

```http
DELETE /api/v1/organizer/posts/:id
```

说明：删除为软删除，后端将 `notes.status` 更新为 `-1`。

## 9. 商家操作日志

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

当前已记录合集、商家资料、核销员状态、商家动态、角色、子账号的新增/编辑/删除操作。等级规则写操作由管理端完成。

## 10. 数据库变更

本次新增：

```sql
CREATE TABLE IF NOT EXISTS organizer_profiles (...);
```

完整字段见 [config/table.sql](/Users/luosmallrui/Hyper/config/table.sql)。
