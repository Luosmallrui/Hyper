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
