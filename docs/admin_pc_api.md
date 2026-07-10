# 管理端 PC 网页端接口说明

本文整理本次新增和已有的管理端接口。默认前缀为 `/api/v1/admin`，除登录外均需要：

```http
Authorization: Bearer <admin_access_token>
```

分页统一参数：

| 参数 | 必填 | 说明 |
|---|---|---|
| page | 否 | 页码，默认 1 |
| pageSize | 否 | 每页数量，默认 20，最大 100 |

分页统一响应：

```json
{
  "code": 200,
  "data": {
    "list": [],
    "total": 0,
    "page": 1,
    "pageSize": 20
  }
}
```

## 登录与个人资料

| 功能 | 方法 | 路径 |
|---|---|---|
| 管理员登录 | POST | `/login` |
| 当前管理员资料 | GET | `/profile` |
| 更新资料 | PUT | `/profile` |
| 修改密码 | PUT | `/profile/password` |

更新资料：

```json
{
  "avatar": "https://cdn.xxx/avatar.png",
  "mobile": "13800138000",
  "email": "admin@hyper.cn",
  "motto": "运营后台"
}
```

修改密码：

```json
{
  "old_password": "old",
  "new_password": "new"
}
```

## 控制台与系统配置

| 功能 | 方法 | 路径 |
|---|---|---|
| 基础统计 | GET | `/dashboard` |
| 系统配置列表 | GET | `/settings` |
| 更新系统配置 | PUT | `/settings` |
| Banner 列表 | GET | `/banners` |
| 新增 Banner | POST | `/banners` |
| 更新 Banner | PUT | `/banners/:id` |
| 删除 Banner | DELETE | `/banners/:id` |
| Banner 排序 | PUT | `/banners/sort` |

## 权限管理

| 功能 | 方法 | 路径 |
|---|---|---|
| 管理员列表 | GET | `/admins?keyword=` |
| 新增管理员 | POST | `/admins` |
| 更新管理员 | PUT | `/admins/:id` |
| 删除管理员 | DELETE | `/admins/:id` |
| 角色列表 | GET | `/roles?keyword=` |
| 新增角色 | POST | `/roles` |
| 更新角色 | PUT | `/roles/:id` |
| 删除角色 | DELETE | `/roles/:id` |
| 操作日志 | GET | `/logs?admin_id=&keyword=` |

管理员请求：

```json
{
  "username": "ops",
  "password": "123456",
  "avatar": "",
  "mobile": "13800138000",
  "email": "",
  "status": 1
}
```

角色请求：

```json
{
  "name": "运营",
  "description": "运营后台角色",
  "permissions": "[\"admin.users\",\"admin.orders\"]",
  "status": 1
}
```

说明：当前已落角色和权限字段，接口层暂未强制 RBAC 拦截；后续可在 `AdminAuth` 后增加权限中间件。

## 分类管理

统一使用分类配置接口：

| 功能 | 方法 | 路径 |
|---|---|---|
| 分类列表 | GET | `/categories?type=activity` |
| 新增分类 | POST | `/categories` |
| 更新分类 | PUT | `/categories/:id` |
| 删除分类 | DELETE | `/categories/:id` |

建议 `type`：

```text
activity          活动分类
note_channel      动态频道
distance          距离筛选
coupon_tag        优惠标签
other             其它分类
```

请求：

```json
{
  "type": "activity",
  "name": "演出",
  "image": "https://cdn.xxx/icon.png",
  "value": "show",
  "sort": 1,
  "status": 1
}
```

## 用户与观演用户

| 功能 | 方法 | 路径 |
|---|---|---|
| 用户列表 | GET | `/users?keyword=` |
| 用户启停 | PUT | `/users/:id/status` |
| 获赞记录 | GET | `/users/:id/records/likes` |
| 收藏记录 | GET | `/users/:id/records/collections` |
| 关注列表 | GET | `/users/:id/records/following` |
| 粉丝列表 | GET | `/users/:id/records/followers` |
| 参加活动记录 | GET | `/users/:id/records/attends` |
| 订阅活动记录 | GET | `/users/:id/records/subscribes` |
| 观演用户列表 | GET | `/viewers?keyword=` |

用户启停：

