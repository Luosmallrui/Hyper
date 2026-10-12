# 管理端商家子账号管理接口（2026-10-04）

对应前端需求 `frontend_api_requirements_20261004.md`。管理端代商家维护子账号，均需管理员 Token，权限沿用 `admin.organizer.*` 映射（`/v1/admin/organizers` 前缀），操作记入管理端操作日志。

## 角色列表

```http
GET /api/v1/admin/organizers/:id/roles
```

响应：`{"list": [{"id","name","description","permissions","status","staff_count"}]}`

## 子账号列表

```http
GET /api/v1/admin/organizers/:id/staff?page=1&pageSize=20&keyword=
```

响应：`{"list": [{"id","user_id","name","phone","role_id","role_name","status","created_at"}], "total": N}`，keyword 匹配姓名/手机号。

## 新增子账号

```http
POST /api/v1/admin/organizers/:id/staff
{"phone": "15082543853", "name": "张瀚", "role_id": 3, "status": 1}
```

- 后端按 `phone` 解析 `user_id`；手机号未注册平台账号返回 400「该手机号未注册平台账号」
- `role_id` 必须为该企业家的角色，否则 400「角色不存在」
- `status` 缺省为 1（启用）

## 编辑 / 删除

```http
PUT    /api/v1/admin/organizers/:id/staff/:staffId   # body 同新增
DELETE /api/v1/admin/organizers/:id/staff/:staffId
```

子账号不存在返回 404。

实现：`service/admin.go` `AdminListOrganizerRoles` / `AdminListOrganizerStaff` / `AdminSaveOrganizerStaff` / `AdminDeleteOrganizerStaff`；路由 `handler/admin.go`。
