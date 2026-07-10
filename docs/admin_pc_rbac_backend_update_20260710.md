# 管理端 PC RBAC 后端接口与上线说明

更新时间：2026-07-10

## 已实现内容

- 所有 `/api/v1/admin/*` 路由在管理员 Token 校验后，实时查询管理员状态、角色状态和角色权限。
- `role_id = 0` 的管理员不再默认拥有全部权限；只有个人资料接口可访问。
- 权限采用模块级白名单：`admin.dashboard`、`admin.system`、`admin.users`、`admin.merchants`、`admin.organizers`、`admin.activities`、`admin.tickets`、`admin.orders`、`admin.verifications`、`admin.content`、`admin.finance`；`["*"]` 为超级管理员。
- 管理员与角色的创建、修改、删除仅超级管理员可调用。
- 权限拒绝写入 `admin_operation_logs`；管理员、角色、积分调整等写操作继续由现有操作日志中间件记录。
- 管理员被禁用、角色被禁用或角色权限被调整后，下一次 API 请求立即生效，无需等待 Token 过期。
- 禁止删除关联管理员的角色；禁止删除、停用或降权最后一个启用的超级管理员；禁止管理员删除或停用自己。

## 权限映射

| 路由 | 所需权限 |
| --- | --- |
| `/admin/dashboard` | `admin.dashboard` |
| `/admin/settings`、`/categories`、`/logs`、管理员/角色查询 | `admin.system` |
| 管理员/角色写操作 | `*` |
| `/admin/users`、`/viewers` | `admin.users` |
| `/admin/organizers` | `admin.organizers` |
| `/admin/organizer-level-rules` | `admin.merchants` |
| `/admin/activities`、`/parties`、`/activity-collections` | `admin.activities` |
| `/admin/tickets`、`/events/:id/tickets` | `admin.tickets` |
| `/admin/orders` | `admin.orders` |
| `/admin/verifiers`、`/verification-records` | `admin.verifications` |
| `/admin/notes`、`/messages`、`/banners` | `admin.content` |
| `/admin/finance`、`/withdraws`、`/bank-account-audits`、`/points` | `admin.finance` |

`GET/PUT /admin/profile` 与 `PUT /admin/profile/password` 仅要求管理员账号处于启用状态。

## 管理员和角色响应变化

`GET /api/v1/admin/profile` 新增：

```json
{
  "role_id": 2,
  "role_name": "运营",
  "permissions": ["admin.dashboard", "admin.activities"]
}
```

`GET /api/v1/admin/admins` 的每一项新增角色摘要：

```json
{
  "role_id": 2,
  "role": {
    "id": 2,
    "name": "运营",
    "permissions": ["admin.dashboard", "admin.activities"]
  }
}
```

`GET /api/v1/admin/roles` 的 `permissions` 统一为数组，并新增 `member_count`。

## 错误响应

无模块权限时返回 HTTP 403：

```json
{
  "code": 403,
  "msg": "无权执行该操作",
  "error_code": "ADMIN_PERMISSION_DENIED",
  "data": {
    "required_permission": "admin.finance"
  }
}
```

