# 2026-07-17 后端接口确认与补齐说明

本文以当前后端实现为准，供小程序、商家 PC 与管理端 PC 接入。

## 1. 核销员绑定

### 获取扫码确认信息

```http
GET /api/v1/verifier/activation-info?v={verifier_id}
```

该接口无需登录，用于小程序扫码页展示核销员和主办方信息。

```json
{
  "code": 200,
  "data": {
    "verifier_id": 12,
    "name": "张三",
    "phone": "13800000000",
    "status": 0,
    "is_bound": false,
    "organizer_id": 3,
    "organizer_name": "测试主办方"
  }
}
```

### 确认绑定

```http
POST /api/v1/verifier/activate
Authorization: Bearer {user_access_token}
Content-Type: application/json
```

规范请求：

```json
{
  "verifier_id": 12,
  "phone": "13800000000",
  "channel": "wechat"
}
```

- `phone` 必填，且必须和当前登录小程序账号手机号一致。
- `verifier_id` 可选；扫码场景应传入，历史仅传 `phone` 的客户端仍兼容。
- `channel` 可选，缺省按 `wechat` 处理。
- 已废弃的 `code` 字段不再使用。

## 2. 账户资料

新增已登录用户资料接口，供 PC 重置密码页读取已绑定手机号：

```http
GET /api/v1/auth/profile
Authorization: Bearer {access_token}
```

```json
{
  "code": 200,
  "data": {
    "user_id": 35,
    "phone": "13800000000",
    "nickname": "用户昵称",
    "avatar": "https://cdn.example.com/avatar.png"
  }
}
```

## 3. 上传类型

```http
POST /api/v1/upload
Content-Type: multipart/form-data
```

`type` 允许值新增：

```text
collection_share
post
```

完整支持：`poster_detail`、`poster_long`、`poster_list`、`poster_wechat`、`qualification_doc`、`avatar`、`organizer_logo`、`collection_share`、`post`、`misc`。

## 4. 经营与票种字段

### 管理端控制台

```http
GET /api/v1/admin/dashboard
```

新增 `total_merchants`，统计已认证主办方数量。

### 商家财务汇总

```http
GET /api/v1/organizer/finance/summary
```

新增字段，金额单位均为分：

```json
{
  "today_order_count": 3,
  "today_order_amount": 12800,
  "today_ticket_count": 5
}
```

今日数据按订单支付时间和已支付相关订单状态统计。

### 活动票种

`GET /api/v1/activity/:id/ticket-specs` 的票种对象新增 `description`。

保存票种时同样提交：

```http
POST /api/v1/activity/:id/ticket-specs
```

```json
{
  "specs": [
    {
      "id": 0,
      "name": "早鸟票",
      "description": "含一杯指定饮品，数量有限",
      "is_enabled": 1,
      "sale_start": "2026-07-20T10:00:00+08:00",
      "sale_end": "2026-07-25T18:00:00+08:00",
      "price": 9900,
      "stock": 100,
      "purchase_limit": 2,
      "max_attendees": 1
    }
  ]
}
```

## 5. 收款资料审核

```http
PUT /api/v1/organizer/withdraw-info
Authorization: Bearer {organizer_access_token}
```

```json
{
  "bank_account_name": "收款主体名称",
  "bank_account_no": "6222020000000000",
  "bank_name": "中国工商银行",
  "contact_name": "张三",
  "contact_phone": "13800000000"
}
```

`GET /api/v1/organizer/withdraw-info`、商家提交的审核记录以及管理端银行账户审核列表都会返回 `contact_name`、`contact_phone`。管理端审核通过后，这两个字段会同步为主办方当前生效收款资料。两个联系人字段对存量已审核账户保持兼容，不影响已有账户提现。

## 6. 管理员昵称

管理员个人资料接口支持昵称：

```http
GET /api/v1/admin/profile
PUT /api/v1/admin/profile
```

```json
{
  "nickname": "运营管理员"
}
```

`GET /api/v1/admin/admins` 也会返回每个管理员的 `nickname`；创建/编辑管理员请求可提交同名字段。

## 7. 管理端核销员维护

除原有列表和状态开关外，新增完整维护接口：

```http
POST   /api/v1/admin/verifiers
PUT    /api/v1/admin/verifiers/:id
DELETE /api/v1/admin/verifiers/:id
```

创建和编辑请求：

```json
{
  "organizer_id": 3,
  "user_id": 0,
  "name": "张三",
  "phone": "13800000000",
  "status": 0,
  "permission_scope": "活动",
  "channel": "wechat"
}
```

- `user_id = 0` 表示未绑定小程序账号。
- 已绑定用户且启用时会记录绑定时间。
- 有核销记录的核销员不能删除，防止核销审计链断裂；可改为停用。
- 保留现有状态接口：`PATCH /api/v1/admin/verifiers/:id/status`。

## 8. 数据库迁移

部署前对已有数据库执行 `config/table.sql` 中新增的四组 `ALTER TABLE`：

```sql
ALTER TABLE `admin` ADD COLUMN `nickname` varchar(50) NOT NULL DEFAULT '' COMMENT '展示昵称' AFTER `username`;
ALTER TABLE `organizers` ADD COLUMN `bank_contact_name` varchar(50) NOT NULL DEFAULT '' COMMENT '收款联系人' AFTER `bank_name`;
ALTER TABLE `organizers` ADD COLUMN `bank_contact_phone` varchar(20) NOT NULL DEFAULT '' COMMENT '收款联系电话' AFTER `bank_contact_name`;
ALTER TABLE `ticket_specs` ADD COLUMN `description` varchar(500) NOT NULL DEFAULT '' COMMENT '票种说明' AFTER `name`;
ALTER TABLE `organizer_bank_account_audits` ADD COLUMN `bank_contact_name` varchar(50) NOT NULL DEFAULT '' COMMENT '收款联系人' AFTER `bank_name`;
ALTER TABLE `organizer_bank_account_audits` ADD COLUMN `bank_contact_phone` varchar(20) NOT NULL DEFAULT '' COMMENT '收款联系电话' AFTER `bank_contact_name`;
```

生产执行前请先确认目标列不存在，避免重复执行报错。
