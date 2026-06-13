# 派对票务系统 API 对接文档

Base URL:

```text
/api/v1
```

除特别说明外，所有接口都需要：

```http
Authorization: Bearer <access_token>
Content-Type: application/json
```

统一响应结构沿用后端现有 `response.Success`：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {}
}
```

错误响应：

```json
{
  "code": 500,
  "msg": "错误信息"
}
```

金额字段统一使用 **分**，例如 `8800` 表示 `88.00` 元。

---

## 1. 状态枚举

### 主办方状态 `organizer.status`

| 值 | 说明 |
|---|---|
| 0 | 待审核 |
| 1 | 审核中 |
| 2 | 已认证 |
| 3 | 审核未通过 |

### 入驻类型 `organizer.type`

| 值 | 说明 |
|---|---|
| venue | 场地 |
| merchant | 商家 |

### 活动状态 `activity.status`

| 值 | 说明 |
|---|---|
| 0 | 草稿 |
| 1 | 待审核 |
| 2 | 审核中 |
| 3 | 已上架 |
| 4 | 审核未通过 |

### 票务订单状态 `order.status`

| 值 | 说明 |
|---|---|
| 0 | 待支付 |
| 1 | 待使用 |
| 2 | 已使用 |
| 3 | 已取消 |
| 4 | 退款中 |
| 5 | 退款成功 |
| 6 | 退款拒绝 |

### 退款状态 `refund.status`

| 值 | 说明 |
|---|---|
| 0 | 审核中 |
| 1 | 退款中 |
| 2 | 退款成功 |
| 3 | 退款拒绝 |

### 核销员状态 `verifier.status`

| 值 | 说明 |
|---|---|
| 0 | 未激活 |
| 1 | 已激活 |

---

## 2. 文件上传

### 通用文件上传

```http
POST /api/v1/upload
Content-Type: multipart/form-data
Authorization: Bearer <access_token>
```

FormData:

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| file | File | 是 | 文件 |
| type | string | 否 | `poster_detail` / `poster_long` / `poster_list` / `poster_wechat` / `qualification_doc` / `avatar` / `organizer_logo` |

响应：

```json
{
  "code": 200,
  "data": {
    "url": "https://cdn.hypercn.cn/ticketing/poster_list/2026/05/31/123.jpg"
  }
}
```

---

## 2.1 首页地图点位

### 获取首页地图点位

该接口用于小程序首页地图，统一返回旧派对/场地点位和新票务活动点位。

```http
GET /api/v1/map/markers?source=all&limit=200
```

Query:

| 字段 | 必填 | 说明 |
|---|---|---|
| source | 否 | `all` / `party` / `activity`，默认 `all` |
| limit | 否 | 每类最多返回数量，默认 `200`，最大 `500` |

数据来源：

| source | 来源表 | 条件 | 说明 |
|---|---|---|---|
| party | `parties` | `status=active` | 旧派对/场地 |
| activity | `activities` | `status=3` | 新票务活动，管理员审核通过后可见 |

响应：

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": "party-1",
        "source": "party",
        "source_id": 1,
        "user_id": 1,
        "user": "派对主理人",
        "username": "派对主理人",
        "user_avatar": "https://cdn.xxx/avatar.jpg",
        "userAvatar": "https://cdn.xxx/avatar.jpg",
        "title": "Hyper Club",
        "type": "场地",
        "location": "Hyper Club",
        "address": "朝阳区 xx 路",
        "lat": 39.9333,
        "lng": 116.4533,
        "cover_image": "https://cdn.xxx/cover.jpg",
        "created_at": "2026-06-03 12:00:00",
        "avg_price": 7600,
        "current_count": 9932,
        "post_count": 372,
        "icon": "https://cdn.hypercn.cn/icon/jiuba.png",
        "is_subscribe": false,
        "is_follow": false,
        "status": "active"
      },
      {
        "id": "activity-1",
        "source": "activity",
        "source_id": 1,
        "user_id": 1,
        "user": "Hyper Club",
        "username": "Hyper Club",
        "user_avatar": "https://cdn.xxx/logo.jpg",
        "userAvatar": "https://cdn.xxx/logo.jpg",
        "title": "周末电音派对",
        "type": "activity",
        "location": "朝阳区 xx 路",
        "address": "朝阳区 xx 路",
        "lat": 39.9333,
        "lng": 116.4533,
        "cover_image": "https://cdn.xxx/list.jpg",
        "created_at": "2026-06-03 12:00:00",
        "avg_price": 8800,
        "current_count": 0,
        "post_count": 0,
        "icon": "https://cdn.hypercn.cn/icon/party.png",
        "is_subscribe": false,
        "is_follow": false,
        "start_time": "2026-06-12 20:00:00",
        "end_time": "2026-06-13 02:00:00",
        "status": 3
      }
    ],
    "total": 2
  }
}
```

