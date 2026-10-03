# 笔记 status 枚举与 C 端展示口径（2026-10-03）

## status 枚举

| 值 | 含义 | 备注 |
|---|---|---|
| `0` | 审核中 / 默认 | 新发布笔记初始值；管理端「隐藏」也写 0（连带 `visible_conf=3`） |
| `1` | 公开 | 管理端「公开」写 1（连带 `visible_conf=1`） |
| `2` | 私密 | 文档定义，实际私密由 `visible_conf` 控制 |
| `3` | 违规 | 文档定义 |
| `-1` | 已删除 | 软删，用户自删也是 -1 |

## visible_conf 枚举

| 值 | 含义 |
|---|---|
| `1` | 公开 |
| `2` | 粉丝可见 |
| `3` | 仅自己可见 |

## C 端展示条件（唯一口径）

```sql
notes.status <> -1 AND notes.visible_conf = 1
```

## 前端标签判断建议

```
status = -1                    → 已删除
status = 0 且 visible_conf ≠ 1 → 已隐藏
status = 0 且 visible_conf = 1 → 审核中（C 端可见）
status = 1                     → 公开
status = 3                     → 违规
```

## 管理端操作

`PUT /api/v1/admin/notes/:id/status`，body `{"status": 0|1|-1}` = 隐藏 / 公开 / 删除。
路径里的 `:id` 使用列表返回的 `id_str`（雪花 ID 字符串，避免 JS 精度丢失）。
