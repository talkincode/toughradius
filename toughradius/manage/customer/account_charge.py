#!/usr/bin/env python
#coding=utf-8

import cyclone.auth
import cyclone.escape
import cyclone.web
import decimal
from toughradius import models
from toughradius.manage.base import BaseHandler
from toughradius.manage.customer import account, account_forms
from toughlib.permit import permit
from toughlib import utils,dispatch
from toughlib import redis_cache
from toughradius.manage.settings import * 
from toughradius.events import settings

@permit.route(r"/admin/account/charge", u"用户充值",MenuUser, order=2.4000)
class AccountChargeHandler(account.AccountHandler):

    @cyclone.web.authenticated
    def get(self):
        account_number = self.get_argument("account_number")
        user = self.query_account(account_number)
        form = account_forms.account_charge_form()
        form.account_number.set_value(account_number)
        return self.render("account_form.html", user=user, form=form)

    @cyclone.web.authenticated
    def post(self):
        account_number = self.get_argument("account_number")
        account = self.db.query(models.TrAccount).get(account_number)
        user = self.query_account(account_number)
        form = account_forms.account_charge_form()

        if account.status not in (1, 4):
            return render("account_form", user=user, form=form, msg=u"无效用户状态")
        if not form.validates(source=self.get_params()):
            return render("account_form", user=user, form=form)

        accept_log = models.TrAcceptLog()
        accept_log.accept_type = 'charge'
        accept_log.accept_source = 'console'
        _new_fee = account.balance + utils.yuan2fen(form.d.fee_value)
        accept_log.accept_desc = u"用户充值：充值前%s元,充值后%s元;%s" % (
            utils.fen2yuan(account.balance),
            utils.fen2yuan(_new_fee),
            (form.d.operate_desc or '')
        )        
        accept_log.account_number = form.d.account_number
        accept_log.accept_time = utils.get_currtime()
        accept_log.operator_name = self.current_user.username
        self.db.add(accept_log)
        self.db.flush()
        self.db.refresh(accept_log)

        order = models.TrCustomerOrder()
        order.order_id = utils.gen_order_id()
        order.customer_id = user.customer_id
        order.product_id = user.product_id
        order.account_number = form.d.account_number
        order.order_fee = utils.yuan2fen(form.d.fee_value)
        order.actual_fee = utils.yuan2fen(form.d.fee_value)
        order.pay_status = 1
        order.accept_id = accept_log.id
        order.order_source = 'console'
        order.create_time = utils.get_currtime()
        order.order_desc = accept_log.accept_desc

        self.db.add(order)
        self.add_oplog(order.order_desc)

        account.balance += order.actual_fee
        self.db.commit()
        self.redirect(self.detail_url_fmt(account_number))






