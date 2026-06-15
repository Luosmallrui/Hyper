# 商家收款账户审核与提现接口说明

本文说明商家填写银行卡收款信息、后台审核、审核通过后发起提现的完整链路。

默认请求头：

```http
Authorization: Bearer <access_token>
```

## 业务流程

```text
商家入驻审核通过
  -> 商家提交收款账户
  -> 后台审核收款账户
  -> 审核通过后写入正式收款账户
  -> 商家可发起提现
  -> 后台审核提现并线下/财务打款
```

## 状态说明

### 收款账户审核状态

| 状态 | 含义 |
|---|---|
| 0 | 待审核 |
| 1 | 通过 |
| 2 | 拒绝 |

### 提现状态

| 状态 | 含义 |
|---|---|
| 0 | 待审核 |
| 1 | 通过 |
| 2 | 拒绝 |

## 商家端接口

### 获取收款账户与审核状态

```http
GET /api/v1/organizer/withdraw-info
```

响应：

```json
{
  "code": 200,
  "data": {
    "bank_account_name": "张三",
    "bank_account_no": "6222000000000000000",
    "bank_name": "招商银行",
    "can_withdraw": true,
    "pending_audit": null,
    "latest_audit": {
      "id": 12,
      "bank_account_name": "张三",
      "bank_account_no": "6222000000000000000",
      "bank_name": "招商银行",
      "status": 1,
      "reject_reason": "",
      "reviewed_at": "2026-06-15T16:30:00+08:00",
      "created_at": "2026-06-15T16:20:00+08:00",
      "updated_at": "2026-06-15T16:30:00+08:00"
    }
  }
}
```

字段说明：

| 字段 | 说明 |
|---|---|
| bank_account_name | 当前已审核通过的正式收款人 |
| bank_account_no | 当前已审核通过的正式收款账户 |
| bank_name | 当前已审核通过的正式银行名称 |
| can_withdraw | 是否可以发起提现 |
| pending_audit | 当前待审核申请；没有则为空 |
| latest_audit | 最近一次收款账户审核申请 |

### 提交收款账户审核

```http
PUT /api/v1/organizer/withdraw-info
```

请求：

```json
{
  "bank_account_name": "张三",
  "bank_account_no": "6222000000000000000",
  "bank_name": "招商银行"
}
```

成功响应：

```json
{
  "code": 200,
  "data": {
    "success": true
  }
}
```

规则：

- 只有入驻审核通过的商家可以提交。
- 提交后生成一条 `status=0` 的收款账户审核申请。
- 不会立即覆盖正式收款账户。
- 同一商家存在待审核申请时，不允许重复提交。
- 审核拒绝后，商家可以重新提交。

### 商家提现列表

```http
GET /api/v1/organizer/withdraws?page=1&size=10&status=0
```

查询参数：

| 参数 | 必填 | 说明 |
|---|---|---|
| page | 否 | 页码，默认 1 |
| size | 否 | 每页数量，默认 10 |
| status | 否 | 提现状态：`0` 待审核；`1` 通过；`2` 拒绝 |

