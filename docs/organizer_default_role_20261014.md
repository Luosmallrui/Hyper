# 商家默认角色自动创建（2026-10-14）

## 规则

商家入驻**审核通过**时，若该商家还没有任何角色，自动创建默认角色：

```json
{
  "name": "核销员",
  "description": "默认角色，可执行扫码核销等基础操作",
  "permissions": ["verifier.manage"],
  "status": 1
}
```

## 幂等与边界

- 只在 `AuditOrganizer`（管理端入驻审核通过）时触发；商家已有任意角色则跳过，重复审核不会重复创建
- **不挂任何列表钩子**：商家后续自行删除/修改该角色不会被重建
- 权限集合按现有商户端权限模型（`verifier.manage` 核销管理，见商户端「权限设置」页权限点）

## 存量补齐（已于 2026-10-14 执行）

```sql
INSERT INTO organizer_roles (organizer_id, name, description, permissions, status, created_at, updated_at)
SELECT o.id, '核销员', '默认角色，可执行扫码核销等基础操作', '["verifier.manage"]', 1, NOW(), NOW()
FROM organizers o
WHERE NOT EXISTS (SELECT 1 FROM organizer_roles r WHERE r.organizer_id = o.id);
```

执行结果：补齐 8 家无角色商家；已有角色的商家（如 OTALAB）未受影响。

## 实现

`service/ticketing.go` `ensureDefaultOrganizerRole`；挂钩 `service/admin.go` `AuditOrganizer`。
