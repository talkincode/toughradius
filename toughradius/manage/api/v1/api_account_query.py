#!/usr/bin/env python
#coding=utf-8
import time
import traceback
from toughlib.btforms import dataform
from toughlib.btforms import rules
from toughlib import utils, apiutils, dispatch
from toughlib.permit import permit
from toughradius.manage.api.apibase import ApiHandler
from toughradius.manage import models

""" 客户账号查询，
"""


@permit.route(r"/api/v1/account/query")
class CustomerAccountsHandler(ApiHandler):
    """ @param: 
        customer_name: str,
    """

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
            account_number = request.get('account_number')

            if not account_number:
                return self.render_verify_err(msg="account_number is required")

            account = self.db.query(models.TrAccount).get(account_number)
            account_data = {}
            if account:
                account_data ={ c.name : getattr(account, c.name) \
                        for c in account.__table__.columns if c.name not in 'password'}
            self.render_success(account=account_data)
        except Exception as err:
            self.render_unknow(err)












