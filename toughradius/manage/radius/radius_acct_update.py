#!/usr/bin/env python
# coding=utf-8

from toughlib.storage import Storage
from toughradius.manage import models
from toughlib import  utils
from toughradius.manage.settings import *
from toughradius.manage.radius.radius_billing import RadiusBilling

class RadiusAcctUpdate(RadiusBilling):

    def __init__(self, app, request):
        RadiusBilling.__init__(self, app, request)

    def acctounting(self):
        if not self.account:
            return self.log.error(
                "[Acct] Received an accounting update request but user[%s] not exists"% self.request['username'])     

        ticket = self.ticket
        online = self.get_online(ticket.nas_addr,ticket.acct_session_id)     
        if not online:         
            sessiontime = ticket.acct_session_time
            updatetime = datetime.datetime.now()
            _starttime = updatetime - datetime.timedelta(seconds=sessiontime)       
            online = models.TrOnline()
            online.account_number = self.account.account_number,
            online.nas_addr = ticket.nas_addr,
            online.acct_session_id = ticket.acct_session_id,
            online.acct_start_time = _starttime.strftime( "%Y-%m-%d %H:%M:%S"),
            online.framed_ipaddr = ticket.framed_ipaddr,
            online.mac_addr = ticket.mac_addr,
            online.nas_port_id = ticket.nas_port_id,
            online.billing_times = ticket.acct_session_time,
            online.input_total = self.get_input_total(),
            online.output_total = self.get_output_total(),
            online.start_source = STATUS_TYPE_UPDATE
            self.db.add(online)   
            self.db.commit()

        self.billing()
        self.log.info('%s Accounting update request, update online'% self.account.account_number)


        
        