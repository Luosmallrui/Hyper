# 平台私信开关接口

更新时间：2026-08-15

## 功能说明

平台管理端“系统配置”新增普通用户私信开关。关闭后，普通用户之间不能新发单聊消息；已配置的客服聊天账号仍可正常收发消息，保证客服入口可用。

后端会在发送消息时强制校验，前端隐藏入口仅用于体验优化，不能替代后端校验。

## 平台配置读取

```http
GET /api/v1/admin/system-config
Authorization: Bearer <admin_access_token>
```

新增字段：

```json
{
  "code": 200,
  "data": {
    "system_name": "Hyper",
    "customer_service_user_id": 77,
    "direct_message_enabled": true
  }
}
```

## 平台配置保存

```http
PUT /api/v1/admin/system-config
Authorization: Bearer <admin_access_token>
Content-Type: application/json
```

关闭普通用户私信：

```json
{
  "system_name": "Hyper",
  "icp_record_no": "蜀ICP备2026000362号",
  "customer_service_phone": "",
  "customer_service_wechat": "",
  "customer_service_email": "",
  "customer_service_hours": "",
  "customer_service_user_id": 77,
  "withdraw_arrival_cycle": "T+1 到 T+3 个工作日",
  "direct_message_enabled": false
}
```

字段未提交时，后端保留原开关值，兼容尚未升级的管理端。

也支持只更新私信开关：

```json
{ "direct_message_enabled": true }
```

## 客户端公开配置

```http
GET /api/v1/system-config
```

该公开接口响应使用 `Cache-Control: no-store`，配置保存后客户端重新请求即可拿到最新开关。

响应同样包含：

```json
{
  "code": 200,
  "data": {
    "system_name": "Hyper",
    "icp_record_no": "蜀ICP备2026000362号",
    "direct_message_enabled": false
  }
}
```

客户端可据此隐藏普通用户主页、关注列表等位置的“私信”入口；客服入口不要隐藏。

## 发送消息行为

普通单聊仍使用现有接口：

```http
POST /api/v1/message/send
Authorization: Bearer <access_token>
```

私信关闭时，普通用户之间发送单聊会返回：

```json
{
  "code": 403,
  "msg": "平台已关闭私信功能"
}
```

群聊不受此开关影响。`customer_service_user_id` 对应账号与用户之间的单聊不受此开关影响。

## 初始化 SQL

新库初始化已写入 `config/table.sql`。已有生产库执行：

```sql
INSERT INTO `platform_settings` (`setting_key`, `setting_value`, `remark`)
VALUES ('direct_message_enabled', '1', '普通用户私信开关，1开启 0关闭')
ON DUPLICATE KEY UPDATE `remark` = VALUES(`remark`);
```

修改开关也可直接执行：

```sql
UPDATE `platform_settings`
SET `setting_value` = '0'
WHERE `setting_key` = 'direct_message_enabled';
```
