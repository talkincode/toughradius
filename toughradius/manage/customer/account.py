#!/usr/bin/env python
#coding=utf-8

import cyclone.auth
import cyclone.escape
import cyclone.web
import decimal
import datetime
from toughradius.manage import models
from toughradius.manage.base import BaseHandler
from toughradius.manage.customer import account_forms
from toughlib.permit import permit
from toughlib import utils
from toughradius.manage.settings import * 


class AccountHandler(BaseHandler):

    detail_url_fmt = "/admin/customer/detail?account_number={0}".format

    def query_account(self, account_number):
        return self.db.query(
            models.TrCustomer.realname,
            models.TrAccount.customer_id,
            models.TrAccount.product_id,
            models.TrAccount.account_number,
            models.TrAccount.expire_date,
            models.TrAccount.balance,
            models.TrAccount.time_length,
            models.TrAccount.flow_length,
            models.TrAccount.user_concur_number,
            models.TrAccount.status,
            models.TrAccount.mac_addr,
            models.TrAccount.vlan_id1,
            models.TrAccount.vlan_id2,
            models.TrAccount.ip_address,
            models.TrAccount.bind_mac,
            models.TrAccount.bind_vlan,
            models.TrAccount.ip_address,
            models.TrAccount.install_address,
            models.TrAccount.create_time,
            models.TrProduct.product_name
        ).filter(
            models.TrProduct.id == models.TrAccount.product_id,
            models.TrCustomer.customer_id == models.TrAccount.customer_id,
            models.TrAccount.account_number == account_number
        ).first()


class AccountCalc:

    def calc(self, months, product_id, old_expire, giftdays):
        product = self.db.query(models.TrProduct).get(product_id)
        # 预付费时长，预付费流量，
        if product.product_policy in (PPTimes,PPFlow):
            return dict(policy=product.product_policy,fee_value=0,expire_date=MAX_EXPIRE_DATE)

        # 买断时长 买断流量
        elif product.product_policy in (BOTimes,BOFlows):
            fee_value = utils.fen2yuan(product.fee_price)
            return dict(policy=product.product_policy,fee_value=fee_value,expire_date=MAX_EXPIRE_DATE)

        # 预付费包月 
        elif product.product_policy == PPMonth:
            fee = decimal.Decimal(months) * decimal.Decimal(product.fee_price)
            fee_value = utils.fen2yuan(int(fee.to_integral_value()))
            start_expire = datetime.datetime.now()
            if old_expire:
                start_expire = datetime.datetime.strptime(old_expire,"%Y-%m-%d")
            expire_date = utils.add_months(start_expire,int(months),days=giftdays)
            expire_date = expire_date.strftime( "%Y-%m-%d")
            return dict(policy=product.product_policy,fee_value=fee_value,expire_date=expire_date)

        # 买断包月
        elif product.product_policy == BOMonth:
            start_expire = datetime.datetime.now()
            if old_expire:
                start_expire = datetime.datetime.strptime(old_expire,"%Y-%m-%d")
            fee_value = utils.fen2yuan(product.fee_price)
            expire_date = utils.add_months(start_expire,product.fee_months,days=giftdays)
            expire_date = expire_date.strftime( "%Y-%m-%d")
            return dict(policy=product.product_policy,fee_value=fee_value,expire_date=expire_date)


@permit.route(r"/admin/account/opencalc")
class OpencalcHandler(AccountHandler,AccountCalc):

    @cyclone.web.authenticated
    def post(self):
        months = int(self.get_argument('months',0))
        product_id = self.get_argument("product_id",None)
        old_expire = self.get_argument("old_expire",None)
        giftdays = int(self.get_argument('giftdays',0))
        return self.render_json(code=0,data=self.calc(months, product_id, old_expire, giftdays))





