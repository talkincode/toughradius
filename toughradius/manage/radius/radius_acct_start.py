#!/usr/bin/env python
# coding=utf-8
from toughradius.manage.radius.radius_basic import  RadiusBasic
from toughlib.storage import Storage
from toughradius.manage import models
from toughlib import  utils, dispatch, logger
from toughradius.manage.settings import *
import datetime

class RadiusAcctStart(RadiusBasic):

    def __init__(self, dbengine=None,cache=None,aes=None,request=None):
        RadiusBasic.__init__(self, dbengine,cache,aes, request)

    def acctounting(self):
        if self.is_online(self.request.nas_addr,self.request.acct_session_id):
            return dispatch.pub(logger.EVENT_ERROR,'online %s is exists' % self.request.acct_session_id)

        if not self.account:
            return dispatch.pub(logger.EVENT_ERROR,'user %s not exists' % self.request.account_number)

        online = Storage(
            account_number = self.request.account_number,
            nas_addr = self.request.nas_addr,
            acct_session_id = self.request.acct_session_id,
            acct_start_time = datetime.datetime.now().strftime( "%Y-%m-%d %H:%M:%S"),
            framed_ipaddr = self.request.framed_ipaddr,
            mac_addr = self.request.mac_addr,
            nas_port_id = self.request.nas_port_id,
            billing_times = 0,
            input_total = 0,
            output_total = 0,
            start_source = STATUS_TYPE_START
        )
        self.add_online(online)
        dispatch.pub(logger.EVENT_INFO,'%s Accounting start request, add new online'%online.account_number)



