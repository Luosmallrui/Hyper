# 票务接口缺口补齐建议

本文档用于记录 `docs/ticketing_api.md` 尚未覆盖、但前端票务页面已经需要替换的接口契约。

Base URL:

```text
/api/v1
```

除特别说明外，接口均需要：

```http
Authorization: Bearer <access_token>
Content-Type: application/json
```

金额字段统一使用分。

---

## 1. 缺口总览

| 前端位置 | 当前临时处理 | 缺口 | 建议优先级 |
|---|---|---|---|
| 活动购票页观演人选择 | 已新增 `/api/v1/viewers` | 正式观演人 CRUD 已补齐 | 已完成 |
| 我的票务订单列表页 | 已使用 `/api/v1/order/list` | 新票务订单列表已补齐；旧商品订单通过 `legacy=1` 兼容 | 已完成 |
| 主办方后台销售数据 | adapter 返回空数组/0 | 缺少销售概览、趋势、票种统计接口 | P1 |
| 主办方后台订单统计 | adapter 返回空数组/0 | 缺少后台订单列表和状态统计接口 | P1 |
| 入驻弹窗扩展字段 | 只提交 `ticketing_api.md` 已允许字段 | 联系人、电话、简介、资质暂无后端契约 | P1 |

---

## 2. 观演人 CRUD

### 现状

当前后端已有旧路由：

```http
POST /api/v1/order/create-viewer
POST /api/v1/order/delete-viewer
GET /api/v1/order/list-viewer
```

已新增标准 CRUD，并保留旧接口作为兼容层。

### 已新增：观演人列表

```http
GET /api/v1/viewers
```

响应：

```json
{
  "code": 200,
  "data": {
    "total": 2,
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
    ]
  }
}
```

字段说明：

| 字段 | 类型 | 说明 |
|---|---|---|
| id | int64 | 观演人 ID |
| real_name | string | 真实姓名 |
| id_card | string | 脱敏身份证号 |
| phone | string | 脱敏手机号 |
| type | int | 年龄类型：1 未成年，2 成年，3 老年 |

### 已新增：创建观演人

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

校验建议：

- 单个用户最多 5 个常用观演人。
- 身份证号必须通过校验码校验。
- 同一用户不可重复添加相同身份证。
- 身份证、手机号是否允许跨用户重复，需要后端明确；如继续沿用当前实现，则身份证和手机号均全局唯一。

### 已新增：更新观演人

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

说明：

- 建议不允许更新 `id_card`。如果业务允许修改身份证，应按“删除后重建”或增加专门的实名变更审核流程。

### 已新增：删除观演人

```http
DELETE /api/v1/viewers/:id
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

删除约束建议：

- 如果观演人已被未结束、未取消、未退款成功的票务订单使用，应禁止删除或仅做软删除。

### 前端替换点

| 当前接口 | 建议替换为 |
|---|---|
| `/api/v1/order/list-viewer` | `GET /api/v1/viewers` |
| `/api/v1/order/create-viewer` | `POST /api/v1/viewers` |
| `/api/v1/order/delete-viewer` | `DELETE /api/v1/viewers/:id` |

---

## 3. 我的票务订单列表

### 现状

`ticketing_api.md` 已有：

```http
POST /api/v1/order/create
GET /api/v1/order/:order_no
POST /api/v1/order/:order_no/cancel
```

已补齐“我的票务订单列表”。前端订单列表页可从 `/api/v1/order/liist` 替换为 `/api/v1/order/list`。注意 `liist` 疑似旧接口拼写，建议不要继续沿用。

### 已新增

```http
GET /api/v1/order/list?page=1&size=10&status=1
```

查询参数：

| 参数 | 必填 | 说明 |
|---|---|---|
| page | 否 | 页码，默认 1 |
| size | 否 | 每页数量，默认 10 |
| status | 否 | 订单状态；不传返回全部 |

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
        "viewers": [
          {
            "viewer_id": 12,
            "real_name": "罗小瑞",
            "id_card_masked": "5001**********0817",
            "phone_masked": "138****8000",
            "type": 1
          }
        ],
        "created_at": "2026-05-31T14:30:00+08:00",
        "expire_time": "2026-05-31T14:45:00+08:00",
        "pay_time": "2026-05-31T14:32:00+08:00"
      }
    ],
    "total": 1
  }
}
```

实现建议：

- 数据源使用 `ticket_orders`，不要混用旧商品订单表 `orders`。
- 仅返回当前登录用户自己的订单。
- 列表页身份证字段统一脱敏，详情页是否展示完整身份证由产品和合规要求决定。
- 待支付订单如果已超过 `expire_time`，后端可在查询时返回 `status=3` 或补充 `is_expired=true`，二者需要明确一种。

### 前端替换点

| 当前接口 | 建议替换为 |
|---|---|
| `/api/v1/order/liist` | `GET /api/v1/order/list` |

---

## 4. 主办方后台订单列表

### 现状

主办方后台需要按活动、订单状态、关键字查看票务订单，但 `ticketing_api.md` 暂未提供后台订单列表接口。

### 建议新增

```http
GET /api/v1/organizer/orders?page=1&size=10&activity_id=1&status=1&keyword=罗小瑞
```

查询参数：