前端跳转建议：

| source | 点击后跳转 |
|---|---|
| party | 旧派对/场地详情，使用 `source_id` |
| activity | 新票务活动详情，使用 `source_id` 调 `/api/v1/activity/:id` |

说明：为兼容旧首页地图卡片，`activity` 类型也会返回 `user`、`username`、`user_avatar`、`userAvatar`、`avg_price`、`current_count`、`post_count`、`is_subscribe`、`is_follow` 等字段。`avg_price` 使用该活动票种最低价，单位为分。

## 3. 主办方模块

### 获取主办方信息

```http
GET /api/v1/organizer/info
```

响应：

```json
{
  "code": 200,
  "data": {
    "id": 1,
    "type": "venue",
    "name": "Hyper Club",
    "logo": "https://cdn.xxx/logo.png",
    "status": 2,
    "level": "LV1",
    "service_fee_rate": 0,
    "join_days": 12,
    "basic_info": {
      "province": "北京市",
      "city": "北京市",
      "district": "朝阳区"
    },
    "account_info": {
      "bank_account_name": "张三",
      "bank_account_no": "6222...",
      "bank_name": "招商银行"
    }
  }
}
```

### 入驻申请

```http
POST /api/v1/organizer/apply
```

请求：

```json
{
  "type": "venue",
  "name": "Hyper Club",
  "logo": "https://cdn.xxx/logo.png",
  "province": "北京市",
  "city": "北京市",
  "district": "朝阳区"
}
```

字段说明：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| type | string | 是 | `venue` 场地 / `merchant` 商家，二选一 |
| name | string | 是 | 入驻名称 |
| logo | string | 否 | Logo URL，可先用 `/api/v1/upload` 上传 |
| province | string | 否 | 省份 |
| city | string | 否 | 城市 |
| district | string | 否 | 区县 |

响应：

```json
{ "code": 200, "data": { "success": true } }
```

### 入驻审核状态

```http
GET /api/v1/organizer/audit-status
```

响应：

```json
{
  "code": 200,
  "data": {
    "type": "venue",
    "status": 1,
    "reject_reason": ""
  }
}
```

### 更新主办方基本信息

```http
PUT /api/v1/organizer/basic
```

请求：

```json
{
  "name": "新主办方名称",
  "logo": "https://cdn.xxx/logo.png",
  "province": "上海市",
  "city": "上海市",
  "district": "黄浦区"
}
```

### 获取提现信息

```http
GET /api/v1/organizer/withdraw-info
```

响应：

```json
{
  "code": 200,
  "data": {
    "bank_account_name": "张三",
    "bank_account_no": "6222...",
    "bank_name": "招商银行"
  }
}
```

### 修改提现信息

```http
PUT /api/v1/organizer/withdraw-info
```

请求：

```json
{
  "bank_account_name": "张三",
  "bank_account_no": "6222000000000000000",
  "bank_name": "招商银行"
}
```

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

## 4. 管理后台入驻审核

以下接口需要管理员 Token：

```http
Authorization: Bearer <admin_access_token>
```

### 入驻申请列表

```http
GET /api/v1/admin/organizers?page=1&pageSize=20&status=1&type=venue
```

Query:

