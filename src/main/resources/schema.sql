SET MODE MYSQL;

CREATE SCHEMA IF NOT EXISTS "toughradius";

CREATE TABLE IF NOT EXISTS `tr_nas` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `identifier` varchar(128) DEFAULT NULL,
  `name` varchar(64) NOT NULL,
  `ipaddr` varchar(32) DEFAULT NULL,
  `vendorid` varchar(32) NOT NULL,
  `secret` varchar(64) NOT NULL,
  `coaport` int(11) NOT NULL,
  `status` enum('enabled','disabled') DEFAULT NULL,
  `update_time` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ;

CREATE TABLE IF NOT EXISTS `tr_option` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL,
  `value` varchar(1024) DEFAULT NULL,
  `remark` varchar(255) DEFAULT NULL,
  `update_time` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
);

CREATE TABLE IF NOT EXISTS `tr_group` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL,
  `remark` varchar(255) DEFAULT NULL,
  `update_time` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
);

CREATE TABLE IF NOT EXISTS `tr_user` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `group_id` int(11) NOT NULL,
  `fullname` varchar(255) DEFAULT NULL,
  `email` varchar(255) DEFAULT NULL,
  `mobile` varchar(16) DEFAULT NULL,
  `bill_type` enum('time','flow') DEFAULT 'time',
  `username` varchar(32) DEFAULT NULL,
  `password` varchar(128) NOT NULL,
  `online_num` int(11) DEFAULT 1,
  `bind_mac` int(11) DEFAULT 0,
  `bind_vlan` int(11) DEFAULT 0,
  `in_vlan` int(11) DEFAULT 0,
  `out_vlan` int(11) DEFAULT 0,
  `ip_addr` varchar(32) DEFAULT NULL,
  `mac_addr` varchar(32) DEFAULT NULL,
  `domain` varchar(32) DEFAULT NULL,
  `policy` varchar(32) DEFAULT NULL,
  `addr_pool` varchar(128) DEFAULT NULL,
  `flow_amount` bigint(20) DEFAULT 0,
  `up_rate` bigint(20) DEFAULT 0,
  `up_peak_rate` bigint(20) DEFAULT 0,
  `down_rate` bigint(20) DEFAULT 0,
  `down_peak_rate` bigint(20) DEFAULT 0,
  `status` enum('enabled','disabled','cancel') DEFAULT 'enabled',
  `remark` varchar(512) DEFAULT NULL,
  `expire_time` datetime NOT NULL,
  `create_time` datetime NOT NULL,
  `update_time` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
);

CREATE TABLE IF NOT EXISTS `tr_user_radattr` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `value` varchar(255) DEFAULT NULL,
  `remark` varchar(255) DEFAULT NULL,
  `update_time` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
);

