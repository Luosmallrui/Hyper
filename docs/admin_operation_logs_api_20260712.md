# 管理端管理员操作日志接口

更新时间：2026-07-12

## 查询列表

```http
GET /api/v1/admin/logs?page=1&pageSize=20&admin_id=1&action=admin.settings.update&resource_type=settings&result=success&start_date=2026-07-01&end_date=2026-07-12&keyword=提现
Authorization: Bearer <admin_access_token>
```

查询参数：

| 参数 | 说明 |
| --- | --- |
| `page` / `pageSize` | 分页，`pageSize` 最大 100 |
| `admin_id` | 操作管理员 ID |
| `action` | 业务动作编码，例如 `admin.settings.update` |
| `resource_type` | `settings`、`role`、`withdraw`、`activity` 等 |
| `result` | `success`、`denied`、`failed` |
| `start_date` / `end_date` | `YYYY-MM-DD`，结束日期包含当天 |
| `keyword` | 搜索动作、资源 ID/名称、备注、错误信息、管理员账号 |

响应：

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": 1001,
        "admin_id": 1,
        "admin_name": "admin",
        "admin_username": "admin",
        "action": "admin.settings.update",
        "action_name": "更新系统配置",
        "resource_type": "settings",
        "resource_id": "",
        "resource_name": "系统配置",
        "method": "PUT",
        "path": "/api/v1/admin/settings",
        "remark": "更新系统配置",
        "result": "success",
        "error_code": "",
        "error_message": "",
        "ip": "171.213.185.29",
        "created_at": "2026-07-12T12:16:00+08:00"
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 20
  }
}
```

## 动作约定

- 系统配置：`admin.settings.update`
- 分类：`admin.category.create`、`admin.category.update`、`admin.category.delete`
- 管理员与角色：`admin.account.*`、`admin.role.*`
- 入驻与活动审核：`admin.organizer.approve/reject`、`admin.activity.approve/reject`
- 售后与提现：`admin.refund.approve/reject`、`admin.withdraw.approve/reject`
- 积分调整：`admin.points.adjust`
- 权限拒绝：`admin.permission.denied`

`result=denied` 表示 RBAC 拒绝；该记录会保留所需权限、请求路径和管理员 ID。`result=failed` 表示已通过权限校验但业务处理失败，`error_code` 与 `error_message` 用于定位。

## 部署迁移

先执行 [config/table.sql](/Users/luosmallrui/Hyper/config/table.sql) 中 `admin_operation_logs` 的字段和索引迁移，再部署后端。旧日志没有业务动作快照，会按旧字段兼容查询；新日志从部署后开始完整记录。
