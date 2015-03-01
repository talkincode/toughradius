#!/usr/bin/env python
#coding=utf-8
from twisted.python import log
from toughradius.radiusd.pyrad import packet
from toughradius.radiusd.settings import *
from toughradius.radiusd import utils
import logging
import datetime

def process(req=None,user=None,radiusd=None,**kwargs):
    if not req.get_acct_status_type() == STATUS_TYPE_START:
        return
    
    runstat=radiusd.runstat
    store = radiusd.store
    
    if store.is_online(req.get_nas_addr(),req.get_acct_sessionid()):
        runstat.acct_drop += 1
        return log.err('online %s is exists'%req.get_acct_sessionid())

    if not user:
        runstat.acct_drop += 1
        return log.err('user %s not exists'%req.get_user_name())

    runstat.acct_start += 1    
    online = utils.Storage(
        account_number = user['account_number'],
        nas_addr = req.get_nas_addr(),
        acct_session_id = req.get_acct_sessionid(),
        acct_start_time = datetime.datetime.now().strftime( "%Y-%m-%d %H:%M:%S"),
        framed_ipaddr = req.get_framed_ipaddr(),
        mac_addr = req.get_mac_addr(),
        nas_port_id = req.get_nas_portid(),
        billing_times = 0,
        input_total = 0,
        output_total = 0,
        start_source = STATUS_TYPE_START
    )

    store.add_online(online)

    log.msg('%s Accounting start request, add new online'%req.get_user_name(),level=logging.INFO)