| 字段 | 必填 | 说明 |
|---|---|---|
| page | 否 | 页码，默认 `1` |
| pageSize | 否 | 每页数量，默认 `20` |
| status | 否 | 主办方状态，不传返回全部 |
| type | 否 | `venue` / `merchant`，不传返回全部 |

### 入驻申请详情

```http
GET /api/v1/admin/organizers/:id
```

### 审核入驻申请

```http
PUT /api/v1/admin/organizers/:id/audit
```

通过：

```json
{
  "status": 2
}
```

拒绝：

```json
{
  "status": 3,
  "reject_reason": "资料不完整"
}
```

说明：

- `status=2` 表示审核通过，用户入驻成功。
- `status=3` 表示审核拒绝，必须填写 `reject_reason`。
- 普通用户申请 `venue` 或 `merchant` 任意一个通过后，即可视为入驻成功。

### 活动审核列表

```http
GET /api/v1/admin/activities?page=1&pageSize=20&status=1&keyword=周末&organizer_id=1
```

Query:

| 字段 | 必填 | 说明 |
|---|---|---|
| page | 否 | 页码，默认 `1` |
| pageSize | 否 | 每页数量，默认 `20` |
| status | 否 | 活动状态；不传返回已提交过审核的活动，不包含草稿；`1` 待审核，`2` 审核中，`3` 已上架，`4` 审核未通过 |
| keyword | 否 | 按活动名称或地址模糊搜索 |
| organizer_id | 否 | 按主办方 ID 筛选 |

