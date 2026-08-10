# 订单销售渠道筛选接口

更新时间：2026-08-10

Base URL：`/api/v1`

本文定义票务订单的销售渠道归因与筛选。`sales_channel` 是订单创建时的来源快照，和 `pay_method`（例如 `JSAPI`、`POINTS`）是两个不同概念。

## 1. 发布前 SQL

生产库已有 `ticket_orders` 时，先执行：

```sql
ALTER TABLE `ticket_orders`
  ADD COLUMN `sales_channel` varchar(20) NOT NULL DEFAULT 'wechat'
  COMMENT '销售渠道：wechat/douyin/web/other'
  AFTER `pay_method`;

ALTER TABLE `ticket_orders`
  ADD KEY `idx_ticket_order_sales_channel` (`sales_channel`);

UPDATE `ticket_orders`
SET `sales_channel` = 'wechat'
WHERE `sales_channel` IS NULL OR `sales_channel` = '';
```

已有系统此前只承接微信小程序支付，因此历史订单统一归因到 `wechat`。若历史订单实际存在其他来源，应在执行更新前按真实数据分别回填。

## 2. 渠道枚举

| 值 | 展示名称 | 说明 |
|---|---|---|
| `wechat` | 微信 | 微信小程序、微信内 H5 等微信来源订单。 |
| `douyin` | 抖音 | 抖音小程序或抖音内来源订单。 |
| `web` | 网页 | PC/H5 独立站来源订单。 |
| `other` | 其他 | 暂无法归类的来源。 |

兼容入参 `wechat_mini_program` 和 `douyin_mini_program`，后端会分别规范化保存为 `wechat`、`douyin`。

## 3. 创建订单

```http
POST /api/v1/order/create
Authorization: Bearer <access_token>
Content-Type: application/json
```

请求示例：

```json
{
  "activity_id": 15,
  "ticket_spec_id": 30,
  "quantity": 2,
  "viewer_ids": [101, 102],
  "use_points": false,
  "sales_channel": "wechat"
}
```

`sales_channel` 可不传；为兼容现有小程序，下单接口默认保存为 `wechat`。订单创建成功后渠道不可修改。

成功响应新增：

```json
{
  "code": 200,
  "data": {
    "order_no": "T2026081015300012ab34cd",
    "status": 0,
    "actual_price": 12800,
    "sales_channel": "wechat"
  }
}
```

## 4. 商家端订单筛选

```http
GET /api/v1/organizer/orders?page=1&size=20&sales_channel=douyin
Authorization: Bearer <organizer_access_token>
```

可与现有 `activity_id`、`status`、`keyword`、`withdraw_status`、`start_date`、`end_date` 组合使用。

订单列表和订单详情都会返回：

```json
{
  "order_no": "T2026081015300012ab34cd",
  "pay_method": "JSAPI",
  "sales_channel": "wechat"
}
```

## 5. 管理端订单筛选

```http
GET /api/v1/admin/orders?page=1&pageSize=20&sales_channel=wechat
Authorization: Bearer <admin_access_token>
```

可与现有 `activity_id`、`status`、`refund_status`、`keyword` 组合使用。管理端订单列表与订单详情中的 `order` 对象均返回 `sales_channel`。

## 6. 前端注意事项

1. 微信小程序下单固定传 `wechat`；当前不传也会默认归为微信。
2. 抖音客户端接入后固定传 `douyin`，不要使用支付方式字段推断渠道。
3. 前端筛选值使用规范值 `wechat`、`douyin`、`web`、`other`。
4. 当前服务的支付回调仍是微信支付链路。`sales_channel=douyin` 只完成订单归因和数据筛选；要完成真实抖音收款，还需要单独接入抖音登录、预下单和支付回调，并按回调更新该订单。
