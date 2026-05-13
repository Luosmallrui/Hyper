
CREATE TABLE IF NOT EXISTS `users`
(
    `id`         int unsigned     NOT NULL AUTO_INCREMENT COMMENT '用户ID',
    `mobile`     varchar(11)      NOT NULL DEFAULT '' COMMENT '手机号',
    `nickname`   varchar(64)      NOT NULL DEFAULT '' COMMENT '用户昵称',
    `avatar`     varchar(255)     NOT NULL DEFAULT '' COMMENT '用户头像',
    `gender`     tinyint unsigned NOT NULL DEFAULT '3' COMMENT '用户性别[1:男 ;2:女;3:未知]',
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
    `password`   varchar(255) NOT NULL COMMENT '密码(bcrypt)',
    `avatar`     varchar(255) NOT NULL DEFAULT '' COMMENT '头像',
    `gender`     tinyint      NOT NULL DEFAULT 3 COMMENT '性别 1:男 2:女 3:未知',
    `mobile`     varchar(11)  NOT NULL DEFAULT '' COMMENT '手机号',
    `email`      varchar(50)  NOT NULL DEFAULT '' COMMENT '邮箱',
    `motto`      varchar(500) NOT NULL DEFAULT '' COMMENT '座右铭',
    `status`     tinyint      NOT NULL DEFAULT 1 COMMENT '状态 1:正常 2:停用',
    `created_at` datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_username` (`username`) USING BTREE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci COMMENT ='管理员表';