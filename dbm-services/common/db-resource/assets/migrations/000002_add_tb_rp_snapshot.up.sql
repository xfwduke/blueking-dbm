CREATE TABLE `tb_rp_daily_snap_shot` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `report_day` varchar(32) NOT NULL COMMENT '''上报日期''',
    `bk_cloud_id` int(11) NOT NULL COMMENT '''云区域 ID''',
    `bk_biz_id` int(11) NOT NULL COMMENT '机器当前所属业务',
    `dedicated_biz` int(11) DEFAULT '0' COMMENT '专属业务',
    `rs_type` varchar(64) DEFAULT 'PUBLIC' COMMENT '资源专用组件类型',
    `bk_host_id` int(11) NOT NULL COMMENT '''bk主机ID''',
    `ip` varchar(20) NOT NULL,
    `device_class` varchar(64) NOT NULL,
    `cpu_num` int(11) NOT NULL COMMENT '''cpu核数''',
    `dram_cap` int(11) NOT NULL COMMENT '''内存大小''',
    `storage_device` json DEFAULT NULL COMMENT '''磁盘设备''',
    `total_storage_cap` int(11) DEFAULT NULL COMMENT '''磁盘总容量''',
    `os_type` varchar(32) NOT NULL COMMENT '''操作系统类型''',
    `os_bit` varchar(32) NOT NULL COMMENT '''操作系统位数''',
    `os_version` varchar(64) NOT NULL COMMENT '''操作系统版本''',
    `os_name` varchar(64) NOT NULL COMMENT '''操作系统名称''',
    `city_id` varchar(64) NOT NULL,
    `city` varchar(128) NOT NULL,
    `sub_zone` varchar(32) NOT NULL,
    `sub_zone_id` varchar(64) NOT NULL,
    `label` json DEFAULT NULL,
    `status` varchar(20) NOT NULL,
    `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_host_id` (`bk_host_id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8;