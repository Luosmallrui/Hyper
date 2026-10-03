# 管理端订单可视化后端支持（2026-10-04）

对应前端需求 `frontend_api_requirements_20261003.md`（P0-1 / P0-2 / P1），金额单位均为**分**。

## P0-1 订单列表：商家字段 + 商家筛选（已上线）

`GET /api/v1/admin/orders` 增量：

- 新增查询参数 `organizer_id`（数字，可选）：按商家（活动所属主办方）过滤，与 `activity_id`/`status`/`refund_status`/`keyword`/`sales_channel` 组合生效
- 列表项新增字段：`organizer_id`、`organizer_name`（取订单所属活动的主办方）

## P0-2 订单聚合统计（已上线）

```http
GET /api/v1/admin/orders/stats?days=30
```

`days` 可选，默认 30，上限 90。响应：

```json
{
  "code": 200,
  "data": {
    "daily": [{"date": "2026-10-01", "order_count": 12, "gross_amount": 94800, "refund_amount": 7900}],
    "status_counts": {"0": 3, "1": 40, "2": 18, "3": 5, "4": 2, "5": 6, "6": 1},
    "channel_counts": {"wechat": 70, "other": 3}
  }
}
```

### 统计口径（重要）

| 字段 | 口径 |
|---|---|
| `daily[].order_count` | **按支付时间**（pay_time）统计，仅已支付订单（状态 1 待使用 / 2 已使用 / 4 退款中 / 5 已退款 / 6 退款拒绝——这些状态都付过钱）；待支付(0)、已取消(3)不计入 |
| `daily[].gross_amount` | 同上范围的 `actual_price` 实付合计（分），按支付时间 |
| `daily[].refund_amount` | 当日**申请**退款金额合计（refunds.created_at，含后续被驳回的） |
| `status_counts` | 窗口内全部订单按**创建时间**，key 为主状态码字符串 |
| `channel_counts` | 窗口内全部订单按**创建时间**，空渠道归入 `other` |

- `daily` 按日期升序，窗口内无数据的日期补零（`order_count: 0, gross_amount: 0, refund_amount: 0`）
- 窗口为「今天往前 days-1 天」至今天（本地时区），含首尾两天

## P1 支付流水字段统一（已上线）

订单详情、退款详情中的 `pay_records[]` 统一为：

```json
{
  "transaction_id": "4500000435202609221946220053",
  "amount": 7900,
  "pay_method": "JSAPI",
  "status": 2,
  "pay_time": "2026-09-22T19:46:22+08:00"
}
```

- `amount` 固定 number 类型、单位分（来自 `amount_total`）
- `status` 为支付流水状态：0 待支付、2 已支付、4 已关闭
- `pay_time` 为 RFC3339（来自 `finished_at`），未支付为空串
- 0 元单/积分单无微信流水时返回空数组
- ⚠️ 原模型的 `amount_total`/`pay_status`/`finished_at` 等字段名不再随 `pay_records` 返回，前端旧字段引用需切换

实现位置：`service/admin.go` `GetOrderStats` / `GetTicketOrderList` / `normalizeAdminPayRecords`；路由 `handler/admin.go`。
