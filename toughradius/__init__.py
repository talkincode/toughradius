#!/usr/bin/env python
#coding=utf-8
import decimal

decimal.getcontext().prec = 11
decimal.getcontext().rounding = decimal.ROUND_UP

__version__ = 'International Edition 3.0'
__license__ = 'Apache License 2.0'


class Setting(dict):

    def __getattr__(self, key): 
        return self[key]

    def __setattr__(self, key, value): 
        self[key] = value
    

settings = Setting()

# package type define
settings.PKG_Days = 1 
settings.PKG_Times = 2 
settings.PKG_Flows = 3

settings.UsrPreAuth = 0
settings.UsrNormal = 1
settings.UsrPause = 2
settings.UsrCancel = 3
settings.UsrExpire = 4


# service accept type
settings.ACCEPT_TYPES = {
    'open'  : u'开户',
    'pause' : u'停机',
    'resume': u'复机',
    'cancel': u'销户',
    'next'  : u'续费',
    'charge': u'充值',
    'change': u'变更'
}

# manage menus define
settings.MenuSys = u"系统管理"
settings.MenuRes =  u"资源管理"
settings.MenuUser = u"用户管理",
settings.MenuStat =  u"统计分析"

settings.ADMIN_MENUS = (
    settings.MenuSys,
    settings.MenuRes,
    settings.MenuUser,
    settings.MenuStat,
)

settings.MENU_ICONS = {
    u"系统管理": "fa fa-cog",
    u"资源管理": "fa fa-desktop",
    u"用户管理": "fa fa-users",
    u"维护管理": "fa fa-wrench",
    u"统计分析": "fa fa-bar-chart"
}


settings.MAX_EXPIRE_DATE = '3000-12-30'

# cache key define

settings.PARAM_CACHE_KEY = 'toughradius.cache.param.{0}'.format
settings.ACCOUNT_CACHE_KEY = 'toughradius.cache.account.{0}'.format
settings.ACCOUNT_ATTR_CACHE_KEY = 'toughradius.cache.account.attr.{0}.{1}'.format
settings.PRODUCT_CACHE_KEY = 'toughradius.cache.product.{0}'.format
settings.PRODUCT_ATTRS_CACHE_KEY = 'toughradius.cache.product.attrs.{0}'.format
settings.BAS_CACHE_KEY = 'toughradius.cache.bas.{0}'.format
settings.RADIUS_STATCACHE_KEY = 'toughradius.cache.radius.stat'
settings.ONLINE_STATCACHE_KEY = 'toughradius.cache.online.stat'
settings.FLOW_STATCACHE_KEY = 'toughradius.cache.flow.stat'







