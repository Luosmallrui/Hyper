# 客户端与网页端 0624 后端更新说明

本文对应 `client_web_requirements_20260624.md`，记录本次后端已完成内容、接口变化、数据库迁移和前端仍需处理的事项。

默认接口前缀为 `/api/v1`。

## 1. 本次已完成

### 1.1 入驻状态与商家启停分离

`organizers.status` 继续表示入驻审核状态：

| status | 含义 |
|---|---|
| 0 | 未提交/待提交 |
| 1 | 审核中 |
| 2 | 已通过 |
| 3 | 已驳回 |

新增 `organizers.enabled`：

| enabled | 含义 |
|---|---|
| 1 | 商家账号可使用 |
| 0 | 商家账号已停用 |

管理员启停商家：

```http
PATCH /api/v1/admin/organizers/:id/status
Authorization: Bearer <admin_token>
Content-Type: application/json
```

```json
{
  "enabled": 0
}
```

限制：

- 只有 `status=2` 的商家可以重新启用。
- 商家被停用后，商家业务接口返回“商家账号已停用”。
- 停用不会删除商家、活动、订单和财务历史。

入驻状态接口补充：

```http
GET /api/v1/organizer/audit-status
```

```json
{
  "code": 200,
  "data": {
    "organizer_id": 3,
    "type": "merchant",
    "status": 2,
    "enabled": 1,
    "reject_reason": "",
    "submitted_at": "2026-06-24T10:00:00+08:00",
    "reviewed_at": "2026-06-24T10:10:00+08:00"
  }
}
```

说明：

- 审核中重复提交仍会被后端拒绝。
- 已驳回申请允许修改资料后重新提交。
- `GET /organizer/info` 可返回待审核、已驳回和已停用商家的状态信息。
- 其它商家业务接口要求商家已通过且已启用。

### 1.2 活动列表筛选

管理端：

```http
GET /api/v1/admin/activities
```

商家端：

```http
GET /api/v1/activity/my-list
```

支持参数：

| 参数 | 说明 |
|---|---|
| keyword | 活动名称或地址 |
| status | 活动状态 |
| organizer_id | 主办方 ID，仅管理端 |
| published_from | 发布时间起点 |
| published_to | 发布时间终点 |
| activity_from | 活动时间范围起点 |
| activity_to | 活动时间范围终点 |
| page | 页码 |
| pageSize / size | 每页数量 |

时间支持：

```text
2026-06-24
2026-06-24 10:00:00
2026-06-24T10:00:00+08:00
```

活动状态：

| status | 含义 |
|---|---|
| 0 | 草稿 |
| 1 | 已提交/待处理 |
| 2 | 审核中 |
| 3 | 已通过并上架 |
| 4 | 已驳回 |

未传 `status` 时，管理端仍默认排除草稿。

### 1.3 管理员角色和权限

管理员新增、编辑请求增加：

```json
{
  "username": "ops",
  "password": "123456",
  "role_id": 2,
  "status": 1
}
```

规则：

- `role_id=0`：超管兼容模式，不做模块权限限制。
- `role_id>0`：角色必须存在且启用。
- 已绑定角色的管理员登录后，会按角色 `permissions` 校验接口权限。

`permissions` 支持以下形式：

```json
["*"]
```

```json
["GET:/v1/admin/users", "PUT:/v1/admin/users/:id/status"]
```

```json
["admin.users", "admin.orders", "admin.messages"]
```

模块权限允许访问对应模块的全部方法；方法加路径适合精确到按钮级权限。

### 1.4 管理员操作日志

所有成功执行的管理端 `POST`、`PUT`、`PATCH`、`DELETE` 请求会自动写入：

```http
GET /api/v1/admin/logs
```

记录字段包括管理员、动作、资源、请求方法、路径、IP、响应状态和操作时间。

### 1.5 动态评论治理

查询评论记录：

```http
GET /api/v1/admin/notes/:note_id/records/comments
```

修改评论状态：

```http
PATCH /api/v1/admin/notes/:note_id/comments/:comment_id/status
```

```json
{
  "status": 0
}
```

状态：

| status | 含义 |
|---|---|
| -1 | 删除 |
| 0 | 隐藏 |
| 1 | 公开 |

动态本身继续使用：

```http
PUT /api/v1/admin/notes/:id/status
```

### 1.6 平台消息发送和阅读记录

发布接口保持不变：

```http
POST /api/v1/admin/messages
```

发布后，后端会为每个目标用户写入 `platform_message_deliveries`，并记录 RocketMQ 发送结果。

