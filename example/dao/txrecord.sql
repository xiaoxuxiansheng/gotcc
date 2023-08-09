CREATE TABLE IF NOT EXISTS `tx_record`
(
    `id`                       bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `status`                   varchar(16) NOT NULL COMMENT '事务状态 hanging/successful/failure',
    `component_try_statuses`   json DEFAULT NULL COMMENT '各组件 try 接口请求状态 hanging/successful/failure',
    `deleted_at`        datetime     DEFAULT NULL COMMENT '删除时间',
    `created_at`        datetime     NOT NULL COMMENT '创建时间',
    `updated_at`        datetime     DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`) USING BTREE COMMENT '主键索引',
    KEY `idx_status` (`status`) COMMENT '事务状态索引'
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT '事务日志记录';