
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