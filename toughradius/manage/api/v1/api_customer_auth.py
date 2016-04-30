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

@permit.route(r"/api/v1/customer/auth")
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
        except apiutils.SignError, err:
            return self.render_sign_err(err)
        except Exception as err:
            return self.render_parse_err(err)

        try:
            account_number = request.get('account_number')
            customer_name = request.get('customer_name')
            password = request.get('password')

            if not any([account_number, customer_name]):
                return self.render_verify_err(msg="account_number, customer_name must one")
            if not password:
                return self.render_verify_err(msg="password is empty")

            customer, account = None,None
            if customer_name:
                customer = self.db.query(models.TrCustomer).filter_by(customer_name=customer_name).first()
            if account_number:
                account = self.db.query(models.TrAccount).filter_by(account_number=account_number).first()

            if not any([customer,account]):
                return self.render_verify_err(msg='auth failure,customer or account not exists')

            if customer and md5(password.encode()).hexdigest() == customer.password:
                return self.render_success(customer_name=customer.customer_name)

            if account and password == self.aes.decrypt(account.password):
                customer = self.db.query(models.TrCustomer).get(account.customer_id)
                return self.render_success(customer_name=customer.customer_name)

            return self.render_verify_err(msg='auth failure, password not match')

        except Exception as err:
            self.render_unknow(err)
            import traceback
            traceback.print_exc()
         















