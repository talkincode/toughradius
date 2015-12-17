#!/usr/bin/env python
#coding=utf-8

import cyclone.auth
import cyclone.escape
import cyclone.web
import decimal
from toughradius.console import models
from toughradius.console.admin.base import BaseHandler
from toughradius.console.admin.customer import account, account_forms
from toughradius.common.permit import permit
from toughradius.common import utils
from toughradius.common.settings import * 

@permit.route(r"/account/delete", u"用户账号删除",MenuUser, order=2.6000)
class AccountDeleteHandler(account.AccountHandler):

    @cyclone.web.authenticated
    def get(self):
        pass