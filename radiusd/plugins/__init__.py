#!/usr/bin/env python
#coding=utf-8

__all__ = [
    'auth_bind_filter',
    'auth_group_filter',
    'auth_policy_filter',
    'auth_user_filter',
    'mac_parse',
    'vlan_parse',
    'acct_start_process',
    'acct_stop_process',
    'acct_update_process',
    'acct_onoff_process',
    'auth_std_accept',
    'admin_trace_global',
    'admin_trace_user',
    'admin_unlock_online',
    'admin_update_cache',
    'admin_stat_query',
    'admin_coa_request'
]

def error_auth(resp,errmsg):
    resp.code = 3 #packet.AccessReject
    resp['Reply-Message'] = errmsg
    return resp    

