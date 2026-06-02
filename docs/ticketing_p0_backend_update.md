# 票务 P0 后端更新说明

更新时间：2026-06-02

本文档记录本次 P0 后端代码更新。前端此前已经做过对接，因此本文重点用于确认接口行为、兼容策略和回归检查点。

---

## 1. 更新范围

本次已落地两个 P0 缺口：

| 模块 | 状态 | 说明 |
|---|---|---|
| 观演人 CRUD | 已实现 | 新增正式 REST 接口 `/api/v1/viewers` |
| 我的票务订单列表 | 已实现 | `/api/v1/order/list` 默认返回票务订单列表 |

本次未处理的 P1 项仍保持原状：

| 模块 | 当前状态 |
|---|---|
| 主办方后台订单列表 | 待补接口 |
| 主办方销售数据/订单统计 | adapter 可继续返回空数组/0 |
| 入驻申请联系人、电话、简介、资质 | 暂不提交，继续只提交已约定字段 |

---

## 2. 观演人接口

### 2.1 列表

```http
GET /api/v1/viewers
```

响应结构：

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

注意事项：

- `id_card`、`phone` 返回时已脱敏。
- `type` 为年龄类型：`1` 未成年，`2` 成年，`3` 老年。
- 列表只返回当前登录用户自己的观演人。

### 2.2 创建

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

后端校验：

- 单个用户最多 5 个常用观演人。
- 身份证号必须满足格式和校验码。
- 同一身份证不可重复绑定。
- 手机号如果已被其他观演人绑定，会返回错误。

### 2.3 更新

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

- `id` 从路径读取，body 中不再要求传 `id`。
- 当前更新接口不支持修改身份证号。

### 2.4 删除

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

删除限制：

- 如果观演人身份证已关联当前用户未完成票务订单，后端会拒绝删除。
- 未完成状态包括：待支付、待使用、退款中、退款拒绝。

### 2.5 旧接口兼容

旧接口暂时保留：

```http
POST /api/v1/order/create-viewer
POST /api/v1/order/delete-viewer
GET /api/v1/order/list-viewer
```

前端已经接入新接口时，不需要再回退到旧接口。

---

## 3. 我的票务订单列表

### 3.1 新行为

```http
GET /api/v1/order/list?page=1&size=10&status=1
```

查询参数：

| 参数 | 必填 | 说明 |
|---|---|---|
| page | 否 | 页码，默认 1 |
| size | 否 | 每页数量，默认 10 |
| status | 否 | 票务订单状态，不传返回全部 |

响应结构：

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

注意事项：

- 数据源为 `ticket_orders`。
- 只返回当前登录用户自己的票务订单。
- `buyer_id_card` 在列表中已脱敏。
- `status` 使用 `ticketing_api.md` 中的票务订单状态枚举。

### 3.2 旧商品订单兼容

由于旧商品订单之前也使用 `/api/v1/order/list`，本次做了兼容分流：

```http
GET /api/v1/order/list?legacy=1
```

行为：

- 不传 `legacy=1`：返回新票务订单列表。
- 传 `legacy=1`：返回旧商品订单列表，保持原有 cursor 分页结构。

前端如果已经全部切到票务订单，不需要传 `legacy=1`。

---

## 4. 前端确认清单

前端已对接后，建议只做以下回归确认：

| 场景 | 预期 |
|---|---|
| 进入购票页选择观演人 | 调用 `GET /api/v1/viewers`，展示脱敏身份证和手机号 |
| 新增观演人 | 调用 `POST /api/v1/viewers`，成功后返回 `id` |
| 编辑观演人 | 调用 `PUT /api/v1/viewers/:id`，body 不需要传 `id` |
| 删除未关联订单的观演人 | 调用 `DELETE /api/v1/viewers/:id` 成功 |
| 删除已关联未完成订单的观演人 | 后端返回错误，前端展示失败提示 |
| 我的票务订单列表 | 调用 `/api/v1/order/list`，不传 `legacy=1` |
| 旧商品订单列表如仍保留入口 | 调用 `/api/v1/order/list?legacy=1` |

---

## 5. 代码改动位置

| 文件 | 改动 |
|---|---|
| `handler/ticketing.go` | 新增 `/api/v1/viewers` 路由和 handler |
| `service/ticketing.go` | 新增观演人 CRUD 和票务订单列表查询 |
| `types/ticketing.go` | 新增 `ViewerItem`、`TicketOrderListItem` |
| `types/order.go` | 调整 `UpdateViewerReq`，body 不再强制要求 `id` |
| `handler/order.go` | `/api/v1/order/list` 默认转到票务订单列表，`legacy=1` 走旧商品订单 |
| `cmd/api-server/wire_gen.go` | 将 `TicketingService` 注入旧订单 handler 用于兼容分流 |

---

## 6. 验证情况

已通过：

```bash
GOCACHE=/private/tmp/hyper-go-build go test ./handler ./service ./types ./cmd/api-server
```

全量测试未作为本次通过依据：

- `pkg/llm` 测试依赖外网 DNS。
- `pkg/socket/adapter` 现有测试存在空指针 panic。

以上失败与本次 P0 接口改动无关。

