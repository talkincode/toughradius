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
from toughlib import utils
from toughradius.manage.settings import * 

@permit.route(r"/admin/account/cancel", u"用户销户",MenuUser, order=2.7000)
class AccountCanceltHandler(account.AccountHandler):

    @cyclone.web.authenticated
    def get(self):
        account_number = self.get_argument("account_number")
        user = self.query_account(account_number)
        form = account_forms.account_cancel_form()
        form.account_number.set_value(account_number)
        return self.render("account_form.html", user=user, form=form)

    def post(self):
        account_number = self.get_argument("account_number")
        account = self.db.query(models.TrAccount).get(account_number)
        user = self.query_account(account_number)
        form = account_forms.account_cancel_form()
        if account.status != 1:
            return self.render("account_form.html", user=user, form=form, msg=u"无效用户状态")
        if not form.validates(source=self.get_params()):
            return self.render("account_form.html", user=user, form=form)

        accept_log = models.TrAcceptLog()
        accept_log.accept_type = 'cancel'
        accept_log.accept_source = 'console'
        accept_log.account_number = form.d.account_number
        accept_log.accept_time = utils.get_currtime()
        accept_log.operator_name = self.current_user.username
        accept_log.accept_desc = u"用户销户退费%s(元);%s" % (
            form.d.fee_value, utils.safeunicode(form.d.operate_desc))
        self.db.add(accept_log)
        self.db.flush()
        self.db.refresh(accept_log)

        old_expire_date = account.expire_date

        order = models.TrCustomerOrder()
        order.order_id = utils.gen_order_id()
        order.customer_id = user.customer_id
        order.product_id = user.product_id
        order.account_number = form.d.account_number
        order.order_fee = 0
        order.actual_fee = -utils.yuan2fen(form.d.fee_value)
        order.pay_status = 1
        order.order_source = 'console'
        order.accept_id = accept_log.id
        order.create_time = utils.get_currtime()
        order.order_desc = accept_log.accept_desc
        self.db.add(order)

        account.status = 3

        self.db.commit()

        onlines = self.db.query(models.TrOnline).filter_by(account_number=account_number)
        for _online in onlines:
            pass

        self.redirect(self.detail_url_fmt(account_number))


