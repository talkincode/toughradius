#!/usr/bin/env python
#coding=utf-8

import cyclone.auth
import cyclone.escape
import cyclone.web
import decimal
from toughradius.manage import models
from toughradius.manage.base import BaseHandler
from toughradius.manage.customer import account, account_forms
from toughlib.permit import permit
from toughlib import utils, dispatch
from toughradius.manage.settings import * 
from toughradius.manage.events.settings import ACCOUNT_NEXT_EVENT

@permit.route(r"/admin/account/next", u"用户续费",MenuUser, order=2.3000)
class AccountNextHandler(account.AccountHandler):

    @cyclone.web.authenticated
    def get(self):
        account_number = self.get_argument("account_number")
        user = self.query_account(account_number)
        form = account_forms.account_next_form()
        form.account_number.set_value(account_number)
        form.old_expire.set_value(user.expire_date)
        form.product_id.set_value(user.product_id)
        return self.render("account_next_form.html", user=user, form=form)

    def post(self):
        account_number = self.get_argument("account_number")
        account = self.db.query(models.TrAccount).get(account_number)
        user = self.query_account(account_number)
        form = account_forms.account_next_form()
        form.product_id.set_value(user.product_id)
        if account.status not in (1, 4):
            return render("account_next_form", user=user, form=form, msg=u"无效用户状态")
        if not form.validates(source=self.get_params()):
            return render("account_next_form", user=user, form=form)

        accept_log = models.TrAcceptLog()
        accept_log.accept_type = 'next'
        accept_log.accept_source = 'console'
        accept_log.accept_desc = u"用户续费：上网账号:%s，续费%s元;%s" % (account_number, form.d.fee_value,form.d.operate_desc or '')
        accept_log.account_number = form.d.account_number
        accept_log.accept_time = utils.get_currtime()
        accept_log.operator_name = self.current_user.username
        self.db.add(accept_log)
        self.db.flush()
        self.db.refresh(accept_log)

        

        order_fee = 0
        product = self.db.query(models.TrProduct).get(user.product_id)

        # 预付费包月
        if product.product_policy == PPMonth:
            order_fee = decimal.Decimal(product.fee_price) * decimal.Decimal(form.d.months)
            order_fee = int(order_fee.to_integral_value())

        # 买断包月,买断流量,买断时长
        elif product.product_policy in (BOMonth, BOTimes, BOFlows):
            order_fee = int(product.fee_price)

        order = models.TrCustomerOrder()
        order.order_id = utils.gen_order_id()
        order.customer_id = user.customer_id
        order.product_id = user.product_id
        order.account_number = form.d.account_number
        order.order_fee = order_fee
        order.actual_fee = utils.yuan2fen(form.d.fee_value)
        order.pay_status = 1
        order.accept_id = accept_log.id
        order.order_source = 'console'
        order.create_time = utils.get_currtime()

        old_expire_date = account.expire_date

        account.status = 1
        account.expire_date = form.d.expire_date
        if product.product_policy == BOTimes:
            account.time_length += product.fee_times
        elif product.product_policy == BOFlows:
            account.flow_length += product.fee_flows

        order.order_desc = u"用户续费,续费前到期:%s,续费后到期:%s, 赠送天数: %s" % (
            old_expire_date, account.expire_date, form.d.giftdays)
        self.db.add(order)
        self.add_oplog(order.order_desc)

        self.db.commit()

        dispatch.pub(ACCOUNT_NEXT_EVENT, order.account_number, async=True)

        self.redirect(self.detail_url_fmt(account_number))