```json
{
  "status": 1
}
```

## 商家与活动

| 功能 | 方法 | 路径 |
|---|---|---|
| 商家列表 | GET | `/organizers?status=&type=` |
| 商家详情 | GET | `/organizers/:id` |
| 商家审核 | PUT | `/organizers/:id/audit` |
| 商家启停 | PATCH | `/organizers/:id/status` |
| 删除商家 | DELETE | `/organizers/:id` |
| 活动列表 | GET | `/activities?status=&keyword=&organizer_id=` |
| 活动详情 | GET | `/activities/:id` |
| 活动审核 | PUT | `/activities/:id/audit` |
| 票务列表 | GET | `/tickets?keyword=` |

删除商家说明：

- 删除后该用户不再具备主办方/商家身份，无法继续进入商家端。
- 关联核销员、门店、活动合集会删除。
- 关联活动会置为 `status=4` 拒绝，并写入 `reject_reason=主办方已被管理员删除`。
- 历史订单、退款、支付流水不删除，避免影响财务追溯。

## 活动合集

| 功能 | 方法 | 路径 |
|---|---|---|
| 合集列表 | GET | `/activity-collections?keyword=&organizer_id=` |
| 新增合集 | POST | `/activity-collections` |
| 更新合集 | PUT | `/activity-collections/:id` |
| 删除合集 | DELETE | `/activity-collections/:id` |

请求：

```json
{
  "organizer_id": 1,
  "title": "本周精选",
  "share_title": "一起出门玩",
  "description": "合集简介",
  "share_image": "https://cdn.xxx/share.png",
  "status": 1,
  "activity_ids": [1, 2, 3]
}
```

## 订单与售后

| 功能 | 方法 | 路径 |
|---|---|---|
| 平台订单列表 | GET | `/orders?activity_id=&status=&refund_status=&keyword=` |
| 订单详情 | GET | `/orders/:order_no` |
| 退款通过 | POST | `/orders/:order_no/refund/approve` |
| 退款拒绝 | POST | `/orders/:order_no/refund/reject` |

订单列表筛选：

| 参数 | 说明 |
|---|---|
| keyword | 支持订单号、买家姓名、身份证、用户昵称、手机号、活动名、票种名 |
| status | 订单主状态 |
| refund_status | 售后状态，支持数字或 `pending_review` / `refunding` / `refunded` / `rejected` / `cancelled` |
| activity_id | 活动 ID |

订单详情额外返回：

| 字段 | 说明 |
|---|---|
| viewers | 订单实名观演人 |
| verification_records | 核销记录 |
| pay_records | 支付流水 |
| refunds | 退款单 |
| refund_logs | 退款处理日志 |

拒绝退款：

```json
{
  "reject_reason": "不符合退款规则"
}
```

## 核销管理

| 功能 | 方法 | 路径 |
|---|---|---|
| 核销员列表 | GET | `/verifiers?keyword=&organizer_id=` |
| 核销记录 | GET | `/verification-records?keyword=&organizer_id=` |

## 动态管理

| 功能 | 方法 | 路径 |
|---|---|---|
| 动态列表 | GET | `/notes?status=&keyword=` |
| 更新动态状态 | PUT | `/notes/:id/status` |
| 点赞记录 | GET | `/notes/:id/records/likes` |
| 收藏记录 | GET | `/notes/:id/records/collections` |
| 评论记录 | GET | `/notes/:id/records/comments` |
| 分享记录 | GET | `/notes/:id/records/shares` |
| 评论隐藏/公开/删除 | PATCH | `/notes/:id/comments/:comment_id/status` |

更新动态状态：

```json
{
  "status": -1
}
```

## 消息管理

| 功能 | 方法 | 路径 |
|---|---|---|
| 消息列表 | GET | `/messages?target=&type=` |
| 发布消息 | POST | `/messages` |
| 消息发送/阅读记录 | GET | `/messages/:id/records` |

请求：

```json
{
  "title": "系统通知",
  "content": "消息内容",
  "type": "system",
  "target": "merchant",
  "organizer_ids": [1, 2],
  "target_user_ids": [],
  "status": 1
}
```

字段说明：