响应：

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": 1,
        "organizer_id": 1,
        "organizer_name": "Hyper Club",
        "organizer_type": "venue",
        "name": "周末电音派对",
        "share_title": "周末见",
        "start_time": "2026-06-12 20:00:00",
        "end_time": "2026-06-13 02:00:00",
        "real_name_mode": 1,
        "minor_check": 1,
        "description": "活动概要",
        "province": "北京市",
        "city": "北京市",
        "district": "朝阳区",
        "address": "朝阳区 xx 路",
        "poster_list": "https://cdn.xxx/list.jpg",
        "qualification_doc": "https://cdn.xxx/doc.jpg",
        "status": 1,
        "reject_reason": "",
        "ticket_spec_count": 2,
        "created_at": "2026-06-03 12:00:00",
        "updated_at": "2026-06-03 12:00:00"
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 20
  }
}
```

### 活动审核详情

```http
GET /api/v1/admin/activities/:id
```

响应包含活动详情、主办方信息和票种列表：

```json
{
  "code": 200,
  "data": {
    "id": 1,
    "organizer_id": 1,
    "name": "周末电音派对",
    "status": 1,
    "reject_reason": "",
    "organizer": {
      "id": 1,
      "type": "venue",
      "name": "Hyper Club"
    },
    "ticket_specs": [
      {
        "id": 1,
        "activity_id": 1,
        "name": "早鸟票",
        "price": 8800,
        "stock": 100
      }
    ]
  }
}
```

### 审核活动

```http
PUT /api/v1/admin/activities/:id/audit
```

通过并上架：

```json
{
  "status": 3
}
```

拒绝：

```json
{
  "status": 4,
  "reject_reason": "活动资质不完整"
}
```

说明：

- `status=3` 表示审核通过并上架。
- `status=4` 表示审核拒绝，必须填写 `reject_reason`。

### 后台票务订单列表

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

### 后台票务订单详情

```http
GET /api/v1/admin/orders/:order_no
```

响应包含订单、退款单和退款进度：

```json
{
  "code": 200,
  "data": {
    "order": {
      "order_no": "T2026053114300012ab34cd",
      "status": 1,
      "actual_price": 8800,
      "activity_name": "周末电音派对",
      "ticket_spec_name": "早鸟票",
      "buyer_name": "罗小瑞"
    },
    "refunds": [],
    "refund_logs": []
  }
}
```

### 后台退款审批

```http
POST /api/v1/admin/orders/:order_no/refund/approve
POST /api/v1/admin/orders/:order_no/refund/reject
```

拒绝请求：

```json
{
  "reject_reason": "不符合退款规则"
}
```

说明：后台订单退款审批会流转退款单状态；实际微信退款仍建议使用 `/api/v1/pay/refund/:refund_no/approve` 发起。

### 后台财务结算

```http
GET /api/v1/admin/finance/summary
GET /api/v1/admin/finance/settlements?page=1&pageSize=20
GET /api/v1/admin/finance/settlements/:organizer_id
GET /api/v1/admin/finance/settlements/:organizer_id/export
```

### 后台用户管理

```http
GET /api/v1/admin/users?page=1&pageSize=20&keyword=13800138000
PUT /api/v1/admin/users/:id/status
```

状态请求：

```json
{
  "status": 0
}
```

说明：`status=1` 正常，`status=0` 封禁。

### 后台 Banner 管理

```http
GET /api/v1/admin/banners
POST /api/v1/admin/banners
PUT /api/v1/admin/banners/:id
DELETE /api/v1/admin/banners/:id
PUT /api/v1/admin/banners/sort
```

创建/更新请求：

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

排序请求：

```json
{
  "list": [
    { "id": 1, "sort": 1 },
    { "id": 2, "sort": 2 }
  ]
}
```

### 后台平台设置

```http
GET /api/v1/admin/settings
PUT /api/v1/admin/settings
```

更新请求：

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

### 绑定管理员微信通知

管理员需要先在小程序端授权订阅消息，再把 `wx.login` 得到的 `code` 传给后端绑定 openid。

前端流程：

```js
await wx.requestSubscribeMessage({
  tmplIds: ['6qWdBuY6p3EJ-rCRq7OWonTVAVlj-b6io1cDtNDlYTw']
})
const login = await wx.login()
await request({
  url: '/api/v1/admin/wechat-subscribe',
  method: 'POST',
  data: { code: login.code }
})
```

后端接口：

```http
POST /api/v1/admin/wechat-subscribe
```

请求：

```json
{
  "code": "wx.login 返回的 code"
}
```

说明：

- 绑定成功后，普通用户提交 `/api/v1/organizer/apply` 会自动给已绑定管理员发送订阅消息。
- 模板 ID：`6qWdBuY6p3EJ-rCRq7OWonTVAVlj-b6io1cDtNDlYTw`
- 模板字段：`thing1` 商家名称、`thing2` 申请人、`time3` 申请时间、`phrase4` 状态、`thing5` 备注。

---

## 5. 活动模块

### 创建/分步保存活动

```http
POST /api/v1/activity/create
```

说明：

- 首次创建不传 `activity_id`
- 后续步骤传上一次返回的 `activity_id`
- 时间支持：`2026-05-31T20:00`、`2026-05-31 20:00:00`、RFC3339
- `price` 使用分

Step 1 请求：

```json
{
  "step": 1,
  "name": "周末电音派对",
  "share_title": "一起蹦到凌晨",
  "start_time": "2026-06-12T20:00",
  "end_time": "2026-06-13T02:00",
  "real_name_mode": 1,
  "minor_check": 1,
  "description": "<p>活动介绍</p>"
}
```

Step 2 请求：

```json
{
  "activity_id": 1,
  "step": 2,
  "province": "北京市",
  "city": "北京市",
  "district": "朝阳区",
  "address": "工体西路",
  "latitude": 39.9333,
  "longitude": 116.4533
}
```

Step 3 请求：

```json
{
  "activity_id": 1,
  "step": 3,
  "poster_detail": "https://cdn.xxx/detail.jpg",
  "poster_long": "https://cdn.xxx/long.jpg",
  "poster_list": "https://cdn.xxx/list.jpg",
  "poster_wechat": "https://cdn.xxx/wechat.jpg"
}
```

Step 4 请求：

```json
{
  "activity_id": 1,
  "step": 4,
  "ticket_specs": [
    {
      "name": "早鸟票",
      "is_enabled": 1,
      "sale_start": "2026-06-01T10:00",
      "sale_end": "2026-06-12T18:00",
      "price": 8800,
      "stock": 100,
      "purchase_limit": 2,
      "max_attendees": 1
    }
  ]
}
```

Step 5 请求：

```json
{
  "activity_id": 1,
  "step": 5,
  "qualification_doc": "https://cdn.xxx/doc.pdf"
}
```

响应：

```json
{
  "code": 200,
  "data": {
    "activity_id": 1
  }
}
```

### 获取活动详情

```http
GET /api/v1/activity/:id
```

响应核心结构：

```json
{
  "code": 200,
  "data": {
    "id": 1,
    "organizer_id": 1,
    "name": "周末电音派对",
    "status": 0,
    "ticket_specs": [
      {
        "id": 1,
        "activity_id": 1,
        "name": "早鸟票",
        "price": 8800,
        "stock": 100,
        "sold_count": 0
      }
    ],
    "organizer": {
      "id": 1,
      "name": "Hyper Club"
    }
  }
}
```

### 我的活动列表

```http
GET /api/v1/activity/my-list?page=1&size=10
```

响应：

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": 1,
        "name": "周末电音派对",
        "poster_list": "https://cdn.xxx/list.jpg",
        "start_time": "2026-06-12T20:00:00+08:00",
        "end_time": "2026-06-13T02:00:00+08:00",
        "status": 0
      }
    ],
    "total": 1
  }
}
```

