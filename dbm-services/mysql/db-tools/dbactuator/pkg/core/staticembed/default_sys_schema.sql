SET SESSION TRANSACTION ISOLATION LEVEL REPEATABLE READ;
set binlog_format=statement;
CREATE DATABASE if not exists test;
CREATE DATABASE IF NOT EXISTS `infodba_schema` DEFAULT CHARACTER SET utf8;
create table IF NOT EXISTS infodba_schema.free_space(a int) engine = InnoDB;
CREATE TABLE if not exists infodba_schema.conn_log(
    conn_id bigint default NULL,
    conn_time datetime default NULL,
    user_name varchar(128) default NULL,
    cur_user_name varchar(128) default NULL,
    ip varchar(15) default NULL,
    key conn_time(conn_time)
) engine = InnoDB;

create table if not exists infodba_schema.`checksum`(
    master_ip char(32) NOT NULL DEFAULT '0.0.0.0',
    master_port int(11) NOT NULL DEFAULT '3306',
    db char(64) NOT NULL,
    tbl char(64) NOT NULL,
    chunk int(11) NOT NULL,
    chunk_time float DEFAULT NULL,
    chunk_index varchar(200) DEFAULT NULL,
    lower_boundary blob,
    upper_boundary blob,
    this_crc char(40) NOT NULL,
    this_cnt int(11) NOT NULL,
    master_crc char(40) DEFAULT NULL,
    master_cnt int(11) DEFAULT NULL,
    ts timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`master_ip`,`master_port`,`db`,`tbl`,`chunk`),
    KEY `ts_db_tbl` (`ts`,`db`,`tbl`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
replace into infodba_schema.checksum
values('0.0.0.0','3306', 'test', 'test', 0, NULL, NULL, '1=1', '1=1', '0', 0, '0', 0, now());

CREATE TABLE if not exists infodba_schema.`checksum_history` (
   `master_ip` char(32) NOT NULL DEFAULT '0.0.0.0',
   `master_port` int(11) NOT NULL DEFAULT '3306',
   `db` char(64) NOT NULL,
   `tbl` char(64) NOT NULL,
   `chunk` int(11) NOT NULL,
   `chunk_time` float DEFAULT NULL,
   `chunk_index` varchar(200) DEFAULT NULL,
   `lower_boundary` blob,
   `upper_boundary` blob,
   `this_crc` char(40) NOT NULL,
   `this_cnt` int(11) NOT NULL,
   `master_crc` char(40) DEFAULT NULL,
   `master_cnt` int(11) DEFAULT NULL,
   `ts` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
   `reported` int(11) DEFAULT '0',
   PRIMARY KEY (`master_ip`,`master_port`,`db`,`tbl`,`chunk`,`ts`),
   KEY `ts_db_tbl` (`ts`,`db`,`tbl`),
   KEY `idx_reported` (`reported`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE if not exists infodba_schema.spes_status(
    ip varchar(15) default '',
    spes_id smallint default 0,
    report_day int default 0,
    PRIMARY KEY ip_id_day (ip, spes_id, report_day)
) engine = InnoDB;
CREATE TABLE IF NOT EXISTS infodba_schema.check_heartbeat (
    uid INT UNSIGNED  NOT NULL PRIMARY KEY,
    ck_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP on  UPDATE CURRENT_TIMESTAMP
) ENGINE = InnoDB;
REPLACE INTO infodba_schema.check_heartbeat(uid) value(@@server_id);
CREATE TABLE IF NOT EXISTS infodba_schema.query_response_time(
    time_min INT(11) NOT NULL DEFAULT '0',
    time VARCHAR(14) NOT NULL DEFAULT '',
    total VARCHAR(100) NOT NULL DEFAULT '',
    update_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (time_min, time)
) engine = InnoDB;
-- conn_log 所有用户可写. 注会导致所有用户可以看见 infodba_schema
REPLACE into `mysql`.`db`(`Host`,`Db`,`User`,`Select_priv`,`Insert_priv`, `Update_priv`,`Delete_priv`,`Create_priv`,`Drop_priv`)
 values('%','infodba_schema','','Y','Y',  'N','N','N','N');

-- 语句来自 mysql-dbbackup dbareport
CREATE TABLE IF NOT EXISTS infodba_schema.local_backup_report (
    backup_id varchar(64) NOT NULL,
    mysql_role varchar(30) NOT NULL DEFAULT '',
    shard_value int(11) NOT NULL DEFAULT 0,
    backup_type varchar(30) NOT NULL,
    cluster_id int(11) NOT NULL,
    cluster_address varchar(255) DEFAULT NULL,
    backup_host varchar(30) NOT NULL,
    backup_port int(11) NOT NULL,
    server_id varchar(10) DEFAULT NULL,
    bill_id varchar(30) DEFAULT NULL,
    bk_biz_id int(11) DEFAULT NULL,
    mysql_version varchar(60) DEFAULT NULL,
    data_schema_grant varchar(30) DEFAULT NULL,
    is_full_backup tinyint(4) DEFAULT NULL,
    backup_begin_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    backup_end_time timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
    backup_consistent_time timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
    backup_status varchar(60) DEFAULT NULL,
    backup_meta_file varchar(255),
    binlog_info text,
    file_list text,
    extra_fields text,
    backup_config_file text,
    PRIMARY KEY (backup_id,mysql_role,shard_value)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS infodba_schema.proxy_user_list(
    proxy_ip varchar(32) NOT NULL,
    username varchar(64) NOT NULL,
    host varchar(32) NOT NULL,
    create_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (proxy_ip, username, host),
    KEY IDX_USERNAME_HOST(username, host, create_at),
    KEY IDX_HOST(host, create_at),
    KEY IDX_IP_HOST(proxy_ip, host, create_at)
) ENGINE=InnoDB;

flush privileges;
flush logs;



-- 授权, 检查 ERROR_MSG 来判断是否成功
-- ERROR STATE CODE
-- 32401 参数验证错误
-- 32402 冲突检测错误
-- 中控不会转发授权语句, 当普通 MySQL 就好
DROP PROCEDURE IF EXISTS infodba_schema.dba_grant;
DELIMITER //
CREATE PROCEDURE infodba_schema.dba_grant(
    IN username VARCHAR(128), 
    IN ip_list VARCHAR(3000), -- 但是限制最大只能传入 2000
    IN db_list VARCHAR(3000), -- 但是限制最大只能传入 2000
    IN long_psw VARCHAR(128), -- 密码是密文
    IN short_psw VARCHAR(32), -- 密码是密文
    IN priv_str VARCHAR(4096), 
    IN global_priv_str VARCHAR(4096)
)
SQL SECURITY INVOKER
BEGIN
    IF LENGTH(ip_list) >= 2000 OR LENGTH(db_list) >= 2000 THEN
        SIGNAL SQLSTATE '32401' SET MESSAGE_TEXT = "input ip_list or db_list too long, max length is 2000";
    END IF;

    IF NOT(long_psw LIKE '*%' AND LENGTH(long_psw) = 41) THEN
        SET @msg = CONCAT('bad password: ', long_psw);
        SIGNAL SQLSTATE '32401' SET MESSAGE_TEXT = @msg;
    END IF;

    -- 不同版本授权语句不兼容, 所以干脆不写
    SET SESSION sql_log_bin = 0;

    -- 初始化结果表
    CALL init_report_table();

    -- 初始化结果标识
    SET @uuid = UUID();
    SET @grant_time = NOW();
    
    SET ip_list = TRIM(BOTH ',' FROM ip_list);
    SET ip_list = CONCAT(ip_list, ",");

    SET db_list = TRIM(BOTH ',' FROM db_list);
    SET db_list = CONCAT(db_list, ",");    

    -- 先做检查
    SET @is_check_failed = 0;
    CALL check_all(@uuid, @grant_time, username, ip_list, db_list, long_psw, short_psw, @is_check_failed);

    IF @is_check_failed = 1 THEN
        SIGNAL SQLSTATE '32402' SET MESSAGE_TEXT = @uuid;
    END IF;

    WHILE (LOCATE(',', ip_list) > 0)
    DO
        SET @ip = TRIM(SUBSTRING(ip_list, 1, LOCATE(',', ip_list) - 1));
        SET ip_list = SUBSTRING(ip_list, LOCATE(',', ip_list) + 1);
        -- 如果涉及新增账号, 只使用新版本密码
        CALL dba_grant_one_ip(username, @ip, db_list, long_psw, priv_str, global_priv_str);
    END WHILE;

    FLUSH PRIVILEGES;
END//
DELIMITER ;

DROP PROCEDURE IF EXISTS infodba_schema.init_report_table;
DROP TABLE IF EXISTS infodba_schema.dba_grant_result;
DELIMITER //
CREATE PROCEDURE infodba_schema.init_report_table()
SQL SECURITY INVOKER
BEGIN
    -- 不同版本授权语句不兼容, 所以干脆不写
    SET SESSION sql_log_bin = 0;

    CREATE TABLE IF NOT EXISTS infodba_schema.dba_grant_result(
        id VARCHAR(64),
        grant_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        username VARCHAR(128),
        client_ip VARCHAR(32),
        dbname VARCHAR(64),
        long_psw VARCHAR(128),
        short_psw VARCHAR(32),
        priv VARCHAR(4096),
        global_priv VARCHAR(4096),
        msg VARCHAR(4096)
    ) ENGINE = InnoDB;
    DELETE FROM dba_grant_result WHERE grant_time < DATE_SUB(NOW(), INTERVAL 1 DAY);
END//
DELIMITER ;

-- 检查入口
DROP PROCEDURE IF EXISTS infodba_schema.check_all;
DELIMITER //
CREATE PROCEDURE infodba_schema.check_all(
    IN uuid VARCHAR(64),
    IN grant_time TIMESTAMP,
    IN username VARCHAR(128), 
    IN ip_list VARCHAR(3000), 
    IN db_list VARCHAR(3000), 
    IN long_psw VARCHAR(128),
    IN short_psw VARCHAR(32),
    OUT is_check_failed INT
)
SQL SECURITY INVOKER
BEGIN
    -- 全量检查入口
    CALL check_password(uuid, grant_time, username, ip_list, long_psw, short_psw, is_check_failed);
    CALL check_db_conflict(uuid, grant_time, username, ip_list, db_list, is_check_failed);
END //
DELIMITER ;

-- 密码一致性检查
DROP PROCEDURE IF EXISTS infodba_schema.check_password;
DELIMITER //
CREATE PROCEDURE infodba_schema.check_password(
    IN uuid VARCHAR(64),
    IN grant_time TIMESTAMP,    
    IN username VARCHAR(128), 
    IN ip_list VARCHAR(3000), 
    IN long_psw VARCHAR(128),
    IN short_psw VARCHAR(32),
    OUT is_check_failed INT
)
SQL SECURITY INVOKER
BEGIN
    -- 不同版本授权语句不兼容, 所以干脆不写
    SET SESSION sql_log_bin = 0;

    WHILE (LOCATE(',', ip_list) > 0)
    DO
        SET @ip = TRIM(SUBSTRING(ip_list, 1, LOCATE(',', ip_list) - 1));
        SET ip_list = SUBSTRING(ip_list, LOCATE(',', ip_list) + 1);  

        SELECT EXISTS(SELECT 1 FROM mysql.user WHERE user = username AND host = @ip) INTO @user_host_exists;
        IF @user_host_exists THEN
            -- 用户存在, 检查密码
            -- 5.6 及之前还有 old_password 函数, 也就是 < 5.7 可能有 old_password
            IF SUBSTRING_INDEX(@@version, ".", 2) < 5.7 THEN
                SELECT password = long_psw OR password = short_psw INTO @psw_match FROM mysql.user WHERE user = username AND host = @ip;
            ELSE
                SELECT authentication_string = long_psw INTO @psw_match FROM mysql.user WHERE user = username AND host =@ip;
            END IF;   

            IF NOT @psw_match THEN
                SET is_check_failed = is_check_failed OR 1;
                INSERT INTO dba_grant_result(id, grant_time, username, client_ip, long_psw, short_psw, msg) 
                    VALUES (uuid, grant_time, username, @ip, long_psw, short_psw, 'password not match');
            END IF;     
        END IF;
    END WHILE;
END //
DELIMITER ;

-- 库模式冲突检查入口
DROP PROCEDURE IF EXISTS infodba_schema.check_db_conflict;
DELIMITER //
CREATE PROCEDURE infodba_schema.check_db_conflict(
    IN uuid VARCHAR(64),
    IN grant_time TIMESTAMP,
    IN username VARCHAR(128), 
    IN ip_list VARCHAR(3000), 
    IN db_list VARCHAR(3000),
    OUT is_check_failed INT
)
SQL SECURITY INVOKER
BEGIN
    WHILE (LOCATE(',', ip_list) > 0)
    DO
        SET @ip = TRIM(SUBSTRING(ip_list, 1, LOCATE(',', ip_list) - 1));
        SET ip_list = SUBSTRING(ip_list, LOCATE(',', ip_list) + 1);  

        SELECT EXISTS(SELECT 1 FROM mysql.db WHERE user = username AND host = @ip) INTO @db_priv_applied;
        IF @db_priv_applied THEN
            CALL check_db_conflict_one_ip(uuid, grant_time, username, @ip, db_list, is_check_failed);
        END IF;
    END WHILE;
END //
DELIMITER ;

-- 单 IP 库模式冲突检查
DROP PROCEDURE IF EXISTS infodba_schema.check_db_conflict_one_ip;
DELIMITER //
CREATE PROCEDURE infodba_schema.check_db_conflict_one_ip(
    IN uuid VARCHAR(64),
    IN grant_time TIMESTAMP,    
    IN username VARCHAR(128), 
    IN ip VARCHAR(15), 
    IN db_list VARCHAR(3000),
    OUT is_check_failed INT
)
SQL SECURITY INVOKER
BEGIN
    DECLARE applied_dbname VARCHAR(64);
    DECLARE cursor_done INT DEFAULT FALSE;
    DECLARE db_cursor CURSOR FOR SELECT db FROM mysql.db WHERE user = username AND host = @ip;
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET cursor_done = TRUE;    

    -- 不同版本授权语句不兼容, 所以干脆不写
    SET SESSION sql_log_bin = 0;

    OPEN db_cursor;
    fetch_loop: LOOP
        FETCH db_cursor INTO applied_dbname;
        IF cursor_done THEN
            LEAVE fetch_loop;
        END IF;

        SET @loop_db_list = db_list;

        WHILE (LOCATE(',', @loop_db_list) > 0)
        DO
            SET @dbname = TRIM(SUBSTRING(@loop_db_list, 1, LOCATE(',', @loop_db_list) - 1));
            SET @loop_db_list = SUBSTRING(@loop_db_list, LOCATE(',', @loop_db_list) + 1);

            -- 申请库名和已有库名不等并且能模式匹配 
            IF @dbname <> applied_dbname AND (@dbname LIKE applied_dbname OR applied_dbname LIKE @dbname) THEN 
                SET is_check_failed = is_check_failed OR 1;
                INSERT INTO dba_grant_result(id, grant_time, username, client_ip, dbname, long_psw, short_psw, msg) 
                    VALUES (uuid, grant_time, username, @ip, @dbname, long_psw, short_psw, CONCAT("conflict with applied db [", applied_dbname, "]"));                
            END IF;
        END WHILE;

    END LOOP fetch_loop;
    CLOSE db_cursor;
END //
DELIMITER ;

-- 单 IP 授权
DROP PROCEDURE IF EXISTS infodba_schema.dba_grant_one_ip;
DELIMITER //
CREATE PROCEDURE infodba_schema.dba_grant_one_ip(
    IN username VARCHAR(128), 
    IN ip VARCHAR(15), 
    IN db_list VARCHAR(3000), 
    IN psw VARCHAR(128), 
    IN priv_str VARCHAR(4096), 
    IN global_priv_str VARCHAR(4096)
)
SQL SECURITY INVOKER
BEGIN
    -- 不同版本授权语句不兼容, 所以干脆不写
    SET SESSION sql_log_bin = 0;

    SELECT EXISTS(SELECT 1 FROM mysql.user WHERE user = username AND host = ip) INTO @user_host_exists;
    IF NOT @user_host_exists THEN
        -- 用户不存在, 直接创建, 只使用新版本密码, 传入的密码是密文
        IF SUBSTRING_INDEX(@@version, ".", 2) < 5.7 THEN
            SET @create_user_sql = CONCAT("GRANT USAGE ON *.* TO '", username, "'@'", ip, "' IDENTIFIED BY PASSWORD '", psw, "'");
            PREPARE stmt FROM @create_user_sql;
            EXECUTE stmt;
        ELSE
            SET @create_user_sql = CONCAT("CREATE USER IF NOT EXISTS '", username, "'@'", ip, "' IDENTIFIED WITH mysql_native_password AS '", psw, "'");
            PREPARE stmt FROM @create_user_sql;
            EXECUTE stmt;            
        END IF;
    END IF;

    WHILE (LOCATE(',', db_list) > 0)
    DO
        SET @db = TRIM(SUBSTRING(db_list, 1, LOCATE(',', db_list) - 1));
        SET db_list = SUBSTRING(db_list, LOCATE(',', db_list) + 1);
        CALL dba_grant_one_ip_db(username, ip, @db, psw, priv_str, global_priv_str);
    END WHILE;
END//
DELIMITER ;

-- 单 IP 单 DB 授权
DROP PROCEDURE IF EXISTS infodba_schema.dba_grant_one_ip_db;
DELIMITER //
CREATE PROCEDURE infodba_schema.dba_grant_one_ip_db(
    IN username VARCHAR(128), 
    IN ip VARCHAR(15), 
    IN dbname VARCHAR(64), 
    IN psw VARCHAR(128), 
    IN priv_str VARCHAR(4096), 
    IN global_priv_str VARCHAR(4096)
)
SQL SECURITY INVOKER
BEGIN
    DECLARE exists_db VARCHAR(64);
    DECLARE cursor_done INT DEFAULT FALSE;
    DECLARE db_cursor CURSOR FOR SELECT db FROM mysql.db WHERE user = username AND host = ip;
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET cursor_done = TRUE;

    -- 不同版本授权语句不兼容, 所以干脆不写
    SET SESSION sql_log_bin = 0;

    SET priv_str = TRIM(BOTH ',' FROM TRIM(priv_str));
    SET global_priv_str = TRIM(BOTH ',' FROM TRIM(global_priv_str));

    -- 全局权限
    IF global_priv_str IS NOT NULL AND global_priv_str <> '' THEN
        SET @global_grant_sql = CONCAT("GRANT ", global_priv_str, " ON *.* TO '", username, "'@'", ip, "'");
        PREPARE stmt FROM @global_grant_sql;
        EXECUTE stmt;
    END IF;

    -- 非全局权限
    IF priv_str IS NOT NULL AND priv_str <> '' THEN
        SET @grant_sql = "";
        if dbname = '*' OR dbname = '%' THEN
            SET @grant_sql = CONCAT("GRANT ", priv_str, " ON *.* TO '", username, "'@'", ip, "'");
        ELSE
            SET @grant_sql = CONCAT("GRANT ", priv_str, " ON `", dbname, "`.* TO '", username, "'@'", ip, "'");
        END IF;
        PREPARE stmt FROM @grant_sql;
        EXECUTE stmt;
    END IF;
END//
DELIMITER ;