响应：

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": 1,
        "organizer_id": 10,
        "amount": 10000,
        "bank_account_name": "张三",
        "bank_account_no": "6222000000000000000",
        "bank_name": "招商银行",
        "status": 0,
        "remark": "",
        "created_at": "2026-06-15T16:40:00+08:00",
        "updated_at": "2026-06-15T16:40:00+08:00"
      }
    ],
    "total": 1
  }
}
```

### 发起商家提现

```http
POST /api/v1/organizer/withdraws
```

请求：

```json
{
  "amount": 10000,
  "remark": "6月活动结算提现"
}
```

成功响应：

```json
{
  "code": 200,
  "data": {
    "id": 1
  }
}
```

规则：

- `amount` 单位为分。
- 只有收款账户审核通过后才能发起提现。
- 同一商家存在待审核提现申请时，不允许重复提交。
- 提现记录会快照当前正式收款账户，后续商家改卡不会影响历史提现打款信息。

## 管理端接口

管理端接口统一前缀：

```text
/api/v1/admin
```

### 收款账户审核列表

```http
GET /api/v1/admin/bank-account-audits?page=1&pageSize=20&status=0&organizer_id=10
```

查询参数：

| 参数 | 必填 | 说明 |
|---|---|---|
| page | 否 | 页码，默认 1 |
| pageSize | 否 | 每页数量，默认 20 |
| status | 否 | 审核状态：`0` 待审核；`1` 通过；`2` 拒绝 |
| organizer_id | 否 | 按商家 ID 筛选 |

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

### 审核收款账户

```http
PUT /api/v1/admin/bank-account-audits/:id/audit
```

请求：

```json
{
  "status": 1,
  "reject_reason": ""
}
```

规则：

- `status=1` 表示通过。
- `status=2` 表示拒绝。
- 通过后，系统会把申请中的 `bank_account_name/bank_account_no/bank_name` 同步写入商家正式收款账户。
- 拒绝后，不修改商家正式收款账户。
- 已审核记录不能重复审核。

成功响应：

```json
{
  "code": 200,
  "data": {
    "success": true
  }
}
```

### 商家提现列表

```http
GET /api/v1/admin/withdraws?page=1&pageSize=20&status=0&organizer_id=10
```

### 审核商家提现

```http
PUT /api/v1/admin/withdraws/:id/audit
```

请求：

```json
{
  "status": 1,
  "remark": "审核通过，已打款"
}
```

说明：

- `status=1` 通过。
- `status=2` 拒绝。
- 当前接口只记录审核结果和备注，真实打款仍由财务线下或后续支付通道处理。

## 数据库变更

新增表：

```sql
CREATE TABLE IF NOT EXISTS `organizer_bank_account_audits`
(
    `id`                bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '收款账户审核ID',
    `organizer_id`      bigint unsigned NOT NULL COMMENT '主办方ID',
    `user_id`           bigint unsigned NOT NULL COMMENT '提交用户ID',
    `bank_account_name` varchar(50)     NOT NULL COMMENT '收款人',
    `bank_account_no`   varchar(50)     NOT NULL COMMENT '收款账户',
    `bank_name`         varchar(50)     NOT NULL COMMENT '银行名称',
    `status`            tinyint         NOT NULL DEFAULT 0 COMMENT '0待审核 1通过 2拒绝',
    `reject_reason`     varchar(255)    NOT NULL DEFAULT '' COMMENT '拒绝原因',
    `reviewed_at`       datetime        NULL COMMENT '审核时间',
    `created_at`        datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`        datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_bank_audit_organizer` (`organizer_id`) USING BTREE,
    KEY `idx_bank_audit_user` (`user_id`) USING BTREE,
    KEY `idx_bank_audit_status` (`status`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='主办方收款账户审核表';
```

已有提现表迁移：

```sql
ALTER TABLE `organizer_withdraws` ADD COLUMN `bank_account_name` varchar(50) NOT NULL DEFAULT '' COMMENT '收款人快照' AFTER `amount`;
ALTER TABLE `organizer_withdraws` ADD COLUMN `bank_account_no` varchar(50) NOT NULL DEFAULT '' COMMENT '收款账户快照' AFTER `bank_account_name`;
ALTER TABLE `organizer_withdraws` ADD COLUMN `bank_name` varchar(50) NOT NULL DEFAULT '' COMMENT '银行名称快照' AFTER `bank_account_no`;
```

## 前端接入建议

1. 商家后台进入提现设置页时，先调用 `GET /api/v1/organizer/withdraw-info`。
2. 如果 `pending_audit.status=0`，展示“审核中”，禁用重复提交。
3. 如果 `latest_audit.status=2`，展示拒绝原因并允许重新提交。
4. 如果 `can_withdraw=true`，提现按钮可用。
5. 管理端新增“收款账户审核”列表，默认筛选 `status=0`。
6. 管理员审核通过后，商家端再次查询 `withdraw-info` 即可看到正式账户和 `can_withdraw=true`。
