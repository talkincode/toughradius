#coding:utf-8

__verson__ = '0.1'

debug = True
cache_timeout = 3600 #second
reject_delay = 9


FEE_MONTH = 0
FEE_TIMES = 1
FEE_BUYOUT = 2

STAT_AUTH_ALL = 'STAT_AUTH_ALL'
STAT_AUTH_ACCEPT = 'STAT_AUTH_ACCEPT'
STAT_AUTH_REJECT = 'STAT_AUTH_REJECT'
STAT_AUTH_DROP = 'STAT_AUTH_DROP'
STAT_ACCT_ALL = 'STAT_ACCT_ALL'
STAT_ACCT_START = 'STAT_ACCT_START'
STAT_ACCT_UPDATE = 'STAT_ACCT_UPDATE'
STAT_ACCT_STOP = 'STAT_ACCT_STOP'
STAT_ACCT_ON = 'STAT_ACCT_ON'
STAT_ACCT_OFF = 'STAT_ACCT_OFF'
STAT_ACCT_DROP = 'STAT_ACCT_DROP'
STAT_ACCT_RETRY = 'STAT_ACCT_RETRY'

STATUS_TYPE_START   = 1
STATUS_TYPE_STOP    = 2
STATUS_TYPE_UPDATE  = 3
STATUS_TYPE_UNLOCK = 4
STATUS_TYPE_ACCT_ON  = 7
STATUS_TYPE_ACCT_OFF = 8

db_config = {
    'database':{
        'maxusage': 10,
        'host': '127.0.0.1',
        'user': 'root',
        'passwd': 'root',
        'db': 'slcrms',
        'charset': "utf8",
    }
}

auth_plugins = [
    'mac_parse',
    'vlan_parse',
    'auth_roster_filter',
    'auth_user_filter',
    'auth_bind_filter',
    'auth_policy_filter',
    'auth_rate_limit',
    'auth_std_accept'
]

acct_before_plugins = [
    'mac_parse'
]

acct_plugins = [
    'acct_start_process',
    'acct_stop_process',
    'acct_update_process',
    'acct_onoff_process'
]

admin_plugins = [
    'admin_unlock_online',
]