查看发送记录：

```http
GET /api/v1/admin/messages/:id/records?page=1&pageSize=20&status=1
```

投递状态：

| status | 含义 |
|---|---|
| 0 | 待发送 |
| 1 | 已发送 |
| 2 | 发送失败 |
| 3 | 已读 |

商家端阅读接口：

```http
POST /api/v1/organizer/messages/:id/read
POST /api/v1/organizer/messages/read-all
```

调用后会同步更新平台消息投递记录的阅读状态。

定向发布给指定 `organizer_ids` 时，只有对应商家能在消息列表看到该消息。历史消息没有投递明细时仍按旧规则兼容展示。

### 1.7 积分规则配置

客户端读取积分规则：

```http
GET /api/v1/points/rules
Authorization: Bearer <access_token>
```

```json
{
  "code": 200,
  "data": {
    "discount_cents_per_point": 10,
    "reward_cents_per_point": 1000
  }
}
```

字段说明：

- `discount_cents_per_point=10`：1 积分抵扣 10 分，即 0.1 元。
- `reward_cents_per_point=1000`：每消费 1000 分，即 10 元，奖励 1 积分。

管理端：

```http
GET /api/v1/admin/points/rules
PUT /api/v1/admin/points/rules
```

```json
{
  "discount_cents_per_point": 10,
  "reward_cents_per_point": 1000
}
```

创建订单抵扣和支付成功返积分现在统一读取此配置，不再使用两套固定值。

## 2. 数据库迁移

生产环境部署代码前需要执行：

```sql
ALTER TABLE `organizers`
ADD COLUMN `enabled` tinyint NOT NULL DEFAULT 1 COMMENT '1启用 0停用' AFTER `status`;

ALTER TABLE `admin`
ADD COLUMN `role_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '后台角色ID' AFTER `motto`;

ALTER TABLE `admin`
ADD KEY `idx_admin_role` (`role_id`);
```

创建消息投递记录表：

```sql
CREATE TABLE IF NOT EXISTS `platform_message_deliveries`
(
    `id`         bigint unsigned NOT NULL AUTO_INCREMENT,
    `message_id` bigint unsigned NOT NULL,
    `user_id`    bigint unsigned NOT NULL,
    `status`     tinyint         NOT NULL DEFAULT 0,
    `sent_at`    datetime        NULL,
    `read_at`    datetime        NULL,
    `error`      varchar(255)    NOT NULL DEFAULT '',
    `created_at` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_message_user` (`message_id`, `user_id`),
    KEY `idx_delivery_user` (`user_id`),
    KEY `idx_delivery_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

初始化积分规则：

```sql
INSERT IGNORE INTO `platform_settings`
(`setting_key`, `setting_value`, `remark`)
VALUES
('points_discount_cents_per_point', '10', '每积分可抵扣金额，单位分'),
('points_reward_cents_per_point', '1000', '消费多少分奖励1积分');
```

完整建表内容已同步到 `config/table.sql`。

## 3. 前端仍需修改

以下属于小程序或网页端交互，不需要新增后端接口：

1. 删除个人中心“账号中心”入口。
2. “我要入驻”只负责申请和审核状态。
3. “管理后台/主办中心”不能用当前用户的 `audit-status` 作为唯一拦截条件。
4. 找回密码弹窗增加发送验证码按钮、倒计时、发送中态和错误提示。
5. 管理端把已有 settings、categories、admins、roles 接口接入新增、编辑、删除按钮。
6. 商家等级规则在商家端只读，写操作只放平台管理端。
7. 首页暗色地图需要腾讯地图自定义地图能力，属于地图服务与前端配置。

## 4. 仍需联调验证

以下项目不是本次新增接口，需使用线上环境和真实数据验证：

1. 场地详情补充稳定测试数据。
2. WebSocket 断线重连、离线消息、未读数一致性。
3. 入驻通过、驳回通知是否通过 RocketMQ 和 Socket 正常到达。
4. 指定商家平台消息是否只对目标商家可见。
5. 管理员角色权限是否与网页端菜单和按钮权限标识一致。

## 5. 本次涉及代码

- `handler/admin.go`
- `handler/ticketing.go`
- `service/admin.go`
- `service/admin_pc.go`
- `service/ticketing.go`
- `service/pay.go`
- `models/auth.go`
- `models/ticketing.go`
- `models/admin_pc.go`
- `types/admin.go`
- `types/admin_pc.go`
- `types/ticketing.go`
- `types/points.go`
- `config/table.sql`

