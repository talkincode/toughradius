#!/usr/bin/env python
# coding=utf-8
import datetime
from toughlib.storage import Storage
from toughradius import models
from toughlib import  utils, logger, dispatch
from toughradius.manage.settings import *
from toughradius.radius.radius_billing import RadiusBilling

class RadiusAcctStop(RadiusBilling):

    def __init__(self, dbengine=None,cache=None,aes=None,request=None):
        RadiusBilling.__init__(self, dbengine,cache,aes, request)

    def acctounting(self):
        if not self.account:
            return logger.error(
                "[Acct] Received an accounting stop request but user[%s] not exists"% self.request.account_number)  

        ticket = Storage(**self.request)
        _datetime = datetime.datetime.now() 
        online = self.get_online(ticket.nas_addr,ticket.acct_session_id)    
        if not online:
            session_time = ticket.acct_session_time 
            stop_time = _datetime.strftime( "%Y-%m-%d %H:%M:%S")
            start_time = (_datetime - datetime.timedelta(seconds=int(session_time))).strftime( "%Y-%m-%d %H:%M:%S")
            ticket.acct_start_time = start_time
            ticket.acct_stop_time = stop_time
            ticket.start_source= STATUS_TYPE_STOP
            ticket.stop_source = STATUS_TYPE_STOP
            self.add_ticket(ticket)
        else:
            self.del_online(ticket.nas_addr,ticket.acct_session_id)
            ticket.acct_start_time = online.acct_start_time
            ticket.acct_stop_time= _datetime.strftime( "%Y-%m-%d %H:%M:%S")
            ticket.start_source = online.start_source
            ticket.stop_source = STATUS_TYPE_STOP
            self.add_ticket(ticket)

            self.billing(online)
            logger.info('%s Accounting stop request, remove online'% self.account.account_number)



