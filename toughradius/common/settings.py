#!/usr/bin/env python
#coding=utf-8

import decimal

decimal.getcontext().prec = 11
decimal.getcontext().rounding = decimal.ROUND_UP


FEES = (PPMonth, PPTimes, BOMonth, BOTimes, PPFlow, BOFlows) = (0, 1, 2, 3, 4, 5)

ACCOUNT_STATUS = (UsrPreAuth, UsrNormal, UsrPause, UsrCancel, UsrExpire) = (0, 1, 2, 3, 4)

CARD_STATUS = (CardInActive, CardActive, CardUsed, CardRecover) = (0, 1, 2, 3)

CARD_TYPES = (ProductCard, BalanceCard) = (0, 1)

ACCEPT_TYPES = {
    'open'  : u'开户',
    'pause' : u'停机',
    'resume': u'复机',
    'cancel': u'销户',
    'next'  : u'续费',
    'charge': u'充值',
    'change': u'变更'
}

ADMIN_MENUS = (MenuSys, MenuRes, MenuUser, MenuOpt, MenuStat) = (
    u"系统管理", u"资源管理", u"用户管理", u"维护管理", u"统计分析")

MENU_ICONS = {
    u"系统管理": "fa fa-cog",
    u"资源管理": "fa fa-desktop",
    u"用户管理": "fa fa-users",
    u"维护管理": "fa fa-wrench",
    u"统计分析": "fa fa-bar-chart"
}

MAX_EXPIRE_DATE = '3000-12-30'

# CACHE_NOTIFY

NOTIFY_NAS_UPDATE = 'nas_update'
NOTIFY_USER_UPDATE = 'user_update'

NOTIFY_NAS_UPDATE_KEY = 'nasaddr'
NOTIFY_USER_UPDATE_KEY = 'username'




