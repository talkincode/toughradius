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

""" 客户资料查询，客户账号查询,客户交易查询
"""


@permit.route(r"/api/v1/customer/query")
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
            customer_name = request.get('customer_name')
            account_number = request.get('account_number')

            if not any([customer_name,account_number]):
                return self.render_verify_err(msg="customer_name,account_number must one")

            customer = None
            if customer_name:
                customer = self.db.query(models.TrCustomer).filter_by(customer_name=customer_name).first()
            else:
                customer = self.db.query(models.TrCustomer).filter(
                    models.TrCustomer.customer_id==models.TrAccount.customer_id,
                    models.TrAccount.account_number==account_number
                ).first()

            if not customer:
                return self.render_verify_err(msg="customer is not exists")

            excludes = ['password','email_active','active_code','mobile_active']
            customer_data = { c.name : getattr(customer, c.name) \
                for c in customer.__table__.columns if c.name not in excludes}

            accounts = self.db.query(models.TrAccount).filter_by(customer_id=customer.customer_id)
            if account_number:
                accounts = accounts.filter_by(account_number=account_number)

            account_datas = []
            for account in accounts:
                account_data = { c.name : getattr(account, c.name) \
                        for c in account.__table__.columns if c.name not in 'password'}
                account_datas.append(account_data)

            self.render_success(customer=customer_data, accounts=account_datas)
        except Exception as err:
            self.render_unknow(err)












