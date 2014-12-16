#!/usr/bin/env python
#coding=utf-8
from twisted.python import log
from pyrad import packet
from store import store
from settings import *
import logging
import datetime
import utils

"""记账更新包处理"""
def process(req=None,user=None,runstat=None):
    if not req.get_acct_status_type() == STATUS_TYPE_UPDATE:
        return   
    runstat.acct_update += 1  
    online = store.get_online(req.get_nas_addr(),req.get_acct_sessionid())  

    if not online:
        user = store.get_user(req.get_user_name())
        if not user:
            return log.err("[Acct] Received an accounting update request but user[%s] not exists"%req.get_user_name())
                        
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
            start_source = STATUS_TYPE_UPDATE
        )
        store.add_online(online)    

