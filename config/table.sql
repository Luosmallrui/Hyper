
CREATE TABLE IF NOT EXISTS `users`
(
    `id`         int unsigned     NOT NULL AUTO_INCREMENT COMMENT '用户ID',
    `mobile`     varchar(11)      NOT NULL DEFAULT '' COMMENT '手机号',
    `nickname`   varchar(64)      NOT NULL DEFAULT '' COMMENT '用户昵称',
    `avatar`     varchar(255)     NOT NULL DEFAULT '' COMMENT '用户头像',
    `gender`     tinyint unsigned NOT NULL DEFAULT '3' COMMENT '用户性别[1:男 ;2:女;3:未知]',
    `status`     tinyint          NOT NULL DEFAULT 1 COMMENT '用户状态: 1正常 0封禁',
    `password`   varchar(255)     NOT NULL COMMENT '用户密码',
    `motto`      varchar(500)     NOT NULL DEFAULT '' COMMENT '用户座右铭',
    `email`      varchar(30)      NOT NULL DEFAULT '' COMMENT '用户邮箱',
    `birthday`   varchar(10)      NOT NULL DEFAULT '' COMMENT '生日',
    `is_robot`   tinyint unsigned NOT NULL DEFAULT '2' COMMENT '是否机器人[1:是;2:否;]',
    `created_at` datetime         NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '注册时间',
    `updated_at` datetime         NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_mobile` (`mobile`) USING BTREE,
    KEY `idx_created_at` (`created_at`) USING BTREE,
    KEY `idx_updated_at` (`updated_at`) USING BTREE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci COMMENT ='用户表';;

-- 如已有 users 表，执行以下 ALTER 添加 status 字段:
-- ALTER TABLE `users` ADD COLUMN `status` tinyint NOT NULL DEFAULT 1 COMMENT '用户状态: 1正常 0封禁' AFTER `gender`;

CREATE TABLE IF NOT EXISTS `user_follow`
(
    `id`          bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
    `follower_id` bigint unsigned NOT NULL COMMENT '关注人',
    `followee_id` bigint unsigned NOT NULL COMMENT '被关注人',
    `status`      tinyint         NOT NULL DEFAULT 1 COMMENT '1:关注 0:取消',
    `created_at`  datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`  datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_follow_pair` (`follower_id`, `followee_id`) USING BTREE,
    KEY `idx_followee` (`followee_id`) USING BTREE,
    KEY `idx_follower` (`follower_id`) USING BTREE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci COMMENT ='用户关注关系表';

-- 内容关注与用户关注分离：activity / venue / party。
CREATE TABLE IF NOT EXISTS `content_follows`
(
    `id`          bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
    `user_id`     bigint unsigned NOT NULL COMMENT '关注用户ID',
    `target_type` varchar(20)     NOT NULL COMMENT 'activity/venue/party',
    `target_id`   bigint unsigned NOT NULL COMMENT '活动ID、场地主办方ID或旧派对ID',
    `created_at`  datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`  datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_content_follow` (`user_id`, `target_type`, `target_id`) USING BTREE,
    KEY `idx_content_follow_target` (`target_type`, `target_id`) USING BTREE,
    KEY `idx_content_follow_user` (`user_id`) USING BTREE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci COMMENT ='活动场地派对关注关系表';

CREATE TABLE IF NOT EXISTS `user_stats`
(
    `id`               bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
    `user_id`          int unsigned    NOT NULL COMMENT '用户ID',
    `follower_count`   int unsigned    NOT NULL DEFAULT 0 COMMENT '粉丝数',
    `following_count`  int unsigned    NOT NULL DEFAULT 0 COMMENT '关注数',
    `like_count`       int unsigned    NOT NULL DEFAULT 0 COMMENT '收到的点赞数',
    `created_at`       datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`       datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_user_id` (`user_id`) USING BTREE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci COMMENT ='用户统计表';

CREATE TABLE IF NOT EXISTS `parties`
(
    `id`            bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '派对ID',
    `user_id`       int unsigned    NOT NULL COMMENT '创建者用户ID',
    `title`         varchar(255)    NOT NULL COMMENT '标题',
    `type`          varchar(50)     NOT NULL COMMENT '类型: 派对/场地',
    `description`   text            COMMENT '描述',
    `cover_image`   text            COMMENT '封面图',
    `location_name` varchar(255)    NOT NULL DEFAULT '' COMMENT '位置名称',
    `address`       varchar(500)    NOT NULL DEFAULT '' COMMENT '详细地址',
    `latitude`      decimal(10,7)   NOT NULL COMMENT '纬度',
    `longitude`     decimal(10,7)   NOT NULL COMMENT '经度',
    `status`        varchar(20)     NOT NULL DEFAULT 'active' COMMENT '状态: active/offline',
    `images_json`   json            COMMENT '图片列表',
    `category`      int             NOT NULL DEFAULT 0 COMMENT '分类',
    `district_id`   int             NOT NULL DEFAULT 0 COMMENT '行政区ID',
    `area_id`       int             NOT NULL DEFAULT 0 COMMENT '商圈/区域ID',
    `tags`          int             NOT NULL DEFAULT 0 COMMENT '标签位图',
    `created_at`    datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`    datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_user_id` (`user_id`) USING BTREE,
    KEY `idx_type` (`type`) USING BTREE,
    KEY `idx_status` (`status`) USING BTREE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci COMMENT ='派对/场地表';

-- 如已有 parties 表，执行以下 ALTER 添加 status 字段:
-- ALTER TABLE `parties` ADD COLUMN `status` varchar(20) NOT NULL DEFAULT 'active' COMMENT '状态: active/offline' AFTER `longitude`;
-- ALTER TABLE `parties` ADD COLUMN `district_id` int NOT NULL DEFAULT 0 COMMENT '行政区ID' AFTER `category`;
-- ALTER TABLE `parties` ADD COLUMN `area_id` int NOT NULL DEFAULT 0 COMMENT '商圈/区域ID' AFTER `district_id`;
-- ALTER TABLE `parties` ADD COLUMN `tags` int NOT NULL DEFAULT 0 COMMENT '标签位图' AFTER `area_id`;

CREATE TABLE IF NOT EXISTS `events`
(
    `id`                bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '活动ID',
    `user_id`           int unsigned    NOT NULL COMMENT '创建者用户ID',
    `title`             varchar(80)     NOT NULL COMMENT '活动名称',
    `share_title`       varchar(20)     NOT NULL COMMENT '分享标题',
    `start_at`          datetime        NOT NULL COMMENT '开始时间',
    `end_at`            datetime        NOT NULL COMMENT '结束时间',
    `summary`           text            COMMENT '活动概要（纯文本）',
    `summary_html`      text            COMMENT '活动概要（HTML）',
    `strong_real_name`  tinyint(1)      NOT NULL DEFAULT 0 COMMENT '是否强实名',
    `minor_protection`  tinyint(1)      NOT NULL DEFAULT 0 COMMENT '是否未成年人校验',
    `address_mode`      varchar(20)     NOT NULL DEFAULT 'default' COMMENT '地址模式',
    `province`          varchar(50)     NOT NULL DEFAULT '' COMMENT '省份',
    `city`              varchar(50)     NOT NULL DEFAULT '' COMMENT '城市',
    `district`          varchar(50)     NOT NULL DEFAULT '' COMMENT '地区',
    `location`          varchar(255)    NOT NULL DEFAULT '' COMMENT '详细地址',
    `detail_poster`     text            COMMENT '详情页海报',
    `detail_long_poster` text           COMMENT '详情长图海报',
    `share_poster`      text            COMMENT '列表及分享海报',
    `group_poster`      text            COMMENT '社群海报',
    `promoter_confirmed` tinyint(1)     NOT NULL DEFAULT 0 COMMENT '主办方确认',
    `safety_agreement`  tinyint(1)      NOT NULL DEFAULT 0 COMMENT '安全保障确认',
    `ticket_statement`  tinyint(1)      NOT NULL DEFAULT 0 COMMENT '票券声明确认',
    `notes`             text            COMMENT '补充说明',
    `status`            varchar(20)     NOT NULL DEFAULT 'draft' COMMENT '状态: draft/已上架/已下架',
    `created_at`        datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`        datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_event_user_id` (`user_id`) USING BTREE,
    KEY `idx_event_status` (`status`) USING BTREE,
    KEY `idx_event_start_at` (`start_at`) USING BTREE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci COMMENT ='活动表';

CREATE TABLE IF NOT EXISTS `event_tickets`
(
    `id`              bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '票券ID',
    `event_id`        bigint unsigned NOT NULL COMMENT '活动ID',
    `name`            varchar(50)     NOT NULL COMMENT '票券名称',
    `price`           bigint          NOT NULL COMMENT '价格（分）',
    `stock`           int             NOT NULL COMMENT '库存',
    `purchase_limit`  int             NOT NULL DEFAULT 1 COMMENT '限购数量',
    `start_at`        datetime        NOT NULL COMMENT '开售时间',
    `end_at`          datetime        NOT NULL COMMENT '停售时间',
    `description`     text            COMMENT '描述',
    `created_at`      datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_ticket_event_id` (`event_id`) USING BTREE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci COMMENT ='活动票券表';

CREATE TABLE IF NOT EXISTS `admin`
(
    `id`         int unsigned NOT NULL AUTO_INCREMENT COMMENT '管理员ID',
    `username`   varchar(50)  NOT NULL COMMENT '用户名',
	`nickname`   varchar(50)  NOT NULL DEFAULT '' COMMENT '展示昵称',
    `password`   varchar(255) NOT NULL COMMENT '密码(bcrypt)',
    `avatar`     varchar(255) NOT NULL DEFAULT '' COMMENT '头像',
    `gender`     tinyint      NOT NULL DEFAULT 3 COMMENT '性别 1:男 2:女 3:未知',
    `mobile`     varchar(11)  NOT NULL DEFAULT '' COMMENT '手机号',
    `email`      varchar(50)  NOT NULL DEFAULT '' COMMENT '邮箱',
    `motto`      varchar(500) NOT NULL DEFAULT '' COMMENT '座右铭',
    `role_id`    bigint unsigned NOT NULL DEFAULT 0 COMMENT '后台角色ID',
    `status`     tinyint      NOT NULL DEFAULT 1 COMMENT '状态 1:正常 2:停用',
    `created_at` datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_username` (`username`) USING BTREE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci COMMENT ='管理员表';

-- Existing admin table migration:
-- ALTER TABLE `admin` ADD COLUMN `role_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '后台角色ID' AFTER `motto`;
-- ALTER TABLE `admin` ADD KEY `idx_admin_role` (`role_id`);
-- ALTER TABLE `admin` ADD COLUMN `nickname` varchar(50) NOT NULL DEFAULT '' COMMENT '展示昵称' AFTER `username`;

CREATE TABLE IF NOT EXISTS `admin_wechat_subscribers`
(
    `id`         bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
    `admin_id`   bigint unsigned NOT NULL COMMENT '管理员ID',
    `open_id`    varchar(64)     NOT NULL COMMENT '管理员微信openid',
    `enabled`    tinyint         NOT NULL DEFAULT 1 COMMENT '1启用 0禁用',
    `created_at` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_admin_openid` (`admin_id`, `open_id`) USING BTREE,
    KEY `idx_enabled` (`enabled`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='管理员微信订阅通知表';

CREATE TABLE IF NOT EXISTS `organizers`
(
    `id`                bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主办方ID',
    `user_id`           bigint unsigned NOT NULL COMMENT '用户ID',
    `type`              varchar(20)     NOT NULL DEFAULT 'merchant' COMMENT '兼容字段: 入驻不再区分场地/派对，场地/派对由 activities.type 决定',
    `name`              varchar(100)    NOT NULL COMMENT '主办方名称',
    `logo`              varchar(255)    NOT NULL DEFAULT '' COMMENT 'Logo',
    `status`            tinyint         NOT NULL DEFAULT 0 COMMENT '0待审核 1审核中 2已认证 3未通过',
    `enabled`           tinyint         NOT NULL DEFAULT 1 COMMENT '1启用 0停用，与审核状态分离',
    `reject_reason`     varchar(500)    NOT NULL DEFAULT '' COMMENT '拒绝原因',
    `level`             varchar(10)     NOT NULL DEFAULT 'LV1' COMMENT '等级',
    `service_fee_rate`  decimal(5,2)    NOT NULL DEFAULT 0.00 COMMENT '服务费比例',
    `province`          varchar(50)     NOT NULL DEFAULT '',
    `city`              varchar(50)     NOT NULL DEFAULT '',
    `district`          varchar(50)     NOT NULL DEFAULT '',
    `bank_account_name` varchar(50)     NOT NULL DEFAULT '' COMMENT '收款人',
    `bank_account_no`   varchar(50)     NOT NULL DEFAULT '' COMMENT '收款账户',
    `bank_name`         varchar(50)     NOT NULL DEFAULT '' COMMENT '银行名称',
	`bank_contact_name` varchar(50)     NOT NULL DEFAULT '' COMMENT '收款联系人',
	`bank_contact_phone` varchar(20)    NOT NULL DEFAULT '' COMMENT '收款联系电话',
    `created_at`        datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`        datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_organizer_user` (`user_id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='主办方表';

-- 如已有 organizers 表，type 字段保留为兼容字段；场地/派对区分请使用 activities.type。
-- ALTER TABLE `organizers` MODIFY COLUMN `type` varchar(20) NOT NULL DEFAULT 'merchant' COMMENT '兼容字段: 入驻不再区分场地/派对，场地/派对由 activities.type 决定';
-- ALTER TABLE `organizers` ADD COLUMN `enabled` tinyint NOT NULL DEFAULT 1 COMMENT '1启用 0停用' AFTER `status`;
-- ALTER TABLE `organizers` ADD COLUMN `bank_contact_name` varchar(50) NOT NULL DEFAULT '' COMMENT '收款联系人' AFTER `bank_name`;
-- ALTER TABLE `organizers` ADD COLUMN `bank_contact_phone` varchar(20) NOT NULL DEFAULT '' COMMENT '收款联系电话' AFTER `bank_contact_name`;

CREATE TABLE IF NOT EXISTS `activities`
(
    `id`                bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '活动ID',
    `organizer_id`      bigint unsigned NOT NULL COMMENT '主办方ID',
    `type`              varchar(20)     NOT NULL DEFAULT 'party' COMMENT '活动类型: party派对 venue场地',
    `name`              varchar(80)     NOT NULL COMMENT '活动名称',
    `share_title`       varchar(20)     NOT NULL DEFAULT '' COMMENT '分享标题',
    `start_time`        datetime        NULL COMMENT '开始时间',
    `end_time`          datetime        NULL COMMENT '结束时间',
    `real_name_mode`    tinyint         NOT NULL DEFAULT 0 COMMENT '实名模式',
    `minor_check`       tinyint         NOT NULL DEFAULT 0 COMMENT '未成年人校验',
    `description`       text            COMMENT '活动概要',
    `province`          varchar(50)     NOT NULL DEFAULT '',
    `city`              varchar(50)     NOT NULL DEFAULT '',
    `district`          varchar(50)     NOT NULL DEFAULT '',
    `address`           varchar(200)    NOT NULL DEFAULT '',
    `latitude`          decimal(10,6)   NOT NULL DEFAULT 0,
    `longitude`         decimal(10,6)   NOT NULL DEFAULT 0,
    `poster_detail`     varchar(255)    NOT NULL DEFAULT '',
    `poster_long`       varchar(255)    NOT NULL DEFAULT '',
    `poster_list`       varchar(255)    NOT NULL DEFAULT '',
    `poster_wechat`     varchar(255)    NOT NULL DEFAULT '',
    `qualification_doc` varchar(255)    NOT NULL DEFAULT '',
    `status`            tinyint         NOT NULL DEFAULT 0 COMMENT '0草稿 1待审核 2审核中 3已上架 4未通过',
    `reject_reason`     varchar(500)    NOT NULL DEFAULT '',
    `created_at`        datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`        datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_activity_organizer` (`organizer_id`) USING BTREE,
    KEY `idx_activity_type` (`type`) USING BTREE,
    KEY `idx_activity_status` (`status`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='活动表v2';

-- 如已有 activities 表，执行以下 ALTER 添加活动类型字段:
-- ALTER TABLE `activities` ADD COLUMN `type` varchar(20) NOT NULL DEFAULT 'party' COMMENT '活动类型: party派对 venue场地' AFTER `organizer_id`;
-- ALTER TABLE `activities` ADD INDEX `idx_activity_type` (`type`);

CREATE TABLE IF NOT EXISTS `ticket_specs`
(
    `id`             bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '票券规格ID',
    `activity_id`    bigint unsigned NOT NULL COMMENT '活动ID',
    `name`           varchar(30)     NOT NULL COMMENT '规格名称',
	`description`    varchar(500)    NOT NULL DEFAULT '' COMMENT '票种说明',
    `is_enabled`     tinyint         NOT NULL DEFAULT 1 COMMENT '启用状态',
    `sale_start`     datetime        NULL COMMENT '开售时间',
    `sale_end`       datetime        NULL COMMENT '停售时间',
    `price`          bigint          NOT NULL COMMENT '价格(分)',
    `stock`          int             NOT NULL DEFAULT 0 COMMENT '库存',
    `sold_count`     int             NOT NULL DEFAULT 0 COMMENT '已售数量',
    `purchase_limit` int             NOT NULL DEFAULT 1 COMMENT '限购数量',
    `max_attendees`  int             NOT NULL DEFAULT 1 COMMENT '观演人数',
    `created_at`     datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_ticket_spec_activity` (`activity_id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='票券规格表';

-- ALTER TABLE `ticket_specs` ADD COLUMN `description` varchar(500) NOT NULL DEFAULT '' COMMENT '票种说明' AFTER `name`;

CREATE TABLE IF NOT EXISTS `activity_daily_stats`
(
    `id`            bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
    `activity_id`   bigint unsigned NOT NULL COMMENT '活动ID',
    `stat_date`     date            NOT NULL COMMENT '统计日期',
    `view_count`    bigint          NOT NULL DEFAULT 0 COMMENT '浏览量PV',
    `visitor_count` bigint          NOT NULL DEFAULT 0 COMMENT '独立访客数UV',
    `created_at`    datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`    datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_activity_daily_stat` (`activity_id`, `stat_date`) USING BTREE,
    KEY `idx_activity_daily_stat_date` (`stat_date`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='活动每日浏览统计';

CREATE TABLE IF NOT EXISTS `activity_daily_visitors`
(
    `id`          bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
    `activity_id` bigint unsigned NOT NULL COMMENT '活动ID',
    `stat_date`   date            NOT NULL COMMENT '统计日期',
    `visitor_key` varchar(80)     NOT NULL COMMENT '用户ID或匿名访客哈希',
    `user_id`     bigint unsigned NOT NULL DEFAULT 0 COMMENT '登录用户ID，匿名为0',
    `created_at`  datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_activity_daily_visitor` (`activity_id`, `stat_date`, `visitor_key`) USING BTREE,
    KEY `idx_activity_daily_visitor_date` (`stat_date`) USING BTREE,
    KEY `idx_activity_daily_visitor_user` (`user_id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='活动每日独立访客';

CREATE TABLE IF NOT EXISTS `activity_subscriptions`
(
    `id`          bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '订阅ID',
    `activity_id` bigint unsigned NOT NULL COMMENT '活动ID',
    `user_id`     bigint unsigned NOT NULL COMMENT '用户ID',
    `created_at`  datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_activity_user` (`activity_id`, `user_id`) USING BTREE,
    KEY `idx_activity_subscription_activity` (`activity_id`) USING BTREE,
    KEY `idx_activity_subscription_user` (`user_id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='活动订阅表';

CREATE TABLE IF NOT EXISTS `venue_subscriptions`
(
    `id`           bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '场地订阅ID',
    `organizer_id` bigint unsigned NOT NULL COMMENT '场地主办方ID',
    `user_id`      bigint unsigned NOT NULL COMMENT '用户ID',
    `created_at`   datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_venue_user` (`organizer_id`, `user_id`) USING BTREE,
    KEY `idx_venue_subscription_organizer` (`organizer_id`) USING BTREE,
    KEY `idx_venue_subscription_user` (`user_id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='场地订阅表';

CREATE TABLE IF NOT EXISTS `ticket_orders`
(
    `id`              bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '订单ID',
    `order_no`        varchar(30)     NOT NULL COMMENT '订单号',
    `user_id`         bigint unsigned NOT NULL COMMENT '用户ID',
    `activity_id`     bigint unsigned NOT NULL COMMENT '活动ID',
    `ticket_spec_id`  bigint unsigned NOT NULL COMMENT '票券规格ID',
    `quantity`        int             NOT NULL COMMENT '数量',
    `total_price`     bigint          NOT NULL COMMENT '总价(分)',
    `actual_price`    bigint          NOT NULL COMMENT '实付(分)',
    `points_amount`   bigint          NOT NULL DEFAULT 0 COMMENT '抵扣积分数量',
    `points_discount` bigint          NOT NULL DEFAULT 0 COMMENT '积分抵扣金额(分)',
    `pay_method`      varchar(20)     NOT NULL DEFAULT '',
    `sales_channel`   varchar(20)     NOT NULL DEFAULT 'wechat' COMMENT '销售渠道：wechat/douyin/web/other',
    `pay_time`        datetime        NULL,
    `buyer_name`      varchar(50)     NOT NULL DEFAULT '',
    `buyer_id_card`   varchar(20)     NOT NULL DEFAULT '',
    `status`          tinyint         NOT NULL DEFAULT 0 COMMENT '0待支付 1待使用 2已使用 3取消 4退款中 5退款成功 6退款拒绝',
    `expire_time`     datetime        NULL,
    `qr_code`         varchar(255)    NOT NULL DEFAULT '',
    `qr_code_url`     varchar(255)    NOT NULL DEFAULT '' COMMENT '取票二维码图片URL',
    `cancel_reason`   varchar(100)    NOT NULL DEFAULT '',
    `user_deleted_at` datetime        NULL COMMENT '用户侧删除/隐藏时间',
    `created_at`      datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`      datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_ticket_order_no` (`order_no`) USING BTREE,
    KEY `idx_ticket_order_user` (`user_id`) USING BTREE,
    KEY `idx_ticket_order_activity` (`activity_id`) USING BTREE,
    KEY `idx_ticket_order_sales_channel` (`sales_channel`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='票务订单表';

-- Existing deployments:
-- ALTER TABLE `ticket_orders` ADD COLUMN `points_amount` bigint NOT NULL DEFAULT 0 COMMENT '抵扣积分数量' AFTER `actual_price`;
-- ALTER TABLE `ticket_orders` ADD COLUMN `points_discount` bigint NOT NULL DEFAULT 0 COMMENT '积分抵扣金额(分)' AFTER `points_amount`;
-- ALTER TABLE `ticket_orders` ADD COLUMN `sales_channel` varchar(20) NOT NULL DEFAULT 'wechat' COMMENT '销售渠道：wechat/douyin/web/other' AFTER `pay_method`;
-- ALTER TABLE `ticket_orders` ADD KEY `idx_ticket_order_sales_channel` (`sales_channel`);
-- ALTER TABLE `ticket_orders` ADD COLUMN `qr_code_url` varchar(255) NOT NULL DEFAULT '' COMMENT '取票二维码图片URL' AFTER `qr_code`;
-- ALTER TABLE `ticket_orders` ADD COLUMN `user_deleted_at` datetime NULL COMMENT '用户侧删除/隐藏时间' AFTER `cancel_reason`;

CREATE TABLE IF NOT EXISTS `ticket_order_viewers`
(
    `id`         bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '订单观演人ID',
    `order_id`   bigint unsigned NOT NULL COMMENT '票务订单ID',
    `order_no`   varchar(30)     NOT NULL COMMENT '订单号',
    `viewer_id`  bigint unsigned NOT NULL DEFAULT 0 COMMENT '观演人ID，直接提交实名信息时为0',
    `real_name`  varchar(50)     NOT NULL COMMENT '观演人姓名快照',
    `id_card`    varchar(20)     NOT NULL COMMENT '身份证号快照',
    `phone`      varchar(20)     NOT NULL DEFAULT '' COMMENT '手机号快照',
    `type`       tinyint         NOT NULL DEFAULT 1 COMMENT '类型：1成人 2儿童 3老人',
    `created_at` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_ticket_order_viewer_order` (`order_id`) USING BTREE,
    KEY `idx_ticket_order_viewer_order_no` (`order_no`) USING BTREE,
    KEY `idx_ticket_order_viewer_viewer` (`viewer_id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='票务订单观演人表';

-- Existing deployments:
-- CREATE TABLE IF NOT EXISTS `ticket_order_viewers` (
--     `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '订单观演人ID',
--     `order_id` bigint unsigned NOT NULL COMMENT '票务订单ID',
--     `order_no` varchar(30) NOT NULL COMMENT '订单号',
--     `viewer_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '观演人ID，直接提交实名信息时为0',
--     `real_name` varchar(50) NOT NULL COMMENT '观演人姓名快照',
--     `id_card` varchar(20) NOT NULL COMMENT '身份证号快照',
--     `phone` varchar(20) NOT NULL DEFAULT '' COMMENT '手机号快照',
--     `type` tinyint NOT NULL DEFAULT 1 COMMENT '类型：1成人 2儿童 3老人',
--     `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
--     `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
--     PRIMARY KEY (`id`) USING BTREE,
--     KEY `idx_ticket_order_viewer_order` (`order_id`) USING BTREE,
--     KEY `idx_ticket_order_viewer_order_no` (`order_no`) USING BTREE,
--     KEY `idx_ticket_order_viewer_viewer` (`viewer_id`) USING BTREE
-- ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='票务订单观演人表';

CREATE TABLE IF NOT EXISTS `refunds`
(
    `id`                 bigint unsigned NOT NULL AUTO_INCREMENT,
    `order_id`           bigint unsigned NOT NULL,
    `refund_no`          varchar(30)     NOT NULL,
    `refund_amount`      bigint          NOT NULL COMMENT '退款金额(分)',
    `deduct_amount`      bigint          NOT NULL DEFAULT 0 COMMENT '扣除金额(分)',
    `reason`             varchar(200)    NOT NULL DEFAULT '',
    `status`             tinyint         NOT NULL DEFAULT 0 COMMENT '0审核中 1退款中 2成功 3拒绝',
    `wechat_refund_id`   varchar(64)     NOT NULL DEFAULT '' COMMENT '微信支付退款号',
    `wechat_status`      varchar(32)     NOT NULL DEFAULT '' COMMENT '微信退款状态',
    `reject_reason`      varchar(500)    NOT NULL DEFAULT '',
    `expect_arrive_date` date            NULL,
    `created_at`         datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`         datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_refund_no` (`refund_no`) USING BTREE,
    KEY `idx_refund_order` (`order_id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='退款表';

CREATE TABLE IF NOT EXISTS `refund_logs`
(
    `id`          bigint unsigned NOT NULL AUTO_INCREMENT,
    `refund_id`   bigint unsigned NOT NULL,
    `status`      varchar(50)     NOT NULL,
    `description` varchar(200)    NOT NULL,
    `created_at`  datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_refund_log_refund` (`refund_id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='退款进度表';

CREATE TABLE IF NOT EXISTS `verifiers`
(
    `id`               bigint unsigned NOT NULL AUTO_INCREMENT,
    `organizer_id`     bigint unsigned NOT NULL,
    `user_id`          bigint unsigned NOT NULL DEFAULT 0 COMMENT '绑定的小程序用户ID',
    `name`             varchar(50)     NOT NULL,
    `phone`            varchar(20)     NOT NULL,
    `status`           tinyint         NOT NULL DEFAULT 0 COMMENT '0未激活 1已激活',
    `permission_scope` varchar(20)     NOT NULL DEFAULT '活动',
    `channel`          varchar(20)     NOT NULL DEFAULT '',
    `bound_at`         datetime        NULL COMMENT '绑定时间',
    `created_at`       datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`       datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_verifier_organizer` (`organizer_id`) USING BTREE,
    KEY `idx_verifier_user` (`user_id`) USING BTREE,
    KEY `idx_verifier_phone` (`phone`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='核销员表';

-- Existing deployments:
-- ALTER TABLE `verifiers` ADD COLUMN `user_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '绑定的小程序用户ID' AFTER `organizer_id`;
-- ALTER TABLE `verifiers` ADD COLUMN `bound_at` datetime NULL COMMENT '绑定时间' AFTER `channel`;
-- ALTER TABLE `verifiers` ADD KEY `idx_verifier_user` (`user_id`) USING BTREE;

CREATE TABLE IF NOT EXISTS `verification_records`
(
    `id`          bigint unsigned NOT NULL AUTO_INCREMENT,
    `order_id`    bigint unsigned NOT NULL,
    `verifier_id` bigint unsigned NOT NULL,
    `activity_id` bigint unsigned NOT NULL,
    `verified_at` datetime        NOT NULL,
    `created_at`  datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_verify_order` (`order_id`) USING BTREE,
    KEY `idx_verify_verifier` (`verifier_id`) USING BTREE,
    KEY `idx_verify_activity` (`activity_id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='核销记录表';

CREATE TABLE IF NOT EXISTS `cancel_reasons`
(
    `id`     bigint unsigned NOT NULL AUTO_INCREMENT,
    `reason` varchar(100)    NOT NULL,
    `sort`   int             NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='取消原因表';

CREATE TABLE IF NOT EXISTS `refund_reasons`
(
    `id`     bigint unsigned NOT NULL AUTO_INCREMENT,
    `reason` varchar(100)    NOT NULL,
    `sort`   int             NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='退款原因表';

CREATE TABLE IF NOT EXISTS `organizer_stores`
(
    `id`           bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '门店ID',
    `organizer_id` bigint unsigned NOT NULL COMMENT '主办方ID',
    `name`         varchar(100)    NOT NULL COMMENT '门店名称',
    `logo`         varchar(255)    NOT NULL DEFAULT '',
    `address`      varchar(200)    NOT NULL DEFAULT '',
    `latitude`     decimal(10,6)   NOT NULL DEFAULT 0,
    `longitude`    decimal(10,6)   NOT NULL DEFAULT 0,
    `phone`        varchar(20)     NOT NULL DEFAULT '',
    `created_at`   datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`   datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_store_organizer` (`organizer_id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='主办方门店表';

CREATE TABLE IF NOT EXISTS `platform_banners`
(
    `id`         bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'Banner ID',
    `title`      varchar(100)    NOT NULL DEFAULT '',
    `image`      varchar(255)    NOT NULL COMMENT '图片URL',
    `link`       varchar(255)    NOT NULL DEFAULT '',
    `position`   varchar(50)     NOT NULL DEFAULT 'home',
    `sort`       int             NOT NULL DEFAULT 0,
    `status`     tinyint         NOT NULL DEFAULT 1 COMMENT '1启用 0禁用',
    `created_at` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_banner_position` (`position`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='平台Banner表';

CREATE TABLE IF NOT EXISTS `platform_settings`
(
    `id`            bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '设置ID',
    `setting_key`   varchar(100)    NOT NULL COMMENT '设置键',
    `setting_value` text            COMMENT '设置值',
    `remark`        varchar(255)    NOT NULL DEFAULT '',
    `created_at`    datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`    datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_platform_setting_key` (`setting_key`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='平台设置表';

INSERT IGNORE INTO `platform_settings` (`setting_key`, `setting_value`, `remark`) VALUES
('points_discount_cents_per_point', '10', '每积分可抵扣金额，单位分'),
('points_reward_cents_per_point', '1000', '消费多少分奖励1积分'),
('withdraw_arrival_cycle', 'T+1 到 T+3 个工作日', '商家提现到账周期展示文案'),
('system_name', 'Hyper', '平台系统名称'),
('icp_record_no', '', 'ICP备案号'),
('customer_service_phone', '', '客服电话'),
('customer_service_wechat', '', '客服微信'),
('customer_service_email', '', '客服邮箱'),
('customer_service_hours', '', '客服服务时间');

CREATE TABLE IF NOT EXISTS `admin_roles`
(
    `id`          bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '角色ID',
    `name`        varchar(50)     NOT NULL COMMENT '角色名称',
    `description` varchar(255)    NOT NULL DEFAULT '' COMMENT '角色说明',
    `permissions` text            NULL COMMENT '权限JSON或逗号列表',
    `status`      tinyint         NOT NULL DEFAULT 1 COMMENT '1启用 0禁用',
    `created_at`  datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='后台角色表';

-- RBAC production migration (run before deploying global RBAC enforcement):
-- 1. Create the enabled super-administrator role once.
-- INSERT INTO `admin_roles` (`name`, `description`, `permissions`, `status`, `created_at`, `updated_at`)
-- SELECT '超级管理员', '系统初始化超级管理员', '["*"]', 1, NOW(), NOW()
-- WHERE NOT EXISTS (SELECT 1 FROM `admin_roles` WHERE `permissions` = '["*"]' AND `status` = 1);
-- 2. Assign the existing production admin account to that role. Replace `admin` when the actual bootstrap username differs.
-- SET @super_role_id := (SELECT `id` FROM `admin_roles` WHERE `permissions` = '["*"]' AND `status` = 1 ORDER BY `id` ASC LIMIT 1);
-- UPDATE `admin` SET `role_id` = @super_role_id, `status` = 1 WHERE `username` = 'admin';
-- 3. Assign all remaining enabled administrators an explicit least-privilege role, or disable them before deployment.
-- UPDATE `admin` SET `status` = 2 WHERE `id` <> (SELECT `id` FROM (SELECT `id` FROM `admin` WHERE `username` = 'admin' LIMIT 1) AS bootstrap_admin) AND `role_id` = 0;

CREATE TABLE IF NOT EXISTS `admin_operation_logs`
(
    `id`         bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '日志ID',
    `admin_id`   bigint unsigned NOT NULL COMMENT '管理员ID',
    `action`     varchar(100)    NOT NULL COMMENT '操作动作',
    `resource`   varchar(100)    NOT NULL DEFAULT '' COMMENT '资源',
    `resource_type` varchar(50)  NOT NULL DEFAULT '' COMMENT '资源类型',
    `resource_id` varchar(100)   NOT NULL DEFAULT '' COMMENT '资源ID或业务单号',
    `resource_name` varchar(255) NOT NULL DEFAULT '' COMMENT '资源展示名称快照',
    `method`     varchar(20)     NOT NULL DEFAULT '' COMMENT 'HTTP方法',
    `path`       varchar(255)    NOT NULL DEFAULT '' COMMENT '请求路径',
    `result`     varchar(20)     NOT NULL DEFAULT 'success' COMMENT 'success/denied/failed',
    `error_code` varchar(100)    NOT NULL DEFAULT '' COMMENT '失败错误码',
    `error_message` varchar(255) NOT NULL DEFAULT '' COMMENT '失败错误信息',
    `ip`         varchar(50)     NOT NULL DEFAULT '' COMMENT 'IP',
    `remark`     varchar(255)    NOT NULL DEFAULT '' COMMENT '备注',
    `created_at` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_admin_log_admin` (`admin_id`) USING BTREE,
    KEY `idx_admin_log_action` (`action`) USING BTREE,
    KEY `idx_admin_log_resource_type` (`resource_type`) USING BTREE,
    KEY `idx_admin_log_result` (`result`) USING BTREE,
    KEY `idx_admin_log_created` (`created_at`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='后台操作日志表';

-- Existing admin_operation_logs table migration:
-- ALTER TABLE `admin_operation_logs` ADD COLUMN `resource_type` varchar(50) NOT NULL DEFAULT '' COMMENT '资源类型' AFTER `resource`;
-- ALTER TABLE `admin_operation_logs` ADD COLUMN `resource_id` varchar(100) NOT NULL DEFAULT '' COMMENT '资源ID或业务单号' AFTER `resource_type`;
-- ALTER TABLE `admin_operation_logs` ADD COLUMN `resource_name` varchar(255) NOT NULL DEFAULT '' COMMENT '资源展示名称快照' AFTER `resource_id`;
-- ALTER TABLE `admin_operation_logs` ADD COLUMN `result` varchar(20) NOT NULL DEFAULT 'success' COMMENT 'success/denied/failed' AFTER `path`;
-- ALTER TABLE `admin_operation_logs` ADD COLUMN `error_code` varchar(100) NOT NULL DEFAULT '' COMMENT '失败错误码' AFTER `result`;
-- ALTER TABLE `admin_operation_logs` ADD COLUMN `error_message` varchar(255) NOT NULL DEFAULT '' COMMENT '失败错误信息' AFTER `error_code`;
-- ALTER TABLE `admin_operation_logs` ADD KEY `idx_admin_log_action` (`action`);
-- ALTER TABLE `admin_operation_logs` ADD KEY `idx_admin_log_resource_type` (`resource_type`);
-- ALTER TABLE `admin_operation_logs` ADD KEY `idx_admin_log_result` (`result`);

CREATE TABLE IF NOT EXISTS `admin_categories`
(
    `id`         bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '分类ID',
    `type`       varchar(50)     NOT NULL COMMENT '分类类型',
    `name`       varchar(100)    NOT NULL COMMENT '分类名称',
    `image`      varchar(255)    NOT NULL DEFAULT '' COMMENT '图片',
    `value`      varchar(100)    NOT NULL DEFAULT '' COMMENT '扩展值',
    `sort`       int             NOT NULL DEFAULT 0 COMMENT '排序',
    `status`     tinyint         NOT NULL DEFAULT 1 COMMENT '1启用 0禁用',
    `created_at` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_admin_category_type` (`type`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='后台分类配置表';

-- type=coupon_tag 的 admin_categories 为平台配置的优惠标签定义。
CREATE TABLE IF NOT EXISTS `content_tag_relations`
(
    `id`          bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '关联ID',
    `tag_id`      bigint unsigned NOT NULL COMMENT 'admin_categories.ID，且类型为coupon_tag',
    `target_type` varchar(20)     NOT NULL COMMENT 'activity/venue/party',
    `target_id`   bigint unsigned NOT NULL COMMENT '活动ID、主办方场地ID或旧派对ID',
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_content_tag_target` (`tag_id`, `target_type`, `target_id`) USING BTREE,
    KEY `idx_content_tag_target` (`target_type`, `target_id`) USING BTREE,
    KEY `idx_content_tag_tag` (`tag_id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='内容优惠标签关联表';

CREATE TABLE IF NOT EXISTS `activity_collections`
(
    `id`           bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '活动合集ID',
    `organizer_id` bigint unsigned NOT NULL COMMENT '主办方ID',
    `title`        varchar(100)    NOT NULL COMMENT '合集标题',
    `share_title`  varchar(100)    NOT NULL DEFAULT '' COMMENT '分享标题',
    `description`  text            NULL COMMENT '合集简介',
    `share_image`  varchar(255)    NOT NULL DEFAULT '' COMMENT '分享图',
    `status`       tinyint         NOT NULL DEFAULT 1 COMMENT '1启用 0禁用',
    `created_at`   datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`   datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_activity_collection_organizer` (`organizer_id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='活动合集表';

CREATE TABLE IF NOT EXISTS `activity_collection_items`
(
    `id`            bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '合集活动ID',
    `collection_id` bigint unsigned NOT NULL COMMENT '合集ID',
    `activity_id`   bigint unsigned NOT NULL COMMENT '活动ID',
    `sort`          int             NOT NULL DEFAULT 0 COMMENT '排序',
    `created_at`    datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_collection_item_collection` (`collection_id`) USING BTREE,
    KEY `idx_collection_item_activity` (`activity_id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='活动合集明细表';

CREATE TABLE IF NOT EXISTS `platform_messages`
(
    `id`         bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '消息ID',
    `title`      varchar(100)    NOT NULL COMMENT '标题',
    `content`    text            NULL COMMENT '内容',
    `content_type` varchar(20)   NOT NULL DEFAULT 'text' COMMENT '内容类型：text/rich_text',
    `cover_image` varchar(255)   NOT NULL DEFAULT '' COMMENT '消息封面图',
    `media_data` text            NULL COMMENT '消息图集JSON数组',
    `type`       varchar(50)     NOT NULL DEFAULT '' COMMENT '消息类型',
    `target`     varchar(50)     NOT NULL DEFAULT '' COMMENT '发送对象',
    `channel`    varchar(30)     NOT NULL DEFAULT 'in_app' COMMENT '发送渠道，当前仅 in_app',
    `creator_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '创建管理员ID',
    `status`     tinyint         NOT NULL DEFAULT 1 COMMENT '1已发布 0草稿',
    `created_at` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_platform_message_creator` (`creator_id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='平台消息表';

-- Existing platform_messages table migration:
-- ALTER TABLE `platform_messages` ADD COLUMN `channel` varchar(30) NOT NULL DEFAULT 'in_app' COMMENT '发送渠道，当前仅 in_app' AFTER `target`;
-- ALTER TABLE `platform_messages` ADD COLUMN `creator_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '创建管理员ID' AFTER `channel`;
-- ALTER TABLE `platform_messages` ADD KEY `idx_platform_message_creator` (`creator_id`);
-- ALTER TABLE `platform_messages` ADD COLUMN `content_type` varchar(20) NOT NULL DEFAULT 'text' COMMENT '内容类型：text/rich_text' AFTER `content`;
-- ALTER TABLE `platform_messages` ADD COLUMN `cover_image` varchar(255) NOT NULL DEFAULT '' COMMENT '消息封面图' AFTER `content_type`;
-- ALTER TABLE `platform_messages` ADD COLUMN `media_data` text NULL COMMENT '消息图集JSON数组' AFTER `cover_image`;

CREATE TABLE IF NOT EXISTS `organizer_message_reads`
(
    `id`           bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '商家消息阅读ID',
    `organizer_id` bigint unsigned NOT NULL COMMENT '主办方ID',
    `message_id`   bigint unsigned NOT NULL COMMENT '消息ID',
    `is_read`      tinyint         NOT NULL DEFAULT 0 COMMENT '1已读 0未读',
    `read_at`      datetime        NULL COMMENT '阅读时间',
    `created_at`   datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`   datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_org_msg` (`organizer_id`, `message_id`) USING BTREE,
    KEY `idx_org_msg_message` (`message_id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='商家消息阅读表';

CREATE TABLE IF NOT EXISTS `platform_message_deliveries`
(
    `id`         bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '投递记录ID',
    `message_id` bigint unsigned NOT NULL COMMENT '平台消息ID',
    `user_id`    bigint unsigned NOT NULL COMMENT '目标用户ID',
    `status`     tinyint         NOT NULL DEFAULT 0 COMMENT '0待发送 1已发送 2失败 3已读',
    `sent_at`    datetime        NULL COMMENT '发送时间',
    `read_at`    datetime        NULL COMMENT '阅读时间',
    `error`      varchar(255)    NOT NULL DEFAULT '' COMMENT '失败原因',
    `created_at` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_message_user` (`message_id`, `user_id`) USING BTREE,
    KEY `idx_delivery_user` (`user_id`) USING BTREE,
    KEY `idx_delivery_status` (`status`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='平台消息投递与阅读记录';

CREATE TABLE IF NOT EXISTS `organizer_profiles`
(
    `id`             bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '商家资料ID',
    `organizer_id`   bigint unsigned NOT NULL COMMENT '主办方ID',
    `cover_image`    varchar(255)    NOT NULL DEFAULT '' COMMENT '封面图',
    `gallery`        text            NULL COMMENT '组图JSON数组',
    `description`    text            NULL COMMENT '简介',
    `business_hours` varchar(100)    NOT NULL DEFAULT '' COMMENT '营业时间',
    `contact_name`   varchar(50)     NOT NULL DEFAULT '' COMMENT '联系人',
    `service_phone`  varchar(20)     NOT NULL DEFAULT '' COMMENT '客服电话',
    `address`        varchar(255)    NOT NULL DEFAULT '' COMMENT '详细地址',
    `latitude`       decimal(10, 6)  NOT NULL DEFAULT 0 COMMENT '纬度',
    `longitude`      decimal(10, 6)  NOT NULL DEFAULT 0 COMMENT '经度',
    `average_spend`  bigint          NOT NULL DEFAULT 0 COMMENT '人均消费金额(分)',
    `created_at`     datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_organizer_profile` (`organizer_id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='商家扩展资料表';

CREATE TABLE IF NOT EXISTS `organizer_level_rules`
(
    `id`                      bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '等级规则ID',
    `level`                   int             NOT NULL COMMENT '等级',
    `name`                    varchar(50)     NOT NULL COMMENT '等级名称',
    `fee_rate`                decimal(5,4)    NOT NULL DEFAULT 0 COMMENT '手续费比例',
    `required_activity_count` bigint          NOT NULL DEFAULT 0 COMMENT '升级所需活动数',
    `description`             text            NULL COMMENT '等级说明',
    `benefits`                text            NULL COMMENT '等级权益',
    `status`                  tinyint         NOT NULL DEFAULT 1 COMMENT '1启用 0禁用',
    `created_at`              datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`              datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_org_level_rule_level` (`level`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='商家等级规则表';

CREATE TABLE IF NOT EXISTS `organizer_roles`
(
    `id`           bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '商家角色ID',
    `organizer_id` bigint unsigned NOT NULL COMMENT '主办方ID',
    `name`         varchar(50)     NOT NULL COMMENT '角色名称',
    `description`  varchar(255)    NOT NULL DEFAULT '' COMMENT '角色说明',
    `permissions`  text            NULL COMMENT '权限JSON数组',
    `status`       tinyint         NOT NULL DEFAULT 1 COMMENT '1启用 0禁用',
    `created_at`   datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`   datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_org_role_organizer` (`organizer_id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='商家角色表';

CREATE TABLE IF NOT EXISTS `organizer_staff`
(
    `id`           bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '商家子账号ID',
    `organizer_id` bigint unsigned NOT NULL COMMENT '主办方ID',
    `user_id`      bigint unsigned NOT NULL DEFAULT 0 COMMENT '绑定用户ID',
    `role_id`      bigint unsigned NOT NULL DEFAULT 0 COMMENT '角色ID',
    `name`         varchar(50)     NOT NULL COMMENT '姓名',
    `phone`        varchar(20)     NOT NULL COMMENT '手机号',
    `status`       tinyint         NOT NULL DEFAULT 1 COMMENT '1启用 0禁用',
    `created_at`   datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`   datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_org_staff_organizer` (`organizer_id`) USING BTREE,
    KEY `idx_org_staff_role` (`role_id`) USING BTREE,
    KEY `idx_org_staff_phone` (`phone`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='商家子账号表';

CREATE TABLE IF NOT EXISTS `organizer_operation_logs`
(
    `id`           bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '商家操作日志ID',
    `organizer_id` bigint unsigned NOT NULL COMMENT '主办方ID',
    `operator_id`  bigint unsigned NOT NULL DEFAULT 0 COMMENT '操作用户ID',
    `operator`     varchar(50)     NOT NULL DEFAULT '' COMMENT '操作人',
    `action`       varchar(100)    NOT NULL COMMENT '动作',
    `resource`     varchar(100)    NOT NULL DEFAULT '' COMMENT '资源',
    `method`       varchar(20)     NOT NULL DEFAULT '' COMMENT 'HTTP方法',
    `path`         varchar(255)    NOT NULL DEFAULT '' COMMENT '请求路径',
    `ip`           varchar(50)     NOT NULL DEFAULT '' COMMENT 'IP',
    `remark`       varchar(255)    NOT NULL DEFAULT '' COMMENT '备注',
    `created_at`   datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_org_log_organizer` (`organizer_id`) USING BTREE,
    KEY `idx_org_log_operator` (`operator_id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='商家操作日志表';

CREATE TABLE IF NOT EXISTS `organizer_withdraws`
(
    `id`           bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '提现ID',
    `organizer_id` bigint unsigned NOT NULL COMMENT '主办方ID',
    `amount`       bigint          NOT NULL COMMENT '提现金额(分)',
    `bank_account_name` varchar(50) NOT NULL DEFAULT '' COMMENT '收款人快照',
    `bank_account_no`   varchar(50) NOT NULL DEFAULT '' COMMENT '收款账户快照',
    `bank_name`         varchar(50) NOT NULL DEFAULT '' COMMENT '银行名称快照',
    `status`       tinyint         NOT NULL DEFAULT 0 COMMENT '0待审核 1通过 2拒绝',
    `remark`       varchar(255)    NOT NULL DEFAULT '' COMMENT '审核备注',
    `created_at`   datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`   datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_withdraw_organizer` (`organizer_id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='主办方提现表';

-- Existing organizer_withdraws table migration:
-- ALTER TABLE `organizer_withdraws` ADD COLUMN `bank_account_name` varchar(50) NOT NULL DEFAULT '' COMMENT '收款人快照' AFTER `amount`;
-- ALTER TABLE `organizer_withdraws` ADD COLUMN `bank_account_no` varchar(50) NOT NULL DEFAULT '' COMMENT '收款账户快照' AFTER `bank_account_name`;
-- ALTER TABLE `organizer_withdraws` ADD COLUMN `bank_name` varchar(50) NOT NULL DEFAULT '' COMMENT '银行名称快照' AFTER `bank_account_no`;

CREATE TABLE IF NOT EXISTS `organizer_withdraw_allocations`
(
    `id`           bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
    `withdraw_id`  bigint unsigned NOT NULL COMMENT '提现申请ID',
    `organizer_id` bigint unsigned NOT NULL COMMENT '主办方ID',
    `order_id`     bigint unsigned NOT NULL COMMENT '订单ID',
    `order_no`     varchar(30)     NOT NULL COMMENT '订单号',
    `activity_id`  bigint unsigned NOT NULL COMMENT '活动ID',
    `amount`       bigint          NOT NULL COMMENT '本提现分配金额(分)',
    `status`       tinyint         NOT NULL DEFAULT 0 COMMENT '0提现审核中 1已提现 2已释放',
    `created_at`   datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`   datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_withdraw_order` (`withdraw_id`, `order_id`) USING BTREE,
    KEY `idx_withdraw_allocation_organizer` (`organizer_id`) USING BTREE,
    KEY `idx_withdraw_allocation_order` (`order_id`) USING BTREE,
    KEY `idx_withdraw_allocation_activity` (`activity_id`) USING BTREE,
    KEY `idx_withdraw_allocation_status` (`status`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='主办方提现订单分配明细';

CREATE TABLE IF NOT EXISTS `organizer_bank_account_audits`
(
    `id`                bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '收款账户审核ID',
    `organizer_id`      bigint unsigned NOT NULL COMMENT '主办方ID',
    `user_id`           bigint unsigned NOT NULL COMMENT '提交用户ID',
    `bank_account_name` varchar(50)     NOT NULL COMMENT '收款人',
    `bank_account_no`   varchar(50)     NOT NULL COMMENT '收款账户',
    `bank_name`         varchar(50)     NOT NULL COMMENT '银行名称',
	`bank_contact_name` varchar(50)     NOT NULL DEFAULT '' COMMENT '收款联系人',
	`bank_contact_phone` varchar(20)    NOT NULL DEFAULT '' COMMENT '收款联系电话',
    `status`            tinyint         NOT NULL DEFAULT 0 COMMENT '0待审核 1通过 2拒绝',
    `reject_reason`     varchar(255)    NOT NULL DEFAULT '' COMMENT '拒绝原因',
    `reviewed_at`       datetime        NULL COMMENT '审核时间',
    `created_at`        datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`        datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_bank_audit_organizer` (`organizer_id`) USING BTREE,
    KEY `idx_bank_audit_user` (`user_id`) USING BTREE,
    KEY `idx_bank_audit_status` (`status`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='主办方收款账户审核表';

-- ALTER TABLE `organizer_bank_account_audits` ADD COLUMN `bank_contact_name` varchar(50) NOT NULL DEFAULT '' COMMENT '收款联系人' AFTER `bank_name`;
-- ALTER TABLE `organizer_bank_account_audits` ADD COLUMN `bank_contact_phone` varchar(20) NOT NULL DEFAULT '' COMMENT '收款联系电话' AFTER `bank_contact_name`;

-- Existing viewers table migration:
-- 观演人允许被不同购票账号重复添加，仅限制同一账号内身份证/手机号重复。
-- ALTER TABLE `viewers` DROP INDEX `uk_id_card`;
-- ALTER TABLE `viewers` DROP INDEX `uk_phone`;
-- ALTER TABLE `viewers` ADD UNIQUE KEY `uk_viewer_user_id_card` (`user_id`, `id_card`);
-- ALTER TABLE `viewers` ADD UNIQUE KEY `uk_viewer_user_phone` (`user_id`, `phone`);

-- Existing notes table migration:
-- 场地详情“相关动态”需要动态和场地/门店建立明确关联。
-- ALTER TABLE `notes` ADD COLUMN `store_id` bigint NOT NULL DEFAULT 0 COMMENT '关联门店/场地ID' AFTER `activity_id`;
-- ALTER TABLE `notes` ADD KEY `idx_store_id` (`store_id`);

-- Existing comments table migration. Status is fixed as 1公开、0隐藏、-1软删除.
-- ALTER TABLE `comments` ADD COLUMN `moderated_by` bigint unsigned NOT NULL DEFAULT 0 COMMENT '最后审核管理员ID' AFTER `status`;
-- ALTER TABLE `comments` ADD COLUMN `moderated_at` datetime NULL COMMENT '最后审核时间' AFTER `moderated_by`;
-- ALTER TABLE `comments` ADD COLUMN `moderate_reason` varchar(255) NOT NULL DEFAULT '' COMMENT '审核原因' AFTER `moderated_at`;
-- ALTER TABLE `comments` ADD KEY `idx_comment_moderated_by` (`moderated_by`);

-- Dynamic share event log. This is the source of truth; note_stats.share_count is only a display counter.
CREATE TABLE IF NOT EXISTS `note_shares`
(
    `id`         bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '分享记录ID',
    `note_id`    bigint unsigned NOT NULL COMMENT '动态ID',
    `user_id`    bigint unsigned NOT NULL COMMENT '分享用户ID',
    `channel`    varchar(50)     NOT NULL DEFAULT '' COMMENT 'wechat_session/wechat_timeline/copy_link/poster/other',
    `created_at` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '分享时间',
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_note_share_note_created` (`note_id`, `created_at`) USING BTREE,
    KEY `idx_note_share_user` (`user_id`) USING BTREE,
    KEY `idx_note_share_created` (`created_at`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='动态真实分享日志';

-- Immutable platform accounting ledger. Correcting a business event must append a reversal row instead of updating history.
CREATE TABLE IF NOT EXISTS `platform_finance_flows`
(
    `id`             bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '流水ID',
    `flow_no`        varchar(40)     NOT NULL COMMENT '平台流水号',
    `business_key`   varchar(100)    NOT NULL COMMENT '业务幂等键',
    `type`           varchar(32)     NOT NULL COMMENT 'order_income/refund/service_fee/settlement/withdraw/manual_adjustment',
    `direction`      varchar(16)     NOT NULL COMMENT 'income/expense',
    `amount`         bigint          NOT NULL COMMENT '金额，单位分',
    `order_no`       varchar(30)     NOT NULL DEFAULT '' COMMENT '订单号',
    `refund_no`      varchar(30)     NOT NULL DEFAULT '' COMMENT '退款单号',
    `withdraw_id`    bigint unsigned NOT NULL DEFAULT 0 COMMENT '提现记录ID',
    `organizer_id`   bigint unsigned NOT NULL DEFAULT 0 COMMENT '商家ID',
    `organizer_name` varchar(100)    NOT NULL DEFAULT '' COMMENT '商家名称快照',
    `pay_method`     varchar(20)     NOT NULL DEFAULT '' COMMENT '支付方式快照',
    `remark`         varchar(255)    NOT NULL DEFAULT '' COMMENT '说明',
    `occurred_at`    datetime        NOT NULL COMMENT '业务发生时间',
    `created_at`     datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '入账时间',
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_platform_flow_no` (`flow_no`) USING BTREE,
    UNIQUE KEY `uk_platform_flow_business` (`business_key`) USING BTREE,
    KEY `idx_platform_flow_type_occurred` (`type`, `occurred_at`) USING BTREE,
    KEY `idx_platform_flow_organizer_occurred` (`organizer_id`, `occurred_at`) USING BTREE,
    KEY `idx_platform_flow_order` (`order_no`) USING BTREE,
    KEY `idx_platform_flow_refund` (`refund_no`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT ='平台资金流水';

-- Existing point_logs table migration. It makes request_no idempotency safe under concurrent submissions.
-- ALTER TABLE `point_logs` ADD UNIQUE KEY `uk_point_log_user_source_type` (`user_id`, `source_id`, `change_type`);

INSERT IGNORE INTO `cancel_reasons` (`id`, `reason`, `sort`) VALUES
    (1, '计划有变', 1),
    (2, '买错票券', 2),
    (3, '其他原因', 99);

INSERT IGNORE INTO `refund_reasons` (`id`, `reason`, `sort`) VALUES
    (1, '行程冲突', 1),
    (2, '重复购买', 2),
    (3, '其他原因', 99);
