# 已上架活动修改二审与延后发布

更新时间：2026-08-14

## 行为

已上架活动或场地修改后，不再将公开活动从 `status=3` 改为待审核。

- 地图、首页、搜索、公开详情、购票和订阅继续展示审核前的线上版本。
- 修改内容保存为待发布快照，状态为 `pending_revision_status=1`。
- 主办方本人查询活动详情时可看到待审核的编辑稿，并收到 `has_pending_revision=true`。
- 管理端 `GET /api/v1/admin/activities?status=1` 会显示该记录，`audit_type=re_audit`。
- 管理员通过后，后端在事务内发布快照中的活动字段、票券、标签和场地经营时间；活动 ID、订单、订阅、核销和统计数据不变。
- 管理员驳回后，旧线上版本继续可见，修改稿保留；主办方本人仍收到 `has_pending_revision=true` 和 `pending_revision_reason`，可据此展示“修改被驳回”。

## 数据库迁移

部署前执行：

```sql
ALTER TABLE `activities`
  ADD COLUMN `pending_revision` MEDIUMTEXT NOT NULL COMMENT '已上架活动待审核修改快照(JSON)' AFTER `reject_reason`,
  ADD COLUMN `pending_revision_status` TINYINT NOT NULL DEFAULT 0 COMMENT '待审核修改状态: 0无 1待审核 4驳回' AFTER `pending_revision`,
  ADD COLUMN `pending_revision_reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '待审核修改驳回原因' AFTER `pending_revision_status`,
  ADD INDEX `idx_activity_pending_revision` (`pending_revision_status`);
```

若字段或索引已经存在，跳过对应语句。

## 前端字段

主办方活动列表、主办方本人活动详情会返回：

```json
{
  "status": 3,
  "has_pending_revision": true,
  "pending_revision_reason": ""
}
```

`status=3` 表示旧线上版本仍在展示。`has_pending_revision=true` 且 `pending_revision_reason` 为空时显示“修改审核中”；`has_pending_revision=true` 且 `pending_revision_reason` 非空时显示“修改被驳回”。不要因为存在二审就把该活动从公开列表、地图或购票页移除。
