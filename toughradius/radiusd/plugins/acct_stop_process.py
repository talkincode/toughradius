#!/usr/bin/env python
#coding=utf-8
from twisted.python import log
from toughradius.radiusd.pyrad import packet
from toughradius.radiusd.settings import *
from toughradius.radiusd import utils
import logging
import datetime

def process(req=None,user=None,radiusd=None,**kwargs):
    if not req.get_acct_status_type() == STATUS_TYPE_STOP:
        return  
    
    runstat=radiusd.runstat
    store = radiusd.store
    
    runstat.acct_stop += 1   
    ticket = req.get_ticket()
    if not ticket.nas_addr:
        ticket.nas_addr = req.source[0]

    _datetime = datetime.datetime.now() 
    online = store.get_online(ticket.nas_addr,ticket.acct_session_id)    
    if not online:
        session_time = ticket.acct_session_time 
        stop_time = _datetime.strftime( "%Y-%m-%d %H:%M:%S")
        start_time = (_datetime - datetime.timedelta(seconds=int(session_time))).strftime( "%Y-%m-%d %H:%M:%S")
        ticket.acct_start_time = start_time
        ticket.acct_stop_time = stop_time
        ticket.start_source= STATUS_TYPE_STOP
        ticket.stop_source = STATUS_TYPE_STOP
        store.add_ticket(ticket)
    else:
        store.del_online(ticket.nas_addr,ticket.acct_session_id)
        ticket.acct_start_time = online['acct_start_time']
        ticket.acct_stop_time= _datetime.strftime( "%Y-%m-%d %H:%M:%S")
        ticket.start_source = online['start_source']
        ticket.stop_source = STATUS_TYPE_STOP
        store.add_ticket(ticket)

    log.msg('%s Accounting stop request, remove online'%req.get_user_name(),level=logging.INFO)



        



        