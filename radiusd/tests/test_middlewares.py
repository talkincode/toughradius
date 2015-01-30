#!/usr/bin/python
#coding:utf-8
import test_config

_plugins = [
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


def test_load_plugins():
    from radiusd import middleware
    md = middleware.Middleware()
    assert md._plugin_loaded == True 
    assert len(md._plugin_modules) > 0 

    for p in _plugins:
        assert p in md._plugin_modules







