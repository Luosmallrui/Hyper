# 管理端系统配置接口

更新时间：2026-07-12

管理端“系统配置”页面统一维护平台基础信息、客服信息及商家提现到账周期。配置底层存储于 `platform_settings`，前端不需要管理零散的设置键。

## 权限

以下接口均需要管理员 Token 和 `admin.system` 权限：

```http
Authorization: Bearer <admin_access_token>
```

## 登录页公开配置

```http
GET /api/v1/system-config
```

该接口无需登录，仅返回可公开展示的系统名称与备案号：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "system_name": "Hyper",
    "icp_record_no": "蜀ICP备2026000362号"
  }
}
```

- 不返回客服信息、提现周期、积分规则或任何管理员配置。
- 响应头包含 `Cache-Control: public, max-age=300`。
- 配置读取失败时，登录页应自行回退显示 `Hyper © 2020-2026`。

## 获取系统配置

```http
GET /api/v1/admin/system-config
```

响应：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "system_name": "Hyper 潮流活动平台",
    "icp_record_no": "蜀ICP备2026000000号",
    "customer_service_phone": "400-000-0000",
    "customer_service_wechat": "hyper_service",
    "customer_service_email": "service@hypercn.cn",
    "customer_service_hours": "工作日 09:00-18:00",
    "withdraw_arrival_cycle": "T+1 到 T+3 个工作日"
  }
}
```

## 更新系统配置

```http
PUT /api/v1/admin/system-config
Content-Type: application/json
```

请求：

```json
{
  "system_name": "Hyper 潮流活动平台",
  "icp_record_no": "蜀ICP备2026000000号",
  "customer_service_phone": "400-000-0000",
  "customer_service_wechat": "hyper_service",
  "customer_service_email": "service@hypercn.cn",
  "customer_service_hours": "工作日 09:00-18:00",
  "withdraw_arrival_cycle": "T+1 到 T+3 个工作日"
}
```

成功响应：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "success": true
  }
}
```

## 字段说明

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| `system_name` | 是 | 平台系统名称，最大 100 字符 |
| `icp_record_no` | 否 | ICP 备案号 |
| `customer_service_phone` | 否 | 客服电话 |
| `customer_service_wechat` | 否 | 客服微信号 |
| `customer_service_email` | 否 | 客服邮箱 |
| `customer_service_hours` | 否 | 客服服务时间 |
| `withdraw_arrival_cycle` | 否 | 商家端展示的预计到账周期文案 |

所有字段提交时应完整传递；空字符串表示清空该项配置。

## 审计与错误处理

- 更新成功会写入管理员操作日志：`action=admin.settings.update`、`resource_type=settings`。
- 无 `admin.system` 权限时返回 HTTP `403`，错误码为 `ADMIN_PERMISSION_DENIED`。
- `system_name` 为空或任一字段超过长度限制时，返回业务错误。

## 初始化配置

部署时执行 [config/table.sql](/Users/luosmallrui/Hyper/config/table.sql) 中 `platform_settings` 的 `INSERT IGNORE` 默认数据。已有数据库可单独执行：

```sql
INSERT IGNORE INTO platform_settings (setting_key, setting_value, remark) VALUES
('system_name', 'Hyper', '平台系统名称'),
('icp_record_no', '', 'ICP备案号'),
('customer_service_phone', '', '客服电话'),
('customer_service_wechat', '', '客服微信'),
('customer_service_email', '', '客服邮箱'),
('customer_service_hours', '', '客服服务时间');
```
