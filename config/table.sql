
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