#!/usr/bin/env python
#coding=utf-8

import cyclone.auth
import cyclone.escape
import cyclone.web
import decimal
import datetime
from toughradius.manage import models
from toughradius.manage.customer import customer_forms
from toughradius.manage.customer.customer import CustomerHandler
from toughlib.permit import permit
from toughlib import utils
from toughradius.manage.settings import * 


@permit.route(r"/admin/customer/delete", u"用户删除",MenuUser, order=1.5000)
class CustomerDeleteHandler(CustomerHandler):

    @cyclone.web.authenticated
    def get(self):
        pass