#!/usr/bin/env python
#coding=utf-8

import cyclone.auth
import cyclone.escape
import cyclone.web
import decimal
import datetime
from toughradius.console import models
from toughradius.console.admin.customer import customer_forms
from toughradius.console.admin.customer.customer import CustomerHandler
from toughradius.common.permit import permit
from toughradius.common import utils
from toughradius.common.settings import * 


@permit.route(r"/customer/delete", u"用户删除",MenuUser, order=1.5000)
class CustomerDeleteHandler(CustomerHandler):

    @cyclone.web.authenticated
    def get(self):
        pass