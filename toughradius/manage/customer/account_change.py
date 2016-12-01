#!/usr/bin/env python
#coding=utf-8

import cyclone.auth
import cyclone.escape
import cyclone.web
import decimal
from toughradius import models
from toughradius.manage.base import BaseHandler
from toughradius.manage.customer import account, account_forms
from toughradius.common.permit import permit
from toughradius.common import utils,dispatch
from toughradius.common import redis_cache
from toughradius.manage.settings import * 
from toughradius.events import settings

@permit.route(r"/admin/account/change/get_policy")
class AccountChangePolicyGetHandler(account.AccountHandler):

    @cyclone.web.authenticated
    def get(self):
        product_id = self.get_argument("product_id")
        product_policy = self.db.query(models.TrProduct.product_policy).filter_by(id=product_id).scalar()
        return self.render_json(data={'id': product_id, 'policy': product_policy})

@permit.route(r"/admin/account/change", u"用户变更资费",MenuUser, order=2.5000)
class AccountChangeHandler(account.AccountHandler):

    @cyclone.web.authenticated
    def get(self):
        account_number = self.get_argument("account_number")
        products = [(n.id, n.product_name) for n in self.get_opr_products()]
        user = self.query_account(account_number)
        form = account_forms.account_change_form(products=products)
        form.expire_date.set_value(user.expire_date)
        form.account_number.set_value(account_number)
        return self.render("account_change_form.html", user=user, form=form)

    def post(self):
        account_number = self.get_argument("account_number")
        products = [(n.id, n.product_name) for n in self.get_opr_products()]
        form = account_forms.account_change_form(products=products)
        account = self.db.query(models.TrAccount).get(account_number)
        user = self.query_account(account_number)
        if account.status not in (1, 4):
            return self.render("account_change_form.html", user=user, form=form, msg=u"无效用户状态")
        if not form.validates(source=self.get_params()):
            return self.render("account_change_form.html", user=user, form=form)

        product = self.db.query(models.TrProduct).get(form.d.product_id)

        accept_log = models.TrAcceptLog()
        accept_log.accept_type = 'change'
        accept_log.accept_source = 'console'
        accept_log.account_number = form.d.account_number
        accept_log.accept_time = utils.get_currtime()
        accept_log.operator_name = self.current_user.username
        accept_log.accept_desc = u"用户资费变更为:%s;%s" % (
            product.product_name, utils.safeunicode(form.d.operate_desc))
        self.db.add(accept_log)
        self.db.flush()
        self.db.refresh(accept_log)

        old_exoire_date = account.expire_date

        account.product_id = product.id
        # (PPMonth,PPTimes,BOMonth,BOTimes,PPFlow,BOFlows)
        if product.product_policy in (PPMonth, BOMonth):
            account.expire_date = form.d.expire_date
            account.balance = 0
            account.time_length = 0
            account.flow_length = 0
        elif product.product_policy in (PPTimes, PPFlow):
            account.expire_date = MAX_EXPIRE_DATE
            account.balance = utils.yuan2fen(form.d.balance)
            account.time_length = 0
            account.flow_length = 0
        elif product.product_policy == BOTimes:
            account.expire_date = MAX_EXPIRE_DATE
            account.balance = 0
            account.time_length = utils.hour2sec(form.d.time_length)
            account.flow_length = 0
        elif product.product_policy == BOFlows:
            account.expire_date = MAX_EXPIRE_DATE
            account.balance = 0
            account.time_length = 0
            account.flow_length = utils.mb2kb(form.d.flow_length)

        order = models.TrCustomerOrder()
        order.order_id = utils.gen_order_id()
        order.customer_id = account.customer_id
        order.product_id = account.product_id
        order.account_number = account.account_number
        order.order_fee = 0
        order.actual_fee = utils.yuan2fen(form.d.add_value) - utils.yuan2fen(form.d.back_value)
        order.pay_status = 1
        order.accept_id = accept_log.id
        order.order_source = 'console'
        order.create_time = utils.get_currtime()


        order.order_desc = u"用户变更资费,变更前到期:%s,变更后到期:%s" % (
            old_exoire_date, account.expire_date)

        self.db.add(order)
        self.add_oplog(accept_log.accept_desc)
        self.db.commit()
        dispatch.pub(settings.CACHE_DELETE_EVENT,account_cache_key(account.account_number), async=True)
        self.redirect(self.detail_url_fmt(account_number))







