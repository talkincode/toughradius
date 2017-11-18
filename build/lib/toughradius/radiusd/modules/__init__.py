# coding:utf-8

import request_logger
import response_logger
import request_mac_parse
import request_vlan_parse
import accept_rate_process

auth_pre = [
    request_logger,
    request_mac_parse,
    request_vlan_parse
]

acct_pre = [
    request_logger,
    request_mac_parse,
    request_vlan_parse
]

auth_post = [
    response_logger,
    accept_rate_process
]