### 活动搜索

```http
GET /api/v1/activity/search?keyword=电音
```

只返回已上架活动。

### 删除活动

```http
DELETE /api/v1/activity/:id
```

### 提交活动审核

```http
POST /api/v1/activity/:id/submit-audit
```

### 活动统计总览

```http
GET /api/v1/activity/:id/statistics
```

响应：

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

响应：

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

## 6. 票券配置模块

### 获取活动票券列表

```http
GET /api/v1/activity/:id/ticket-specs
```

响应：

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": 1,
        "name": "早鸟票",
        "is_enabled": 1,
        "sale_start": "2026-06-01T10:00:00+08:00",
        "sale_end": "2026-06-12T18:00:00+08:00",
        "price": 8800,
        "stock": 100,
        "sold_count": 0,
        "purchase_limit": 2,
        "max_attendees": 1
      }
    ]
  }
}
```

### 保存票券配置

```http
POST /api/v1/activity/:id/ticket-specs
```

请求：

```json
{
  "specs": [
    {
      "id": 1,
      "name": "早鸟票",
      "is_enabled": 1,
      "sale_start": "2026-06-01T10:00",
      "sale_end": "2026-06-12T18:00",
      "price": 8800,
      "stock": 100,
      "purchase_limit": 2,
      "max_attendees": 1
    }
  ]
}
```

说明：

- 有 `id` 表示更新
- 无 `id` 表示新增

### 删除票券规格

```http
DELETE /api/v1/ticket-spec/:id
```

---

## 7. 订单模块

### 观演人列表

```http
GET /api/v1/viewers
```

响应：

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": 1,
        "real_name": "罗小瑞",
        "id_card": "5001**********0817",
        "phone": "138****8000",
        "type": 2,
        "created_at": "2026-06-02T12:00:00+08:00",
        "updated_at": "2026-06-02T12:00:00+08:00"
      }
    ],
    "total": 1
  }
}
```

### 创建观演人

```http
POST /api/v1/viewers
```

请求：

```json
{
  "real_name": "罗小瑞",
  "id_card": "500101199811040817",
  "phone": "13800138000"
}
```

响应：

```json
{
  "code": 200,
  "data": {
    "success": true,
    "id": 1
  }
}
```

说明：

- 单个用户最多 5 个常用观演人。
- 身份证号必须通过校验码校验。
- 返回列表时身份证号、手机号会脱敏。

### 更新观演人

```http
PUT /api/v1/viewers/:id
```

请求：

```json
{
  "real_name": "罗小瑞",
  "phone": "13800138000"
}
```

### 删除观演人

```http
DELETE /api/v1/viewers/:id
```

说明：如果观演人已关联未完成票务订单，暂不可删除。

### 创建票务订单

```http
POST /api/v1/order/create
```

请求：

