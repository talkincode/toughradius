#!/usr/bin/env python
#coding=utf-8

import cyclone.auth
import cyclone.escape
import cyclone.web
import decimal
import datetime
from toughradius.console import models
from toughradius.console.admin.base import BaseHandler
from toughradius.console.admin.customer import account_forms
from toughradius.common.permit import permit
from toughradius.common import utils
from toughradius.common.settings import * 


class AccountHandler(BaseHandler):

    detail_url_fmt = "/customer/detail?account_number={0}".format


@permit.route(r"/account/opencalc")
class OpencalcHandler(AccountHandler):

    @cyclone.web.authenticated
    def post(self):
        months = self.get_argument('months',0)
        product_id = self.get_argument("product_id",None)
        old_expire = self.get_argument("old_expire",None)
        product = self.db.query(models.TrProduct).get(product_id)

        # 预付费时长，预付费流量，
        if product.product_policy in (PPTimes,PPFlow):
            return self.render_json(code=0,
                data=dict(policy=product.product_policy,fee_value=0,expire_date=MAX_EXPIRE_DATE))

        # 买断时长 买断流量
        elif product.product_policy in (BOTimes,BOFlows):
            fee_value = utils.fen2yuan(product.fee_price)
            return self.render_json(code=0,
                data=dict(policy=product.product_policy,fee_value=fee_value,expire_date=MAX_EXPIRE_DATE))

        # 预付费包月 
        elif product.product_policy == PPMonth:
            fee = decimal.Decimal(months) * decimal.Decimal(product.fee_price)
            fee_value = utils.fen2yuan(int(fee.to_integral_value()))
            start_expire = datetime.datetime.now()
            if old_expire:
                start_expire = datetime.datetime.strptime(old_expire,"%Y-%m-%d")
            expire_date = utils.add_months(start_expire,int(months))
            expire_date = expire_date.strftime( "%Y-%m-%d")
            return self.render_json(code=0,
                data=dict(policy=product.product_policy,fee_value=fee_value,expire_date=expire_date))

        # 买断包月
        elif product.product_policy == BOMonth:
            start_expire = datetime.datetime.now()
            if old_expire:
                start_expire = datetime.datetime.strptime(old_expire,"%Y-%m-%d")
            fee_value = utils.fen2yuan(product.fee_price)
            expire_date = utils.add_months(start_expire,product.fee_months)
            expire_date = expire_date.strftime( "%Y-%m-%d")
            return self.render_json(code=0,data=dict(policy=product.product_policy,fee_value=fee_value,expire_date=expire_date))




