# 活动退票开关（2026-10-12）

## 背景

活动详情页「支持退票」标记此前由小程序前端写死，后端无字段。现改为活动级可配置。

## 改动

- `activities` 新增 `support_refund` 列：`0 不支持 / 1 支持`，**默认 1**（与历史行为一致，老数据自动支持退票）
- 发布/编辑活动 `POST /api/v1/activity/create`：请求体新增 `support_refund`（int，可选；nil 不修改）
- 活动详情 `GET /api/v1/activity/:id`：响应自动透出 `support_refund`（嵌在 activity 模型中）
- 退款强校验：`POST /api/v1/refund/apply` 对 `support_refund=0` 的活动拒绝申请，返回「该活动不支持退票」

## 前端配合

- 详情页标记：`support_refund == 1` 显示「支持退票」，否则不显示（或显示「不支持退票」）
- 发布/编辑表单：加「支持退票」开关，默认开

## 注意

`support_refund=0` 只拦截**新的退款申请**；已在退款流程中的订单不受影响。