| 字段 | 说明 |
|---|---|
| target | `all` 全部用户；`merchant`/`organizer` 商家账号；`user` 指定用户 |
| type | 消息类型，例如 `system`、`notice`、`activity`、`refund` |
| organizer_ids | 当 `target=merchant` 时可指定商家 ID；不传则推送全部已通过商家 |
| target_user_ids | 当 `target=user` 时指定用户 ID |
| status | `1` 发布并推送；`0` 草稿，只落库不推送 |

推送链路：

```text
POST /api/v1/admin/messages
-> 写入 platform_messages
-> 组装 SystemMessage(type=platform_message)
-> RocketMQ topic: HYPER_SYSTEM_MSGS
-> socket/process/notice.go 消费
-> RPC/socket 推送事件 notice.platform_message
```

前端 socket 收到事件：

```json
{
  "event": "notice.platform_message",
  "data": {
    "message_id": 1,
    "target_id": 1001,
    "title": "系统通知",
    "content": "消息内容",
    "type": "system",
    "target": "merchant",
    "created_at": "2026-06-15 12:00:00"
  }
}
```

## 积分与财务

| 功能 | 方法 | 路径 |
|---|---|---|
| 积分流水 | GET | `/points/logs?user_id=` |
| 积分规则 | GET | `/points/rules` |
| 更新积分规则 | PUT | `/points/rules` |
| 财务汇总 | GET | `/finance/summary` |
| 商家对账 | GET | `/finance/settlements?organizer_id=` |
| 商家提现列表 | GET | `/withdraws?status=&organizer_id=` |
| 商家提现审核 | PUT | `/withdraws/:id/audit` |
| 收款账户审核列表 | GET | `/bank-account-audits?status=&organizer_id=` |
| 收款账户审核 | PUT | `/bank-account-audits/:id/audit` |

提现审核：

商家对账返回关键字段：

| 字段 | 说明 |
|---|---|
| gross_amount | 订单收入，单位分 |
| refund_amount | 已退款金额，单位分 |
| withdraw_amount | 已审核通过提现金额，单位分 |
| pending_withdraw_amount | 待审核提现金额，单位分 |
| settle_amount | 订单收入 - 已退款金额 |
| pending_settle_amount | 订单收入 - 已退款金额 - 已提现 - 待审核提现 |
| order_count | 订单数 |
| updated_at | 最近更新时间 |

```json
{
  "status": 1,
  "remark": "审核通过"
}
```

收款账户审核列表：

```http
GET /api/v1/admin/bank-account-audits?page=1&pageSize=20&status=0&organizer_id=10
Authorization: Bearer <admin_access_token>
```

响应：

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": 12,
        "organizer_id": 10,
        "organizer_name": "Hyper Club",
        "organizer_type": "venue",
        "user_id": 1001,
        "user_name": "商家负责人",
        "user_mobile": "13800138000",
        "bank_account_name": "张三",
        "bank_account_no": "6222000000000000000",
        "bank_name": "招商银行",
        "status": 0,
        "reject_reason": "",
        "reviewed_at": null,
        "created_at": "2026-06-15T16:20:00+08:00",
        "updated_at": "2026-06-15T16:20:00+08:00"
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 20
  }
}
```

收款账户审核：

```http
PUT /api/v1/admin/bank-account-audits/:id/audit
Authorization: Bearer <admin_access_token>
```

请求：

```json
{
  "status": 1,
  "reject_reason": ""
}
```

说明：

- `status=1` 通过：同步写入商家正式收款账户，商家随后可发起提现。
- `status=2` 拒绝：不修改商家正式收款账户，前端展示 `reject_reason` 并允许商家重新提交。
- 已审核记录不能重复审核。

状态约定：

```text
提现 status: 0 待审核, 1 通过, 2 拒绝
收款账户审核 status: 0 待审核, 1 通过, 2 拒绝
通用 status: 1 启用/正常, 0 禁用
```

## 本次新增数据表

- `admin_roles`
- `admin_operation_logs`
- `admin_categories`
- `activity_collections`
- `activity_collection_items`
- `platform_messages`
- `organizer_withdraws`
- `organizer_bank_account_audits`

部署前请执行 `config/table.sql` 中对应建表语句。
