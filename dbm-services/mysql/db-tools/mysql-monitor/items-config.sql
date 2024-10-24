DELETE FROM tb_config_name_def WHERE namespace = 'tendb' AND  conf_type = 'mysql_monitor' AND conf_file = 'items-config.yaml';
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'character-consistency', 'STRING', '{"role":[],"schedule":"0 0 14 * * 1","machine_type":["single","backend","remote","spider"],"name":"character-consistency","enable":true}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'routine-definer', 'STRING', '{"role":[],"machine_type":["single","backend","remote"],"schedule":"0 0 15 * * 1","name":"routine-definer","enable":true}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'view-definer', 'STRING', '{"enable":true,"name":"view-definer","machine_type":["single","backend","remote"],"schedule":"0 0 15 * * 1","role":[]}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'trigger-definer', 'STRING', '{"role":[],"schedule":"0 0 15 * * 1","machine_type":["single","backend","remote"],"name":"trigger-definer","enable":true}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'engine', 'STRING', '{"enable":true,"name":"engine","schedule":"0 10 1 * * *","machine_type":["single","backend","remote"],"role":["slave","orphan"]}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'ext3-check', 'STRING', '{"enable":true,"name":"ext3-check","schedule":"0 0 16 * * 1","machine_type":["single","backend","remote"],"role":[]}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'ibd-statistic', 'STRING', '{"role":["slave","orphan"],"machine_type":["single","backend","remote"],"schedule":"0 45 23 * * *","name":"ibd-statistic","enable":true}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'master-slave-heartbeat', 'STRING', '{"role":["master","repeater","slave","spider_master"],"schedule":"@every 1m","machine_type":["backend","remote","spider"],"name":"master-slave-heartbeat","enable":true}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'mysql-config-diff', 'STRING', '{"role":[],"machine_type":["single","backend","remote","spider"],"schedule":"0 5 10 * * *","name":"mysql-config-diff","enable":true}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'mysql-connlog-size', 'STRING', '{"schedule":"0 0 12 * * *","machine_type":["single","backend","remote","spider"],"role":[],"enable":false,"name":"mysql-connlog-size"}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'mysql-connlog-rotate', 'STRING', '{"role":[],"machine_type":["single","backend","remote","spider"],"schedule":"0 30 23 * * *","name":"mysql-connlog-rotate","enable":false}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'mysql-err-notice', 'STRING', '{"role":[],"schedule":"@every 1m","machine_type":["single","backend","remote"],"name":"mysql-err-notice","enable":false}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'mysql-err-critical', 'STRING', '{"schedule":"@every 1m","machine_type":["single","backend","remote"],"role":[],"enable":false,"name":"mysql-err-critical"}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'spider-err-notice', 'STRING', '{"enable":false,"name":"spider-err-notice","machine_type":["spider"],"schedule":"@every 1m","role":[]}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'spider-err-warn', 'STRING', '{"schedule":"@every 1m","machine_type":["spider"],"role":[],"enable":false,"name":"spider-err-warn"}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'spider-err-critical', 'STRING', '{"machine_type":["spider"],"schedule":"@every 1m","role":[],"enable":false,"name":"spider-err-critical"}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'mysqld-restarted', 'STRING', '{"role":[],"machine_type":["single","backend","remote","spider"],"schedule":"@every 1m","name":"mysqld-restarted","enable":true}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'mysql-lock', 'STRING', '{"enable":true,"name":"mysql-lock","machine_type":["single","backend","remote","spider"],"schedule":"@every 1m","role":["master","spider_master","orphan"]}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'mysql-inject', 'STRING', '{"schedule":"@every 1m","machine_type":["single","backend","spider"],"role":[],"enable":true,"name":"mysql-inject"}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'proxy-backend', 'STRING', '{"enable":true,"name":"proxy-backend","schedule":"@every 1m","machine_type":["proxy"],"role":[]}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'proxy-user-list', 'STRING', '{"role":[],"schedule":"0 55 23 * * *","machine_type":["proxy"],"name":"proxy-user-list","enable":true}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'rotate-slowlog', 'STRING', '{"enable":true,"name":"rotate-slowlog","schedule":"0 55 23 * * *","machine_type":["single","backend","remote","spider"],"role":[]}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'slave-status', 'STRING', '{"schedule":"@every 1m","machine_type":["backend","remote"],"role":["slave","repeater"],"enable":true,"name":"slave-status"}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'ctl-replicate', 'STRING', '{"name":"ctl-replicate","enable":true,"role":["spider_master"],"machine_type":["spider"],"schedule":"@every 1m"}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'spider-remote', 'STRING', '{"schedule":"@every 1m","machine_type":["spider"],"role":[],"enable":true,"name":"spider-remote"}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'spider-table-schema-consistency', 'STRING', '{"role":["spider_master"],"schedule":"0 10 1 * * *","machine_type":["spider"],"name":"spider-table-schema-consistency","enable":false}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'dbha-heartbeat', 'STRING', '{"name":"dbha-heartbeat","enable":true,"role":[],"machine_type":["spider","remote","backend"],"schedule":"@every 2h"}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'unique-ctl-master', 'STRING', '{"enable":false,"name":"unique-ctl-master","schedule":"@every 1m","machine_type":["spider"],"role":["spider_master"]}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'scene-snapshot', 'STRING', '{"role":[],"schedule":"@every 1m","machine_type":["spider","remote","backend","single"],"name":"scene-snapshot","enable":false}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'mysql-timezone-change', 'STRING', '{"schedule":"@every 1m","machine_type":["spider","remote","backend","single"],"role":[],"enable":true,"name":"mysql-timezone-change"}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'sys-timezone-change', 'STRING', '{"enable":true,"name":"sys-timezone-change","schedule":"@every 1m","machine_type":["spider","proxy","remote","backend","single"],"role":[]}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'rotate-proxy-connlog', 'STRING', '{"role":[],"schedule":"0 55 23 * * *","machine_type":["proxy"],"name":"rotate-proxy-connlog","enable":true}', '', 'MAP', 1, 0, 0, 0, 1);
REPLACE INTO tb_config_name_def( namespace, conf_type, conf_file, conf_name, value_type, value_default, value_allowed, value_type_sub, flag_status, flag_disable, flag_locked, flag_encrypt, need_restart) VALUES( 'tendb', 'mysql_monitor', 'items-config.yaml', 'get-ctl-primary', 'STRING', '{"role":["spider_master"],"machine_type":["spider"],"schedule":"@every 1m","name":"get-ctl-primary","enable":false}', '', 'MAP', 1, 0, 0, 0, 1);