```json
{
  "activity_id": 1,
  "ticket_spec_id": 1,
  "quantity": 1,
  "buyer_name": "罗小瑞",
  "buyer_id_card": "500101199811040817"
}
```

响应：

```json
{
  "code": 200,
  "data": {
    "success": true,
    "order_no": "T2026053114300012ab34cd"
  }
}
```

说明：

- 活动必须为 `status=3` 已上架
- 票券必须启用并且在销售时间内
- 开启实名模式时，必须传 `buyer_name` 和 `buyer_id_card`
- 开启未成年人校验时，18 岁以下不可购票

### 发起微信支付

票务订单创建成功后，直接使用现有支付接口：

```http
POST /api/v1/pay/prepay
```

请求：

```json
{
  "order_no": "T2026053114300012ab34cd"
}
```

响应：

```json
{
  "code": 200,
  "data": {
    "prepay_id": "wx201410272009395522657a690389285100",
    "appId": "wx8888888888888888",
    "timeStamp": "1414561699",
    "nonceStr": "5K8264ILTKCH16CQ2502SI8ZNMTM67VS",
    "package": "prepay_id=wx201410272009395522657a690389285100",
    "signType": "RSA",
    "paySign": "xxx",
    "out_trade_no": "T2026053114300012ab34cd"
  }
}
```

说明：

- 票务支付只需要传 `order_no`
- 后端会从 `ticket_orders` 读取金额，前端不需要传 `amount`
- 微信支付成功回调后，后端会把 `ticket_orders.status` 从 `0` 更新为 `1`
- `pay_method` 会写入微信回调里的交易类型，`pay_time` 写入回调处理时间

### 获取订单详情

```http
GET /api/v1/order/:order_no
```

响应：

```json
{
  "code": 200,
  "data": {
    "order_no": "T2026053114300012ab34cd",
    "status": 0,
    "total_price": 8800,
    "actual_price": 8800,
    "quantity": 1,
    "activity": {
      "id": 1,
      "name": "周末电音派对",
      "start_time": "2026-06-12T20:00:00+08:00",
      "end_time": "2026-06-13T02:00:00+08:00",
      "poster_list": "https://cdn.xxx/list.jpg"
    },
    "ticket_spec": {
      "name": "早鸟票"
    },
    "buyer_name": "罗小瑞",
    "buyer_id_card": "500101199811040817",
    "pay_method": "",
    "pay_time": null,
    "created_at": "2026-05-31T14:30:00+08:00",
    "qr_code": "TICKET:T2026053114300012ab34cd:xxxx",
    "qr_code_url": "https://cdn.hypercn.cn/ticket/qrcode/2026/06/12/T2026053114300012ab34cd.png",
    "expire_time": "2026-05-31T14:45:00+08:00"
  }
}
```

说明：`qr_code_url` 是用户购票后的线下取票/核销二维码图片；图片内容为 `qr_code` 字符串，核销端扫码后将扫码内容传给 `/api/v1/verifier/scan`。

### 我的票务订单列表

```http
GET /api/v1/order/list?page=1&size=10&status=1
```

查询参数：

| 参数 | 必填 | 说明 |
|---|---|---|
| page | 否 | 页码，默认 1 |
| size | 否 | 每页数量，默认 10 |
| status | 否 | 订单状态；不传返回全部 |
| legacy | 否 | 传 `1` 时返回旧商品订单列表；不传时返回票务订单列表 |

