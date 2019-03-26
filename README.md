# TOUGHRADIUS


ToughRADIUS is a Radius server software developed based on Java & SpringBoot (since v6.x), which implements the standard Radius protocol and supports the extension of Radius protocol.

ToughRADIUS can be understood as a Radius middleware, and it does not implement all of the business functions. But it's easy to Easier to extended development.

ToughRADIUS is similar to freeRADIUS, But it's simpler to use, Easier to extended development.

## install

### System environment dependence

operating system

- Linux
- Windows
- MacOS

java version: 1.8+

database server

MySQL/MariaDB

### Database init

> First confirm that the database has started

create database

    create database toughradius DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;
    GRANT ALL ON toughradius.* TO raduser@'127.0.0.1' IDENTIFIED BY 'radpwd' WITH GRANT OPTION;FLUSH PRIVILEGES;

create tables

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

insert test data

    INSERT INTO toughradius.tr_bras
    (identifier, name, ipaddr, vendor_id, secret, coa_port, auth_limit, acct_limit, STATUS, remark, create_time)
    VALUES ('radius-tester', 'radius-tester', '127.0.0.1', '14988', 'secret', 3799, 1000, 1000, NULL, '0', '2019-03-01 14:07:46');
    
    INSERT INTO toughradius.tr_subscribe
    (node_id, area_id, subscriber, realname, password, bill_type, domain, addr_pool, policy, is_online, active_num, flow_amount,
     bind_mac, bind_vlan, ip_addr, mac_addr, in_vlan, out_vlan, up_rate, down_rate, up_peak_rate, down_peak_rate, up_rate_code,
     down_rate_code, status, remark, begin_time, expire_time, create_time, update_time)
    VALUES (0, 0, 'test01', '', '888888', 'time', null, null, null, null, 10, 0, 0, 0, '', '', 0, 0, 10.000, 10.000, 100.000, 100.000,
            '10', '10', 'enabled', '', '2019-03-01 14:13:02', '2019-03-01 14:13:00', '2019-03-01 14:12:59', '2019-03-01 14:12:56');
            
### Running the main program

    java -jar -Xms256M -Xmx1024M /opt/toughradius-latest.jar  --spring.profiles.active=prod
    
> Note the file (toughradius-latest.jar) path

### Linux systemd config

write config files( see scripts dir)

    /etc/toughradius.env
    /usr/lib/systemd/system/toughradius.service

Run the following command

    systemctl enable toughradius
    systemctl start toughradius
    
