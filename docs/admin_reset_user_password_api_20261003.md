# 管理端重置用户密码（2026-10-03）

```http
PUT /api/v1/admin/users/:id/password
Content-Type: application/json

{"password": "新密码（不少于 6 位）"}
```

- 管理端直接设置/重置任意用户（含商家）的登录密码，**免短信验证**，用于商家无法走短信流程时后台兜底。
- 需要管理员登录态，权限沿用 `admin.user.*`；操作记入管理端操作日志（`admin.user.reset_password`）。
- 成功返回 `{"code":200,"data":{"success":true}}`；用户不存在返回 400。
- 按手机号查用户 ID 可先用 `GET /v1/admin/users?keyword=手机号` 搜索。
