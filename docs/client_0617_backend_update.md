# 客户端 0617 后端更新与前端对接说明

本文对应 `docs/Hyper小程序开发进度表_0617.xlsx` 中需要后端配合的部分，聚焦本次已补齐的接口能力和前端接入口径。

默认请求头：

```http
Authorization: Bearer <access_token>
```

## 1. 首页搜索、区域和筛选

### 全局搜索

```http
GET /api/v1/search/?keyword=电音&type=0&limit=10&distance=5&district=锦江区&area=春熙路&business_area=太古里&tags=1,4&lat=30.657&lng=104.066
```

已支持参数：

| 参数 | 说明 |
|---|---|
| keyword | 关键词；筛选场景下可为空 |
| type | `0` 综合；`1` 用户；`2` 动态；`3` 商家/场地；`4` 票务活动 |
| limit | 每类返回数量 |
| distance | 距离，单位 km |
| district | 行政区。票务活动按 `activities.district` 精确匹配；旧商家按 `district_id` 兼容数字 ID |
| area | 商圈/区域。旧商家按 `area_id` 兼容数字 ID；票务活动按地址模糊匹配 |
| business_area | 商圈名称，按地址/位置模糊匹配 |
| tags | 标签位，例如 `1,4` |
| lat/lng | 用户当前位置，配合距离筛选 |

响应新增 `activities`：

```json
{
  "code": 200,
  "data": {
    "parties": [],
    "activities": [
      {
        "id": 1,
        "name": "电音派对",
        "poster_list": "https://cdn.xxx/poster.png",
        "start_time": "2026-06-20T20:00:00+08:00",
        "end_time": "2026-06-20T23:00:00+08:00",
        "province": "四川省",
        "city": "成都市",
        "district": "锦江区",
        "address": "太古里",
        "lat": 30.657,
        "lng": 104.066,
        "avg_price": 9900,
        "status": 3
      }
    ]
  }
}
```

### 地图点位

```http
GET /api/v1/map/markers?source=all&keyword=电音&distance=5&district=锦江区&area=春熙路&business_area=太古里&tags=1,4&lat=30.657&lng=104.066
```

已支持：

- `keyword`
- `distance`
- `lat/lng`
- `district`
- `area`
- `business_area`
- `tags` / `tag_ids`
- `category_id`
- `district_id`
- `area_id`

返回字段补充：

```json
{
  "distance": 1.23,
  "district": "锦江区",
  "area": "春熙路",
  "support_points": true,
  "discount_tags": []
}
```

说明：

- `distance` 单位 km。
- 未传 `lat/lng` 时不做距离过滤。
- 票务活动暂不做优惠标签硬过滤，返回 `discount_tags: []`，前端可先做选中态和空态。

### 商家/场地列表

```http
GET /api/v1/merchant/list?page=1&pageSize=20&keyword=酒吧&distance=5&district=1&area=2&business_area=太古里&tags=1,4&lat=30.657&lng=104.066
```

本次修复：

- 支持 `keyword`。
- 支持 `district` / `district_id`。
- 支持 `area` / `area_id`。
- 支持 `business_area`。
- 支持 `tags` / `tag_ids`。
- 支持 `distance + lat/lng`。
- 修复 `total` 写死问题，改为真实总数。

## 2. 积分体系与购票抵扣

### 积分余额

已有接口：

```http
GET /api/v1/points/balance
```

响应兼容字段：

```json
{
  "code": 200,
  "data": {
    "balance": 300,
    "points": 300,
    "available": 300,
    "available_points": 300
  }
}
```

### 积分流水

已有接口：

```http
GET /api/v1/points/records?action=income&cursor=0&limit=20
```

### 创建票务订单抵扣积分

```http
POST /api/v1/order/create
```

请求字段：

```json
{
  "activity_id": 1,
  "ticket_spec_id": 2,
  "quantity": 1,
  "use_points": true,
  "points_amount": 100
}
```

返回字段：

```json
{
  "order_no": "T202606170001",
  "total_price": 19900,
  "points_amount": 100,
  "points_discount": 1000,
  "actual_price": 18900
}
```

规则：