响应：

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "order_no": "T2026053114300012ab34cd",
        "status": 1,
        "total_price": 8800,
        "actual_price": 8800,
        "quantity": 1,
        "activity": {
          "id": 1,
          "name": "周末电音派对",
          "start_time": "2026-06-12T20:00:00+08:00",
          "end_time": "2026-06-13T02:00:00+08:00",
          "poster_list": "https://cdn.xxx/list.jpg"
        },
        "ticket_spec": {
          "id": 1,
          "name": "早鸟票"
        },
        "buyer_name": "罗小瑞",
        "buyer_id_card": "5001**********0817",
        "created_at": "2026-05-31T14:30:00+08:00",
        "expire_time": "2026-05-31T14:45:00+08:00",
        "pay_time": "2026-05-31T14:32:00+08:00"
      }
    ],
    "total": 1
  }
}
```

### 获取取消原因列表

```http
GET /api/v1/order/cancel-reasons
```

### 取消订单

```http
POST /api/v1/order/:order_no/cancel
```

请求：

```json
{
  "reason_id": 1
}
```

---

## 8. 退款模块

### 获取退款原因列表

```http
GET /api/v1/refund/reasons
```

### 申请退款

```http
POST /api/v1/refund/apply
```

请求：

```json
{
  "order_no": "T2026053114300012ab34cd",
  "reason_id": 1
}
```

响应：

```json
{
  "code": 200,
  "data": {
    "success": true,
    "refund_no": "R20260531150000abcd"
  }
}
```

### 主办方查看退款审核列表

```http
GET /api/v1/organizer/refunds?page=1&size=10&status=0
Authorization: Bearer <access_token>
```

参数：

| 参数 | 必填 | 说明 |
|---|---|---|
| page | 否 | 页码，默认 1 |
| size | 否 | 每页数量，默认 10 |
| status | 否 | 退款状态，不传返回全部；`0` 表示待审核 |

响应：

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "refund_no": "R20260531150000abcd",
        "status": 0,
        "refund_amount": 8800,
        "deduct_amount": 0,
        "reason": "行程冲突",
        "reject_reason": "",
        "expect_arrive_date": "2026-06-03",
        "wechat_refund_id": "",
        "wechat_status": "",
        "order_no": "T2026053114300012ab34cd",
        "user_id": 1,
        "buyer_name": "罗小瑞",
        "buyer_id_card": "500101199811040817",
        "activity_id": 1,
        "activity_name": "周末电音派对",
        "ticket_spec_id": 1,
        "ticket_spec_name": "早鸟票",
        "quantity": 1,
        "created_at": "2026-05-31T15:00:00+08:00"
      }
    ],
    "total": 1
  }
}
```

### 审核通过并发起微信退款

后台审核退款通过后调用：

