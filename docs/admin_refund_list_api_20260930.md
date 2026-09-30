# 售后订单列表接口（2026-09-30）

```http
GET /api/v1/admin/refunds?page=1&pageSize=20&refund_status=pending_review&keyword=xxx
```

管理端「售后订单」页数据源，返回全部退款单（含订单/活动上下文）。

## Query 参数

| 参数 | 说明 |
|---|---|
| `page` / `pageSize` | 分页，pageSize 上限 100 |
| `refund_status` | 可选：`pending_review`（待审核/审核中）、`refunding`（退款中）、`refunded`（已退款）、`rejected`（已驳回）、`cancelled`（已取消），缺省全部；也支持直接传状态数字 |
| `keyword` | 可选：退款单号/订单号/购票人/用户昵称/手机号/活动名/票种名模糊搜索 |

## 响应

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "refund_no": "R202609291716490daca51b",
        "order_no": "T2026092812161856b82df5",
        "status": 0,
        "reason": "行程冲突",
        "refund_amount": 7900,
        "activity_id": 86,
        "activity_name": "OTALAB二周年特别活动Hypereal",
        "ticket_spec_name": "预售票",
        "buyer_name": "张三",
        "user_name": "昵称",
        "user_mobile": "138****0000",
        "created_at": "2026-09-29 17:16:00",
        "updated_at": "2026-09-29 17:16:00"
      }
    ],
    "total": 138,
    "page": 1,
    "pageSize": 20
  }
}
```

- `status`：0 待审核 1 退款中 2 已退款 3 已驳回 4 已取消
- `refund_amount` 单位分
- 详情/审核操作沿用已有接口：`GET /v1/admin/refunds/:refund_no`、`POST /v1/admin/orders/:order_no/refund/approve|reject`