- 当前抵扣比例：`1 积分 = 10 分`。
- 后端校验积分余额。
- 抵扣金额不能超过订单金额。
- 积分抵扣会写入积分流水，`change_type=11`。

### 支付成功返积分

本次新增：

- 票务订单支付成功后自动返积分。
- 默认规则：`10 元消费 = 1 积分`，按实际支付金额四舍五入。
- 使用订单号做幂等，微信回调重复不会重复发放。
- 积分流水 `change_type=5`，备注为 `票务订单消费返积分`。

## 3. 订单与售后状态

### 订单列表

```http
GET /api/v1/order/list?status=0&page=1&size=10
```

按售后状态筛选：

```http
GET /api/v1/order/list?refund_status=pending_review&page=1&size=10
```

售后状态枚举：

| refund_status | 含义 |
|---|---|
| pending_review | 待审核 |
| refunding | 退款中 |
| refunded | 已退款 |
| rejected | 已驳回 |
| cancelled | 已取消 |

返回新增字段：

```json
{
  "order_no": "T202606170001",
  "status": 4,
  "refund_no": "R202606170001",
  "refund_status": "pending_review",
  "refund_status_text": "待审核"
}
```

说明：

- `status` 仍为订单主状态。
- `refund_status` 为最新一条退款记录状态。
- 已退款订单会自动同步订单主状态为 `5`。

### 订单详情

```http
GET /api/v1/order/:order_no
```

详情同样返回：

```json
{
  "refund_status": "refunded",
  "refund_status_text": "已退款",
  "refund_no": "R202606170001"
}
```

### 继续支付

已有接口：

```http
POST /api/v1/pay/prepay
```

请求：

```json
{
  "order_no": "T202606170001"
}
```

规则：

- 仅 `status=0` 的待支付票务订单可继续支付。
- 订单过期后返回错误。
- 后端会按订单实际支付金额重新拉起微信预支付。

## 4. 入驻审核与商家后台准入

### 提交入驻申请

```http
POST /api/v1/organizer/apply
```

### 查询审核状态

```http
GET /api/v1/organizer/audit-status
```

当前响应：

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

状态：

| status | 含义 |
|---|---|
| 0 | 待提交/待处理 |
| 1 | 审核中 |
| 2 | 已通过 |
| 3 | 已驳回 |

前端准入建议：

- `status=1`：展示审核中，不允许重复提交，不允许进入商家后台。
- `status=2`：允许进入商家后台。
- `status=3`：展示驳回原因，允许修改后重新提交。

## 5. 商家端核销

已有接口：

```http
POST /api/v1/verifier/scan
POST /api/v1/verifier/confirm
GET /api/v1/verifier/verified-list?page=1&size=10
```

说明：

- 核销成功后会写入 `verification_records`。
- `GET /api/v1/verifier/verified-list` 可立即查询到核销记录。
- 已核销订单再次核销会返回重复核销提示。

## 6. 商家等级体系

### 商家信息

```http
GET /api/v1/organizer/info
```

本次新增/补齐字段：

```json
{
  "level": "LV1",
  "service_fee_rate": 0.05,
  "level_value": 1,
  "fee_rate": 0.05,
  "completed_activity_count": 3,
  "next_level_required_count": 5
}
```

默认规则：

| 等级 | 条件 | 手续费 |
|---|---|---|
| 1 | 已举办活动数 `< 5` | 5% |
| 2 | 已举办活动数 `>= 5` 且 `< 10` | 3% |
| 3 | 已举办活动数 `>= 10` | 0% |

说明：

- 已举办活动数按当前时间已结束的活动统计。
- 暂未接后台等级规则配置，先使用默认规则。

## 7. 前端仍需处理

- 删除或隐藏旧商品购买、商品订单、商品管理入口。
- 删除或隐藏商家端活动抽奖及配置入口。
- 首页更多筛选中的标签 UI 选中态、清除态、默认距离项。
- 场地详情相关动态空态。
- 动态评论回复后的本地列表刷新和黑色蒙版问题。
- 订单页售后 tab UI，不新增独立售后模块。

## 8. 后续可增强

- 平台后台配置积分抵扣比例。
- 活动维度配置是否允许积分抵扣。
- 后台配置商家等级规则。
- 活动优惠标签字段落库后，票务活动可支持真实标签过滤。
