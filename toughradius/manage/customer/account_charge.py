#!/usr/bin/env python
#coding=utf-8

import cyclone.auth
import cyclone.escape
import cyclone.web
import decimal
from toughradius.manage import models
from toughradius.manage.base import BaseHandler
from toughradius.manage.customer import account, account_forms
from toughlib.permit import permit
from toughlib import utils
from toughradius.manage.settings import * 

@permit.route(r"/admin/account/charge", u"用户充值",MenuUser, order=2.4000)
class AccountChargeHandler(account.AccountHandler):

    @cyclone.web.authenticated
    def get(self):
        pass