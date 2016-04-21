#!/usr/bin/env python
#coding=utf-8

import traceback
import datetime
import random
from toughlib import apiutils, dispatch
from toughlib import db_cache as cache
from toughlib.permit import permit
from toughradius.manage.api.apibase import ApiHandler
from toughradius.manage import models
from toughradius.manage.settings import * 

""" 客户上网账号自动生成
"""
@permit.route(r"/api/v1/account/gen")
class APIBuildAccountHandler(ApiHandler):

    _base_id = 0

    def next_account_number(self):
        if APIBuildAccountHandler._base_id >= 99999:
            APIBuildAccountHandler._base_id=0
        APIBuildAccountHandler._base_id += 1
        next_num = str(APIBuildAccountHandler._base_id).zfill(5)
        year = datetime.datetime.now().year
        account_number = "{0}{1}".format(year,next_num)
        if self.db.query(models.TrAccount).get(account_number):
            return self.next_account_number()
        else:
            return account_number

    def get(self):
        self.post()

    def post(self):
        try:
            request = self.parse_form_request()
        except apiutils.SignError, err:
            return self.render_sign_err(err)
        except Exception as err:
            return self.render_parse_err(err)

        try:
            self.render_success(account=str(self.next_account_number()))
        except Exception as err:
            self.render_unknow(err)






