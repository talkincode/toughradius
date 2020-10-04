create schema if not exists toughradius collate utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS toughradius@'172.%.%.%' identified by 'radpwd';
GRANT ALL PRIVILEGES ON toughradius.* TO toughradius@'172.%.%.%';
ALTER USER 'toughradius'@'172.%.%.%' IDENTIFIED WITH mysql_native_password BY 'radpwd';

use toughradius;

create table if not exists tr_bras
(
	id bigint auto_increment primary key,
	identifier varchar(128) null,
	name varchar(64) not null,
	ipaddr varchar(32) null,
	vendor_id varchar(32) not null,
	portal_vendor varchar(32) not null,
	secret varchar(64) not null,
	coa_port int not null,
	ac_port int not null,
	auth_limit int null,
	acct_limit int null,
	status enum('enabled', 'disabled') null,
	remark varchar(512) null,
	create_time datetime not null
);

create index ix_tr_bras_identifier on tr_bras (identifier);

create index ix_tr_bras_ipaddr on tr_bras (ipaddr);

create table if not exists tr_config
(
	id bigint auto_increment primary key,
	type varchar(32) not null,
	name varchar(128) not null,
	value varchar(255) null,
	remark varchar(255) null
);

create table if not exists tr_subscribe
(
	id bigint auto_increment primary key,
	node_id bigint default 0 not null,
	subscriber varchar(32) null,
	realname varchar(32) null,
	password varchar(128) not null,
	domain varchar(128) null,
	addr_pool varchar(128) null,
	policy varchar(512) null,
	is_online int null,
	active_num int null,
	bind_mac tinyint(1) null,
	bind_vlan tinyint(1) null,
	ip_addr varchar(32) null,
	mac_addr varchar(32) null,
	in_vlan int null,
	out_vlan int null,
	up_rate bigint null,
	down_rate bigint null,
	up_peak_rate bigint null,
	down_peak_rate bigint null,
	up_rate_code varchar(32) null,
	down_rate_code varchar(32) null,
	status enum('enabled', 'disabled') null,
	remark varchar(512) null,
	begin_time datetime not null,
	expire_time datetime not null,
	create_time datetime not null,
	update_time datetime null
);

create index ix_tr_subscribe_create_time
	on tr_subscribe (create_time);

create index ix_tr_subscribe_expire_time
	on tr_subscribe (expire_time);

create index ix_tr_subscribe_status
	on tr_subscribe (status);

create index ix_tr_subscribe_subscriber
	on tr_subscribe (subscriber);

create index ix_tr_subscribe_update_time
	on tr_subscribe (update_time);


use toughradius;
INSERT INTO toughradius.tr_bras
(identifier, name, ipaddr, vendor_id, portal_vendor,secret, coa_port,ac_port, auth_limit, acct_limit, STATUS, remark, create_time)
VALUES ('radius-tester', 'radius-tester', '127.0.0.1', '14988',"cmccv1", 'secret', 3799,2000, 1000, 1000, NULL, '0', '2019-03-01 14:07:46');

INSERT INTO toughradius.tr_subscribe
(node_id,  subscriber, realname, password, domain, addr_pool, policy, is_online, active_num,
 bind_mac, bind_vlan, ip_addr, mac_addr, in_vlan, out_vlan, up_rate, down_rate, up_peak_rate, down_peak_rate, up_rate_code,
 down_rate_code, status, remark, begin_time, expire_time, create_time, update_time)
VALUES (0, 'test01', '', '888888',  null, null, null, null, 10, 0, 0, '', '', 0, 0, 10.000, 10.000, 100.000, 100.000,
        '10', '10', 'enabled', '', '2019-03-01 14:13:02', '2019-03-01 14:13:00', '2019-03-01 14:12:59', '2019-03-01 14:12:56');