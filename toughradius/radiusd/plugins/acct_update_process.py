#!/usr/bin/env python
#coding=utf-8
from twisted.python import log
from toughradius.radiusd.pyrad import packet
from toughradius.radiusd.settings import *
from toughradius.radiusd import utils
import logging
import datetime

def process(req=None,user=None,radiusd=None,**kwargs):
    if not req.get_acct_status_type() == STATUS_TYPE_UPDATE:
        return   

    if not user:
        return log.err("[Acct] Received an accounting update request but user[%s] not exists"%req.get_user_name())      

    runstat=radiusd.runstat
    store = radiusd.store
    
    runstat.acct_update += 1  
    online = store.get_online(req.get_nas_addr(),req.get_acct_sessionid())  

    if not online:         
        sessiontime = req.get_acct_sessiontime()
        updatetime = datetime.datetime.now()
        _starttime = updatetime - datetime.timedelta(seconds=sessiontime)       
        online = utils.Storage(
            account_number = user['account_number'],
            nas_addr = req.get_nas_addr(),
            acct_session_id = req.get_acct_sessionid(),
            acct_start_time = _starttime.strftime( "%Y-%m-%d %H:%M:%S"),
            framed_ipaddr = req.get_framed_ipaddr(),
            mac_addr = req.get_mac_addr(),
            nas_port_id = req.get_nas_portid(),
            billing_times = req.get_acct_sessiontime(),
            input_total = req.get_input_total(),
            output_total = req.get_output_total(),
            start_source = STATUS_TYPE_UPDATE
        )
        store.add_online(online)   

    log.msg('%s Accounting update request, update online'%req.get_user_name(),level=logging.INFO)
        