```http
POST /api/v1/pay/refund/:refund_no/approve
Authorization: Bearer <access_token>
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

后端处理逻辑：

- 校验退款单必须是 `status=0` 审核中
- 校验对应票务订单必须是 `status=4` 退款中
- 校验支付流水 `pay_records.pay_status=2`
- 调用微信支付退款 API：`/v3/refund/domestic/refunds`
- 微信受理后把 `refunds.status` 更新为 `1` 退款中
- 微信退款成功回调后把 `refunds.status` 更新为 `2`，订单状态更新为 `5`

### 审核拒绝退款

```http
POST /api/v1/refund/:refund_no/reject
Authorization: Bearer <access_token>
```

请求：

```json
{
  "reject_reason": "活动开始前24小时内不可退款"
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

后端处理逻辑：

- 只允许主办方拒绝自己活动的退款
- 退款单必须是 `status=0` 审核中
- 拒绝后 `refunds.status=3`
- 拒绝后 `ticket_orders.status=6`
- 写入一条 `refund_logs`

### 微信退款回调

微信商户平台退款通知地址配置为：

```http
POST /api/v1/pay/refund-notify
```

服务端会验签、解密并处理退款结果。前端不需要调用。

### 获取退款详情

```http
GET /api/v1/refund/:refund_no
```

### 取消退款

```http
POST /api/v1/refund/:refund_no/cancel
```

---

## 9. 核销模块

### 核销员列表

```http
GET /api/v1/organizer/verifiers?page=1&size=10
```

### 添加核销员

```http
POST /api/v1/organizer/verifier
```

请求：

```json
{
  "name": "核销员A",
  "phone": "13800138000"
}
```

### 删除核销员

```http
DELETE /api/v1/organizer/verifier/:id
```

### 获取核销员激活码

```http
GET /api/v1/organizer/verifier/:id/activation-qr
```

响应：

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

### 获取核销员绑定确认信息

```http
GET /api/v1/verifier/activation-info?v=1
```

说明：微信小程序码只携带短参数 `scene=v%3D1`，不会直接带 `organizerName` 这种中文 query。小程序绑定页解析 `v` 后调用本接口获取主办方名称。

响应：

```json
{
  "code": 200,
  "data": {
    "verifier_id": 1,
    "name": "核销员A",
    "phone": "13800138000",
    "status": 0,
    "is_bound": false,
    "organizer_id": 10,
    "organizer_name": "测试主办方"
  }
}
```

### 核销员激活

```http
POST /api/v1/verifier/activate
Authorization: Bearer <access_token>
```

说明：核销员激活本质是“绑定邀请记录到小程序用户”。小程序扫码进入后，应先完成登录/注册；未注册用户先走现有手机号登录/注册流程，拿到 `access_token` 后再调用本接口。后端会校验登录用户手机号和后台添加的核销员手机号一致。

请求：

```json
{
  "phone": "13800138000"
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

### 扫码识别订单

```http
POST /api/v1/verifier/scan
```

请求：

```json
{
  "qr_code": "TICKET:T2026053114300012ab34cd:xxxx",
  "activity_id": 1
}
```

响应成功：

```json
{
  "code": 200,
  "data": {
    "success": true,
    "order": {
      "activity_name": "周末电音派对",
      "ticket_spec_name": "早鸟票",
      "quantity": 1,
      "buyer_name_masked": "罗**",
      "buyer_id_card_masked": "5001**********0817"
    }
  }
}
```

响应失败：

```json
{
  "code": 200,
  "data": {
    "success": false,
    "error_code": "ALREADY_VERIFIED"
  }
}
```

错误码：

| error_code | 说明 |
|---|---|
| INVALID_QR | 无效订单码 |
| ALREADY_VERIFIED | 已核销 |
| NOT_VERIFIABLE_TIME | 非可核销时间 |
| ORDER_NOT_FOUND | 订单不存在 |
| ORDER_CANCELLED | 订单已取消 |
| WRONG_ACTIVITY | 订单不属于当前活动 |

### 确认核销

```http
POST /api/v1/verifier/confirm
X-Verifier-Id: <verifier_id>
```

请求：

```json
{
  "order_no": "T2026053114300012ab34cd"
}
```

说明：当前版本先用 `X-Verifier-Id` 传核销员 ID，后续可以替换为独立核销员 token。

### 已核销列表

```http
GET /api/v1/verifier/verified-list?page=1&size=10
X-Verifier-Id: <verifier_id>
```

---

## 10. 前端推荐流程

### 活动发布

1. `POST /api/v1/organizer/apply`
2. 上传海报：`POST /api/v1/upload`
3. 分步保存：`POST /api/v1/activity/create`
4. 保存票券：`POST /api/v1/activity/:id/ticket-specs`
5. 提交审核：`POST /api/v1/activity/:id/submit-audit`

### 用户购票

1. `GET /api/v1/activity/:id`
2. `POST /api/v1/order/create`
3. `POST /api/v1/pay/prepay`，只传 `order_no` 发起微信支付
4. `GET /api/v1/order/:order_no` 查询订单和二维码

### 核销

1. 主办方添加核销员：`POST /api/v1/organizer/verifier`
2. 主办方获取激活码：`GET /api/v1/organizer/verifier/:id/activation-qr`
3. 核销员扫码进入小程序，未注册则先登录/注册
4. 核销员绑定邀请：`POST /api/v1/verifier/activate`
5. 扫码识别：`POST /api/v1/verifier/scan`
6. 确认核销：`POST /api/v1/verifier/confirm`

---

## 11. 当前实现边界

- 新票务订单使用 `ticket_orders`，旧商品订单仍使用 `orders/products/pay`。
- 微信支付已支持通过 `/api/v1/pay/prepay` + `order_no` 支付 `ticket_orders`。
- 微信退款已支持通过 `/api/v1/pay/refund/:refund_no/approve` 发起，回调地址为 `/api/v1/pay/refund-notify`。
- 核销员当前临时使用 `X-Verifier-Id`，后续可升级为独立核销员 token。
