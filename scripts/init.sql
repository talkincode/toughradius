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