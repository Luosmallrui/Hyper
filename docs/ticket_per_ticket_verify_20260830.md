# 门票按张核销（2026-08-30）

## 背景

购票人买 N 张票，朋友先到场：原来一个订单一个二维码，扫一次整单核销，N 张票无法分开进场。
现改为**按张核销**：同一二维码可分多次核销，每次默认核销 1 张，全部核销完成后订单才置为已使用。

## 契约变更（全部为增量字段，向后兼容）

### `POST /api/v1/verifier/scan` 响应

`data.order` 新增两个字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| `verified_count` | int | 累计已核销张数 |
| `remaining` | int | 剩余可核销张数（= quantity - verified_count） |

### `POST /api/v1/verifier/confirm` 请求

新增可选字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| `quantity` | int | 本次核销张数，缺省为 1；超过剩余张数时按剩余张数核销 |

### `POST /api/v1/verifier/confirm` 响应

原响应为 `{"success": true}`，现 `data` 扩展为：

```json
{
  "success": true,
  "order_no": "T...",
  "quantity": 2,
  "verified_count": 1,
  "remaining": 1
}
```

## 行为说明

- 一张票对应一条 `verification_records` 记录，已核销张数 = 该订单的核销记录数（无表结构变更）。
- 订单部分核销时 `status` 保持"可使用"，最后一张核销完成才置为"已使用"。
- 历史已整单核销的老订单（只有 1 条核销记录但 quantity>1）：scan 时 `verified_count` 按 quantity 展示、`remaining` 为 0，confirm 仍提示不可核销，行为与改造前一致。
- 重复核销（剩余 0 张）confirm 返回错误"订单已全部核销"。
- 并发安全：订单行 `FOR UPDATE` 锁 + 事务内计数，同一订单并发核销不会超核。

## 退款限制（产品决策 2026-08-24）

部分核销的订单不允许退款：

- `POST /api/v1/refund/apply`：订单存在任意核销记录（含部分核销）时拒绝申请，返回错误「已核销门票不支持退款」。
- 订单详情接口（用户侧 `GET /api/v1/order/:order_no`、主办方侧、核销员侧）新增字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| `verified_count` | int | 已核销张数 |
| `remaining` | int | 剩余可核销张数 |

前端据此在 `verified_count > 0` 时隐藏「申请退款」入口。
