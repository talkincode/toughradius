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
from hashlib import md5

""" 客户登陆校验，支持自助服务名登陆，支持上网账号登陆
"""

@permit.route(r"/api/customer/auth")
class CustomerAuthHandler(ApiHandler):
    """ @param: 
        account_number: str, 
        customer_name: str,
        password: str,
    """

    def get(self):
        self.post()

    def post(self):
        try:
            request = self.parse_form_request()
            account_number = request.get('account_number')
            customer_name = request.get('customer_name')
            password = request.get('password')

            if not any([account_number, customer_name]):
                raise Exception("account_number, customer_name must one")
            if not password:
                raise Exception("password is empty")

            customer, account = None,None
            if customer_name:
                customer = self.db.query(models.TrCustomer).filter_by(customer_name=customer_name).first()
            if account_number:
                account = self.db.query(models.TrAccount).filter_by(account_number=account_number).first()

            if not any([customer,account]):
                raise Exception('auth failure,customer or account not exists')

            if customer and md5(password.encode()).hexdigest() == customer.password:
                return self.render_result(code=0, msg='success')

            if account and password == self.aes.decrypt(account.password):
                return self.render_result(code=0, msg='success')

            raise Exception('auth failure, password not match')

        except Exception as err:
            self.render_result(code=1, msg=utils.safeunicode(err.message))
            import traceback
            traceback.print_exc()
            return















