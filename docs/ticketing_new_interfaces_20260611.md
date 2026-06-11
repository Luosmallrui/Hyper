# 票务新增接口说明

更新时间：2026-06-11

本文档只记录本轮新增的后端接口，方便前端按差距分析逐项联调。

Base URL:

```text
/api/v1
```

除特别说明外，均需要登录态 Token。

---

## 1. 活动数据统计

### 活动统计总览

```http
GET /api/v1/activity/:id/statistics
```

返回：

```json
{
  "code": 200,
  "data": {
    "verify_rate": 0.5,
    "ticket_count": 20,
    "buyer_count": 18,
    "gross_amount": 176000,
    "refund_amount": 8800,
    "net_amount": 167200,
    "average_ticket_price": 8800,
    "verified_count": 10
  }
}
```

### 活动每日统计

```http
GET /api/v1/activity/:id/statistics/daily?days=7
```

说明：`days` 默认 `7`，最大 `90`。

返回：

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "date": "2026-06-11",
        "amount": 8800,
        "ticket_count": 1,
        "order_count": 1
      }
    ],
    "total": 1
  }
}
```

---

## 2. 主办方门店管理

### 门店列表

```http
GET /api/v1/organizer/stores?page=1&size=10&keyword=三里屯
```

### 创建门店

```http
POST /api/v1/organizer/stores
```

请求：

```json
{
  "name": "Hyper 三里屯店",
  "logo": "https://cdn.xxx/store.jpg",
  "address": "北京市朝阳区三里屯",
  "latitude": 39.9333,
  "longitude": 116.4533,
  "phone": "13800138000"
}
```

### 编辑门店

```http
PUT /api/v1/organizer/stores/:id
```

请求字段同创建门店。

### 删除门店

```http
DELETE /api/v1/organizer/stores/:id
```

---

## 3. 后台票务订单管理

以下接口需要管理员 Token。

### 订单列表

```http
GET /api/v1/admin/orders?page=1&pageSize=20&activity_id=1&status=1&keyword=罗小瑞
```

Query:

| 字段 | 必填 | 说明 |
|---|---|---|
| page | 否 | 页码，默认 `1` |
| pageSize | 否 | 每页数量，默认 `20` |
| activity_id | 否 | 按活动筛选 |
| status | 否 | 按票务订单状态筛选 |
| keyword | 否 | 订单号、购票人姓名、身份证号模糊搜索 |

### 订单详情

```http
GET /api/v1/admin/orders/:order_no
```

返回包含：

- `order`
- `refunds`
- `refund_logs`

### 后台退款审批通过

```http
POST /api/v1/admin/orders/:order_no/refund/approve
```

说明：该接口只做后台审核流转；真正发起微信退款仍建议走：

```http
POST /api/v1/pay/refund/:refund_no/approve
```

### 后台退款审批拒绝

```http
POST /api/v1/admin/orders/:order_no/refund/reject
```

请求：

```json
{
  "reject_reason": "不符合退款规则"
}
```

---

## 4. 后台财务结算

以下接口需要管理员 Token。

### 财务汇总

```http
GET /api/v1/admin/finance/summary
```

返回：

```json
{
  "code": 200,
  "data": {
    "total_revenue": 176000,
    "pending_settle": 167200,
    "settled_amount": 0,
    "refund_amount": 8800,
    "order_count": 20
  }
}
```

### 结算列表

```http
GET /api/v1/admin/finance/settlements?page=1&pageSize=20
```

### 单个主办方结算详情

```http
GET /api/v1/admin/finance/settlements/:organizer_id
```

### 单个主办方结算导出

```http
GET /api/v1/admin/finance/settlements/:organizer_id/export
```

说明：当前返回结算数据结构，未生成真实文件。

---

## 5. 后台用户管理

以下接口需要管理员 Token。

### 用户列表

```http
GET /api/v1/admin/users?page=1&pageSize=20&keyword=13800138000
```

返回字段包含：

- `id`
- `nickname`
- `avatar`
- `mobile`
- `status`
- `total_amount`
- `order_count`
- `created_at`
- `updated_at`

### 修改用户状态

```http
PUT /api/v1/admin/users/:id/status
```

请求：

```json
{
  "status": 0
}
```

说明：

- `status=1` 正常
- `status=0` 封禁

数据库新增字段：

```sql
ALTER TABLE `users` ADD COLUMN `status` tinyint NOT NULL DEFAULT 1 COMMENT '用户状态: 1正常 0封禁' AFTER `gender`;
```

---

## 6. 后台 Banner 管理

以下接口需要管理员 Token。

### Banner 列表

```http
GET /api/v1/admin/banners
```

### 创建 Banner

```http
POST /api/v1/admin/banners
```

请求：

```json
{
  "title": "首页活动",
  "image": "https://cdn.xxx/banner.jpg",
  "link": "/pages/activity/detail?id=1",
  "position": "home",
  "sort": 1,
  "status": 1
}
```

### 编辑 Banner

```http
PUT /api/v1/admin/banners/:id
```

请求字段同创建。

### 删除 Banner

```http
DELETE /api/v1/admin/banners/:id
```

### Banner 排序

```http
PUT /api/v1/admin/banners/sort
```

请求：

```json
{
  "list": [
    { "id": 1, "sort": 1 },
    { "id": 2, "sort": 2 }
  ]
}
```

---

## 7. 后台平台设置

以下接口需要管理员 Token。

### 获取平台设置

```http
GET /api/v1/admin/settings
```

### 更新平台设置

```http
PUT /api/v1/admin/settings
```

请求：

```json
{
  "settings": [
    {
      "key": "service_fee_rate",
      "value": "0.05",
      "remark": "平台服务费率"
    }
  ]
}
```

---

## 8. 核销员扫码绑定

本轮调整了核销员激活语义：后台添加核销员时只录入姓名和手机号，生成激活码；核销员扫码进入小程序后，先完成登录/注册，再把这条核销员邀请绑定到自己的小程序用户。

### 获取核销员激活码

```http
GET /api/v1/organizer/verifier/:id/activation-qr
```

返回微信小程序码图片 URL 和小程序码参数：

```json
{
  "code": 200,
  "data": {
    "wechat_mini_program_code_url": "https://cdn.hypercn.cn/verifier/qrcode/2026/06/11/1.png",
    "wechat_qr_url": "https://cdn.hypercn.cn/verifier/qrcode/2026/06/11/1.png",
    "wechat_qr_page": "pages/user-sub/verifier-bind/index",
    "wechat_scene": "v=1",
    "wechat_qr": "https://cdn.hypercn.cn/verifier/qrcode/2026/06/11/1.png",
    "douyin_qr": "hyper://verifier/activate?verifier_id=1&channel=douyin"
  }
}
```

### 激活并绑定核销员

```http
POST /api/v1/verifier/activate
Authorization: Bearer <access_token>
```

小程序侧流程：

1. 扫码解析 `verifier_id` 和 `channel`。
2. 如果当前用户未登录/未注册，先走现有手机号登录/注册流程。
3. 带登录态调用本接口完成绑定。

请求：

```json
{
  "verifier_id": 1,
  "phone": "13800138000",
  "channel": "wechat"
}
```

响应：

```json
{
  "code": 200,
  "data": {
    "success": true,
    "verifier_id": 1,
    "user_id": 10001,
    "status": 1
  }
}
```

校验规则：

- 登录用户 `users.mobile` 必须等于后台添加的核销员手机号。
- 同一核销员邀请已绑定其他用户时，返回错误：`该核销员已绑定其他账号`。
- 同一用户重复扫码绑定同一邀请时幂等成功。

上线 DDL：

```sql
ALTER TABLE `verifiers` ADD COLUMN `user_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '绑定的小程序用户ID' AFTER `organizer_id`;
ALTER TABLE `verifiers` ADD COLUMN `bound_at` datetime NULL COMMENT '绑定时间' AFTER `channel`;
ALTER TABLE `verifiers` ADD KEY `idx_verifier_user` (`user_id`) USING BTREE;
```
