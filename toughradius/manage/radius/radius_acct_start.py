#!/usr/bin/env python
# coding=utf-8

from toughradius.manage.radius.radius_basic import  RadiusBasic
from toughlib.storage import Storage
from toughradius.manage import models
from toughlib import  utils
from toughradius.manage.settings import *

class RadiusAcctStart(RadiusBasic):

    def __init__(self, app, request):
        RadiusBasic.__init__(self, app, request)

    def acctounting(self):
        if self.is_online(self.request['nasaddr'],self.request['session_id']):
            return self.log.error('online %s is exists' % self.request['session_id'])

        if not self.account:
            return self.log.error('user %s not exists' % self.request['username'])

        online = models.TrOnline()
        online.account_number = self.request['username'],
        online.nas_addr = self.request['nasaddr'],
        online.acct_session_id = self.request['session_id'],
        online.acct_start_time = datetime.datetime.now().strftime( "%Y-%m-%d %H:%M:%S"),
        online.framed_ipaddr = self.request['ipaddr'],
        online.mac_addr = self.request['macaddr'],
        online.nas_port_id = self.request['nas_port_id'],
        online.billing_times = 0,
        online.input_total = 0,
        online.output_total = 0,
        online.start_source = STATUS_TYPE_START
        self.db.add(online)
        self.db.commit()
        self.log.info('%s Accounting start request, add new online'%online.account_number)



