# 管理端内容与消息管理后端更新

更新时间：2026-07-12

已完成：

- 新增全局互动日志 `GET /api/v1/admin/note-interactions`，支持互动类型、动态、用户、关键词、分享渠道和时间筛选。
- 新增全局评论审核队列 `GET /api/v1/admin/note-comments` 与 `PATCH /api/v1/admin/note-comments/:comment_id/status`。
- 评论状态确认：`1=公开`、`0=隐藏`、`-1=软删除`。审核操作记录管理员、时间和原因；不会物理删除评论。
- 消息列表补充创建管理员、渠道及投递/阅读聚合数据。
- 消息投递记录补充接收对象、脱敏手机号、投递状态、失败原因与阅读状态，并支持 `delivery_status`、`read_status` 筛选。

部署前执行 [config/table.sql](/Users/luosmallrui/Hyper/config/table.sql) 中 `comments` 与 `platform_messages` 的迁移语句。

完整请求与响应契约见 [admin_pc_api.md](/Users/luosmallrui/Hyper/docs/admin_pc_api.md)。