| 参数 | 必填 | 说明 |
|---|---|---|
| page | 否 | 页码，默认 1 |
| size | 否 | 每页数量，默认 10 |
| activity_id | 否 | 活动 ID |
| status | 否 | 订单状态 |
| keyword | 否 | 订单号、购票人姓名、手机号或身份证后四位 |
| start_date | 否 | 下单开始日期，格式 `YYYY-MM-DD` |
| end_date | 否 | 下单结束日期，格式 `YYYY-MM-DD` |

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
        "activity_id": 1,
        "activity_name": "周末电音派对",
        "ticket_spec_id": 1,
        "ticket_spec_name": "早鸟票",
        "buyer_name": "罗小瑞",
        "buyer_id_card": "5001**********0817",
        "pay_method": "JSAPI",
        "pay_time": "2026-05-31T14:32:00+08:00",
        "created_at": "2026-05-31T14:30:00+08:00"
      }
    ],
    "total": 1
  }
}
```

权限要求：

- 只返回当前登录主办方名下活动的订单。
- 若当前用户不是已认证主办方，应返回权限错误。

---

## 5. 主办方销售数据与订单统计

### 现状

后台销售数据、订单统计暂无接口，前端 adapter 暂时保持空数组和 0。

### 建议新增：销售概览

```http
GET /api/v1/organizer/stats/overview?activity_id=1&start_date=2026-06-01&end_date=2026-06-30
```

响应：

```json
{
  "code": 200,
  "data": {
    "gross_sales": 17600,
    "paid_sales": 17600,
    "refund_amount": 0,
    "net_sales": 17600,
    "order_count": 2,
    "paid_order_count": 2,
    "refund_order_count": 0,
    "ticket_count": 2,
    "verified_count": 1,
    "pending_verify_count": 1
  }
}
```

字段口径建议：

| 字段 | 口径 |
|---|---|
| gross_sales | 已支付、退款中、退款成功、已使用等支付成功后产生的原始销售额 |
| paid_sales | 支付成功且未退款成功的订单实付金额 |
| refund_amount | 退款成功金额 |
| net_sales | `paid_sales - refund_amount`，如 `paid_sales` 已排除退款成功，则需后端明确 |
| ticket_count | 支付成功订单的票数合计 |
| verified_count | 已核销票数 |

### 建议新增：销售趋势

```http
GET /api/v1/organizer/stats/trend?activity_id=1&start_date=2026-06-01&end_date=2026-06-30&granularity=day
```

响应：

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "date": "2026-06-01",
        "sales": 8800,
        "order_count": 1,
        "ticket_count": 1
      }
    ]
  }
}
```

### 建议新增：票种统计

```http
GET /api/v1/organizer/stats/ticket-specs?activity_id=1
```

响应：

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "ticket_spec_id": 1,
        "ticket_spec_name": "早鸟票",
        "price": 8800,
        "stock": 100,
        "sold_count": 20,
        "paid_count": 18,
        "refunded_count": 2,
        "verified_count": 10,
        "sales": 158400
      }
    ]
  }
}
```

---

## 6. 入驻申请扩展字段

### 现状

入驻弹窗 UI 目前包含：

- 联系人
- 电话
- 简介
- 资质

但 `POST /api/v1/organizer/apply` 当前只允许提交：

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

因此前端真实提交时只能提交文档允许字段，其余字段暂不提交。

### 建议扩展

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
  "district": "朝阳区",
  "contact_name": "罗小瑞",
  "contact_phone": "13800138000",
  "intro": "北京市朝阳区 Livehouse 与派对活动场地",
  "qualification_docs": [
    "https://cdn.xxx/qualification/001.jpg"
  ]
}
```

字段说明：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| contact_name | string | 建议必填 | 联系人 |
| contact_phone | string | 建议必填 | 联系电话 |
| intro | string | 否 | 主办方/场地/商家简介 |
| qualification_docs | string[] | 建议必填 | 资质文件 URL 列表，通过 `/api/v1/upload?type=qualification_doc` 或现有上传接口上传 |

### 响应建议

```json
{
  "code": 200,
  "data": {
    "success": true,
    "organizer_id": 1,
    "status": 0
  }
}
```

### 数据模型建议

可直接扩展 `organizers` 表：

| 字段 | 类型 | 说明 |
|---|---|---|
| contact_name | varchar(50) | 联系人 |
| contact_phone | varchar(20) | 联系电话 |
| intro | varchar(1000) | 简介 |
| qualification_docs | json | 资质文件 URL 列表 |

也可以新建 `organizer_profiles` 表承载扩展信息。若后续入驻审核需要版本化记录申请内容，建议使用独立申请表，例如 `organizer_applications`。

---

## 7. 建议落地顺序

1. P1：补齐 `GET /api/v1/organizer/orders`，后台订单列表先可用。
2. P1：补齐 `GET /api/v1/organizer/stats/overview`，让后台核心统计不再返回 0。
3. P1：补齐销售趋势和票种统计，替换 adapter 空数组。
4. P1：扩展 `/api/v1/organizer/apply` 入驻字段，或明确这些 UI 字段由哪个接口承接。

---

## 8. 前端临时兼容策略

在后端契约补齐前，前端可继续采用以下策略：

| 模块 | 临时策略 |
|---|---|
| 观演人 | 改用 `/api/v1/viewers`；旧 `/api/v1/order/list-viewer`、`create-viewer`、`delete-viewer` 暂时兼容 |
| 我的票务订单列表 | 改用 `/api/v1/order/list`；旧商品订单列表如仍需调用，传 `legacy=1` |
| 后台销售数据 | 返回空数组和 0，并避免展示误导性的真实统计 |
| 入驻申请 | 只提交 `type/name/logo/province/city/district`，联系人、电话、简介、资质暂存 UI，不进入真实请求 |
