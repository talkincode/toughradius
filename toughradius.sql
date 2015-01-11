/*
 Navicat Premium Data Transfer

 Source Server         : 127.0.0.1
 Source Server Type    : MySQL
 Source Server Version : 50613
 Source Host           : localhost
 Source Database       : toughradius

 Target Server Type    : MySQL
 Target Server Version : 50613
 File Encoding         : utf-8

 Date: 01/12/2015 00:42:08 AM
*/

SET NAMES utf8;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
--  Table structure for `slc_member`
-- ----------------------------
DROP TABLE IF EXISTS `slc_member`;
CREATE TABLE `slc_member` (
  `member_id` int(11) NOT NULL AUTO_INCREMENT,
  `node_id` int(11) NOT NULL,
  `realname` varchar(64) NOT NULL,
  `idcard` varchar(32) DEFAULT NULL,
  `sex` smallint(6) DEFAULT NULL,
  `age` int(11) DEFAULT NULL,
  `email` varchar(255) DEFAULT NULL,
  `mobile` varchar(16) DEFAULT NULL,
  `address` varchar(255) DEFAULT NULL,
  `create_time` varchar(19) NOT NULL,
  `update_time` varchar(19) NOT NULL,
  PRIMARY KEY (`member_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
--  Table structure for `slc_member_order`
-- ----------------------------
DROP TABLE IF EXISTS `slc_member_order`;
CREATE TABLE `slc_member_order` (
  `order_id` varchar(32) NOT NULL,
  `member_id` int(11) NOT NULL,
  `product_id` int(11) NOT NULL,
  `account_number` varchar(32) NOT NULL,
  `order_fee` int(11) NOT NULL,
  `actual_fee` int(11) NOT NULL,
  `pay_status` int(11) NOT NULL,
  `order_source` varchar(64) NOT NULL,
  `create_time` varchar(19) NOT NULL,
  PRIMARY KEY (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
--  Table structure for `slc_member_refund`
-- ----------------------------
DROP TABLE IF EXISTS `slc_member_refund`;
CREATE TABLE `slc_member_refund` (
  `refund_id` varchar(32) NOT NULL,
  `member_id` int(11) NOT NULL,
  `product_id` int(11) NOT NULL,
  `account_number` varchar(32) NOT NULL,
  `refund_fee` int(11) NOT NULL,
  `status` int(11) NOT NULL,
  `order_detail` varchar(1024) NOT NULL,
  `create_time` varchar(19) NOT NULL,
  PRIMARY KEY (`refund_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
--  Table structure for `slc_node`
-- ----------------------------
DROP TABLE IF EXISTS `slc_node`;
CREATE TABLE `slc_node` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `node_name` varchar(32) NOT NULL,
  `node_desc` varchar(64) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;

-- ----------------------------
--  Records of `slc_node`
-- ----------------------------
BEGIN;
INSERT INTO `slc_node` VALUES ('1', 'default', 'default');
COMMIT;

-- ----------------------------
--  Table structure for `slc_param`
-- ----------------------------
DROP TABLE IF EXISTS `slc_param`;
CREATE TABLE `slc_param` (
  `param_name` varchar(64) NOT NULL,
  `param_value` varchar(255) NOT NULL,
  `param_desc` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`param_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
--  Records of `slc_param`
-- ----------------------------
BEGIN;
INSERT INTO `slc_param` VALUES ('max_session_timeout', '86400', '最大会话时长(秒)'), ('reject_delay', '7', '拒绝延迟时间(秒),0-9');
COMMIT;

-- ----------------------------
--  Table structure for `slc_rad_account`
-- ----------------------------
DROP TABLE IF EXISTS `slc_rad_account`;
CREATE TABLE `slc_rad_account` (
  `account_number` varchar(32) NOT NULL,
  `member_id` int(11) NOT NULL,
  `product_id` int(11) NOT NULL,
  `group_id` varchar(32) DEFAULT NULL,
  `password` varchar(128) NOT NULL,
  `status` int(11) NOT NULL,
  `install_address` varchar(128) NOT NULL,
  `balance` int(11) NOT NULL,
  `time_length` int(11) NOT NULL,
  `expire_date` varchar(10) NOT NULL,
  `user_concur_number` int(11) NOT NULL,
  `user_mac` smallint(6) NOT NULL,
  `user_vlan` smallint(6) NOT NULL,
  `mac_addr` varchar(17) DEFAULT NULL,
  `vlan_id` int(11) DEFAULT NULL,
  `vlan_id2` int(11) DEFAULT NULL,
  `ip_address` varchar(15) DEFAULT NULL,
  `create_time` varchar(19) NOT NULL,
  `update_time` varchar(19) NOT NULL,
  PRIMARY KEY (`account_number`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
--  Table structure for `slc_rad_account_attr`
-- ----------------------------
DROP TABLE IF EXISTS `slc_rad_account_attr`;
CREATE TABLE `slc_rad_account_attr` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `account_number` varchar(32) NOT NULL,
  `attr_name` varchar(255) NOT NULL,
  `attr_value` varchar(255) NOT NULL,
  `attr_desc` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
--  Table structure for `slc_rad_bas`
-- ----------------------------
DROP TABLE IF EXISTS `slc_rad_bas`;
CREATE TABLE `slc_rad_bas` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `vendor_id` varchar(32) NOT NULL,
  `ip_addr` varchar(15) NOT NULL,
  `bas_name` varchar(64) NOT NULL,
  `bas_secret` varchar(64) NOT NULL,
  `time_type` smallint(6) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;

-- ----------------------------
--  Records of `slc_rad_bas`
-- ----------------------------
BEGIN;
INSERT INTO `slc_rad_bas` VALUES ('1', '0', '192.168.1.102', 'test_bas', '123456', '0');
COMMIT;

-- ----------------------------
--  Table structure for `slc_rad_group`
-- ----------------------------
DROP TABLE IF EXISTS `slc_rad_group`;
CREATE TABLE `slc_rad_group` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `group_name` varchar(64) NOT NULL,
  `group_desc` varchar(255) DEFAULT NULL,
  `bind_mac` smallint(6) NOT NULL,
  `bind_vlan` smallint(6) NOT NULL,
  `concur_number` int(11) NOT NULL,
  `update_time` varchar(19) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
--  Table structure for `slc_rad_online`
-- ----------------------------
DROP TABLE IF EXISTS `slc_rad_online`;
CREATE TABLE `slc_rad_online` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `account_number` varchar(32) NOT NULL,
  `nas_addr` varchar(32) NOT NULL,
  `acct_session_id` varchar(64) NOT NULL,
  `acct_start_time` varchar(19) NOT NULL,
  `framed_ipaddr` varchar(32) NOT NULL,
  `mac_addr` varchar(32) NOT NULL,
  `nas_port_id` varchar(64) NOT NULL,
  `start_source` smallint(6) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=MEMORY DEFAULT CHARSET=utf8;

-- ----------------------------
--  Table structure for `slc_rad_operator`
-- ----------------------------
DROP TABLE IF EXISTS `slc_rad_operator`;
CREATE TABLE `slc_rad_operator` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `node_id` int(11) NOT NULL,
  `operator_type` int(11) NOT NULL,
  `operator_name` varchar(32) NOT NULL,
  `operator_pass` varchar(128) NOT NULL,
  `operator_status` int(11) NOT NULL,
  `operator_desc` varchar(255) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;

-- ----------------------------
--  Records of `slc_rad_operator`
-- ----------------------------
BEGIN;
INSERT INTO `slc_rad_operator` VALUES ('1', '1', '1', 'admin', '63a9f0ea7bb98050796b649e85481845', '1', 'admin');
COMMIT;

-- ----------------------------
--  Table structure for `slc_rad_product`
-- ----------------------------
DROP TABLE IF EXISTS `slc_rad_product`;
CREATE TABLE `slc_rad_product` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `product_name` varchar(64) NOT NULL,
  `product_policy` int(11) NOT NULL,
  `product_status` smallint(6) NOT NULL,
  `bind_mac` smallint(6) NOT NULL,
  `bind_vlan` smallint(6) NOT NULL,
  `concur_number` int(11) NOT NULL,
  `fee_period` varchar(11) DEFAULT NULL,
  `fee_price` int(11) NOT NULL,
  `input_max_limit` int(11) NOT NULL,
  `output_max_limit` int(11) NOT NULL,
  `create_time` varchar(19) NOT NULL,
  `update_time` varchar(19) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;

-- ----------------------------
--  Records of `slc_rad_product`
-- ----------------------------
BEGIN;
INSERT INTO `slc_rad_product` VALUES ('1', '10元包月套餐', '0', '1', '0', '0', '0', '0', '0', '102400', '102400', '2014-12-10 23:23:21', '2014-12-10 23:23:21');
COMMIT;

-- ----------------------------
--  Table structure for `slc_rad_product_attr`
-- ----------------------------
DROP TABLE IF EXISTS `slc_rad_product_attr`;
CREATE TABLE `slc_rad_product_attr` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `product_id` int(11) NOT NULL,
  `attr_name` varchar(255) NOT NULL,
  `attr_value` varchar(255) NOT NULL,
  `attr_desc` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
--  Table structure for `slc_rad_roster`
-- ----------------------------
DROP TABLE IF EXISTS `slc_rad_roster`;
CREATE TABLE `slc_rad_roster` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `mac_addr` varchar(17) NOT NULL,
  `account_number` varchar(32) DEFAULT NULL,
  `begin_time` varchar(19) NOT NULL,
  `end_time` varchar(19) NOT NULL,
  `roster_type` smallint(6) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
--  Table structure for `slc_rad_ticket`
-- ----------------------------
DROP TABLE IF EXISTS `slc_rad_ticket`;
CREATE TABLE `slc_rad_ticket` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `account_number` varchar(253) NOT NULL,
  `acct_input_gigawords` int(11) DEFAULT NULL,
  `acct_input_octets` int(11) DEFAULT NULL,
  `acct_input_packets` int(11) DEFAULT NULL,
  `acct_output_gigawords` int(11) DEFAULT NULL,
  `acct_output_octets` int(11) DEFAULT NULL,
  `acct_output_packets` int(11) DEFAULT NULL,
  `acct_session_id` varchar(253) NOT NULL,
  `acct_session_time` int(11) NOT NULL,
  `acct_start_time` varchar(19) NOT NULL,
  `acct_stop_time` varchar(19) NOT NULL,
  `acct_terminate_cause` int(11) DEFAULT NULL,
  `mac_addr` varchar(128) DEFAULT NULL,
  `calling_station_id` varchar(128) DEFAULT NULL,
  `frame_id_netmask` varchar(15) DEFAULT NULL,
  `framed_ipaddr` varchar(15) DEFAULT NULL,
  `nas_class` varchar(253) DEFAULT NULL,
  `nas_addr` varchar(15) NOT NULL,
  `nas_port` int(11) DEFAULT NULL,
  `nas_port_id` varchar(253) DEFAULT NULL,
  `nas_port_type` int(11) DEFAULT NULL,
  `service_type` int(11) DEFAULT NULL,
  `session_timeout` int(11) DEFAULT NULL,
  `start_source` int(11) NOT NULL,
  `stop_source` int(11) NOT NULL,
  `acct_fee` int(11) NOT NULL,
  `fee_receivables` int(11) NOT NULL,
  `is_deduct` int(11) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

SET FOREIGN_KEY_CHECKS = 1;
