#!/usr/bin/env python
# coding=utf-8
import datetime
from toughlib.storage import Storage
from toughradius import models
from toughlib import  utils, logger, dispatch
from toughradius.manage.settings import *
from toughradius.radius.radius_billing import RadiusBilling
from toughradius.events.settings import UNLOCK_ONLINE_EVENT

class RadiusAcctUpdate(RadiusBilling):

    def __init__(self, dbengine=None,cache=None,aes=None,request=None):
        RadiusBilling.__init__(self, dbengine,cache,aes, request)

    def acctounting(self):
        if not self.account:
            dispatch.pub(UNLOCK_ONLINE_EVENT,
                self.request.account_number,self.request.nas_addr, self.request.acct_session_id,async=True)
            return logger.error(
                "[Acct] Received an accounting update request but user[%s] not exists"% self.request.account_number)     

        ticket = Storage(**self.request)
        online = self.get_online(ticket.nas_addr,ticket.acct_session_id)     
        if not online:         
            sessiontime = ticket.acct_session_time
            updatetime = datetime.datetime.now()
            _starttime = updatetime - datetime.timedelta(seconds=sessiontime)       
            online = Storage(
                account_number = self.account.account_number,
                nas_addr = ticket.nas_addr,
                acct_session_id = ticket.acct_session_id,
                acct_start_time = _starttime.strftime( "%Y-%m-%d %H:%M:%S"),
                framed_ipaddr = ticket.framed_ipaddr,
                mac_addr = ticket.mac_addr or '',
                nas_port_id = ticket.nas_port_id or '',
                billing_times = ticket.acct_session_time,
                input_total = self.get_input_total(),
                output_total = self.get_output_total(),
                start_source = STATUS_TYPE_UPDATE
            )
            self.add_online(online)

        self.billing(online)
        logger.info('%s Accounting update request, update online'% self.account.account_number)


        
        