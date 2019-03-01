CREATE TABLE `tr_bras` (
  `id` INT(11) NOT NULL AUTO_INCREMENT,
  `identifier` VARCHAR(128) NULL DEFAULT NULL,
  `name` VARCHAR(64) NOT NULL,
  `ipaddr` VARCHAR(32) NULL DEFAULT NULL,
  `vendor_id` VARCHAR(32) NOT NULL,
  `secret` VARCHAR(64) NOT NULL,
  `coa_port` INT(11) NOT NULL,
  `auth_limit` INT(11) NULL DEFAULT NULL,
  `acct_limit` INT(11) NULL DEFAULT NULL,
  `status` ENUM('enabled','disabled') NULL DEFAULT NULL,
  `remark` VARCHAR(512) NULL DEFAULT NULL,
  `create_time` DATETIME NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `ix_tr_bras_identifier` (`identifier`),
  INDEX `ix_tr_bras_ipaddr` (`ipaddr`)
)
  COLLATE='utf8_general_ci'
  ENGINE=InnoDB
;

CREATE TABLE `tr_config` (
  `id` INT(11) NOT NULL AUTO_INCREMENT,
  `type` VARCHAR(32) NOT NULL,
  `name` VARCHAR(128) NOT NULL,
  `value` VARCHAR(255) NULL DEFAULT NULL,
  `remark` VARCHAR(255) NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
)
  COLLATE='utf8_general_ci'
  ENGINE=InnoDB
;


CREATE TABLE `tr_subscribe` (
  `id` INT(11) NOT NULL AUTO_INCREMENT,
   `node_id` INT(11) NOT NULL DEFAULT 0,
   `area_id` int(11) NOT NULL DEFAULT 0,
  `subscriber` VARCHAR(32) NULL DEFAULT NULL,
  `realname` VARCHAR(32) NULL DEFAULT NULL,
  `password` VARCHAR(128) NOT NULL,
  `bill_type` ENUM('flow','time') NOT NULL DEFAULT 'time',
  `domain` VARCHAR(128) NULL DEFAULT NULL,
  `addr_pool` VARCHAR(128) NULL DEFAULT NULL,
  `policy` VARCHAR(512) NULL DEFAULT NULL,
  `is_online` INT(11) NULL DEFAULT NULL,
  `active_num` INT(11) NULL DEFAULT NULL,
  `flow_amount` BIGINT(20) NULL DEFAULT NULL,
  `bind_mac` TINYINT(1) NULL DEFAULT NULL,
  `bind_vlan` TINYINT(1) NULL DEFAULT NULL,
  `ip_addr` VARCHAR(32) NULL DEFAULT NULL,
  `mac_addr` VARCHAR(32) NULL DEFAULT NULL,
  `in_vlan` INT(11) NULL DEFAULT NULL,
  `out_vlan` INT(11) NULL DEFAULT NULL,
  `up_rate` DECIMAL(6,3) NULL DEFAULT NULL,
  `down_rate` DECIMAL(6,3) NULL DEFAULT NULL,
  `up_peak_rate` DECIMAL(6,3) NULL DEFAULT NULL,
  `down_peak_rate` DECIMAL(6,3) NULL DEFAULT NULL,
  `up_rate_code` VARCHAR(32) NULL DEFAULT NULL,
  `down_rate_code` VARCHAR(32) NULL DEFAULT NULL,
  `status` ENUM('enabled','disabled') NULL DEFAULT NULL,
  `remark` VARCHAR(512) NULL DEFAULT NULL,
  `begin_time` DATETIME NOT NULL,
  `expire_time` DATETIME NOT NULL,
  `create_time` DATETIME NOT NULL,
  `update_time` DATETIME NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  INDEX `ix_tr_subscribe_subscriber` (`subscriber`),
  INDEX `ix_tr_subscribe_expire_time` (`expire_time`),
  INDEX `ix_tr_subscribe_status` (`status`),
  INDEX `ix_tr_subscribe_create_time` (`create_time`),
  INDEX `ix_tr_subscribe_update_time` (`update_time`)
)
  COLLATE='utf8_general_ci'
  ENGINE=InnoDB
;
