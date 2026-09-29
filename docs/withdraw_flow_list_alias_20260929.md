# 提现流水接口字段别名（2026-09-29）

接口：`GET /api/v1/organizer/bank/withdraw/flow/list`（与 `/v1/organizer/withdraws` 同一 handler）

为兼容新旧前端，响应**双写**两组字段，老字段不删不改：

## 顶层

| 字段 | 说明 |
|---|---|
| `list` / `total` | 旧版形状（`models.OrganizerWithdraw` 平铺字段，保留） |
| `flow_list` / `total_items` | 新版形状 |

## `flow_list[]` 字段映射

| 新字段 | 来源 | 说明 |
|---|---|---|
| `id` | `id` | |
| `flow_no` | 生成：`WD + yyyyMMdd + id(6位)` | |
| `status` | `status` | 0/1 处理中，2 已打款，其余为驳回等 |
| `total_amount` | `amount` | 金额，单位分 |
| `reason` | `remark` | 驳回原因等 |
| `bank_account.account_holder` | `bank_account_name` | 收款人 |
| `bank_account.bank_name` | `bank_name` | 银行 |
| `create_time` | `created_at` | 申请时间 `yyyy-MM-dd HH:mm:ss` |
| `arrival_time` | `updated_at`（仅 status=2） | 未打款为空串 |

实现：`handler/ticketing.go` `ListOrganizerWithdraws` + `buildWithdrawFlowItems`。
