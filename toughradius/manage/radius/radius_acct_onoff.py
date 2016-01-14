#!/usr/bin/env python
# coding=utf-8

from toughradius.manage.radius.radius_basic import  RadiusBasic
from toughlib.storage import Storage
from toughradius.manage import models
from toughlib import  utils
from toughradius.manage.settings import *

class RadiusAcctOnoff(RadiusBasic):

    def __init__(self, app, request):
        RadiusBasic.__init__(self, app, request)

    def acctounting(self):
        if not self.account:
            return self.log.error(
                "[Acct] Received an accounting onoff request but user[%s] not exists"% self.request.account_number)     

        self.unlock_online(self.request.account_number,None)
        self.log.info('bas accounting onoff success')


        
        