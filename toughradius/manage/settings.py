#!/usr/bin/env python
#coding=utf-8

import decimal,os
from toughlib.storage import Storage

decimal.getcontext().prec = 11
decimal.getcontext().rounding = decimal.ROUND_UP

# ToughCloud 定义

TOUGHCLOUD = Storage(
    apiurl = "http://www.toughcloud.net/api/v1"
)


# 资费类型定义 预付费包月，预付费时长，买断包月，买断时长，预付费流量，买断流量，自由资费
FEES = (PPMonth, PPTimes, BOMonth, BOTimes, PPFlow, BOFlows, FreeFee) = (0, 1, 2, 3, 4, 5, 6)

# 账号状态 预授权，正常，停机，销户，过期
ACCOUNT_STATUS = (UsrPreAuth, UsrNormal, UsrPause, UsrCancel, UsrExpire) = (0, 1, 2, 3, 4)

# 自由资费类型 自由日期，自由流量，自由时长
FREE_FEE_BILLS = (FreeFeeDate,FreeFeeFlow,FreeTimeLen) = (0,1,2)
#自由资费ID
FREE_FEE_PID = 1

# Radius记账类型
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

# Radius 记账类型
STATUS_TYPE_START   = 1
STATUS_TYPE_STOP    = 2
STATUS_TYPE_UPDATE  = 3
STATUS_TYPE_UNLOCK = 4
STATUS_TYPE_CHECK_ONLINE = 5
STATUS_TYPE_ACCT_ON  = 7
STATUS_TYPE_ACCT_OFF = 8


# 业务受理类型
ACCEPT_TYPES = {
    'open'  : u'开户',
    'pause' : u'停机',
    'resume': u'复机',
    'cancel': u'销户',
    'next'  : u'续费',
    'charge': u'充值',
    'change': u'变更'
}

# 管理菜单定义
ADMIN_MENUS = (MenuSys, MenuRes, MenuUser, MenuOpt, MenuStat) = (
    u"系统管理", u"资源管理", u"用户管理", u"维护管理", u"统计分析")

MENU_ICONS = {
    u"系统管理": "fa fa-cog",
    u"资源管理": "fa fa-desktop",
    u"用户管理": "fa fa-users",
    u"维护管理": "fa fa-wrench",
    u"统计分析": "fa fa-bar-chart"
}

# 用户最大过期时间定义
MAX_EXPIRE_DATE = '3000-12-30'

# 系统缓存名称定义
param_cache_key = 'toughradius.cache.param.{0}'.format
account_cache_key = 'toughradius.cache.account.{0}'.format
account_attr_cache_key = 'toughradius.cache.account.attr.{0}.{1}'.format
product_cache_key = 'toughradius.cache.product.{0}'.format
product_attrs_cache_key = 'toughradius.cache.product.attrs.{0}'.format
bas_cache_key = 'toughradius.cache.bas.{0}'.format
radius_statcache_key = 'toughradius.cache.radius.stat'

# exit_signal defined
signal_worker_exit = "signal_tr_rad_worker_exit"
signal_master_exit = "signal_tr_rad_master_exit"
signal_manage_exit = "signal_tr_manage_exit"
signal_task_exit = "signal_tr_task_exit"

signal_all_exit = (signal_worker_exit,signal_master_exit,signal_manage_exit,signal_task_exit)

# redis配置定义
def redis_conf(config):
    eredis_url = os.environ.get("REDIS_URL")
    eredis_port = os.environ.get("REDIS_PORT")
    eredis_pwd = os.environ.get("REDIS_PWD")
    eredis_db = os.environ.get("REDIS_DB")

    is_update = any([eredis_url,eredis_port,eredis_pwd,eredis_db])

    if eredis_url:
        config['redis']['host'] = eredis_url
    if eredis_port:
        config['redis']['port'] = int(eredis_port)
    if eredis_pwd:
        config['redis']['passwd'] = eredis_pwd 
    if eredis_db:
        config['redis']['db'] = int(eredis_db)
    if is_update:
        config.save()

    return config['redis']






