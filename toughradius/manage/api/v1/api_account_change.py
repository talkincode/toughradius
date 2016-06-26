#!/usr/bin/env python
#coding=utf-8
import cyclone.auth
import cyclone.escape
import cyclone.web
import decimal
from toughradius.manage import models
from toughradius.manage.customer import account, account_forms
from toughradius.manage.api.apibase import ApiHandler
from toughlib.permit import permit
from toughlib import utils,dispatch
from toughlib import redis_cache
from toughradius.manage.settings import *
from toughradius.manage.events import settings

@permit.route(r"/api/v1/account/change")
class AccountChangeHandler(ApiHandler):

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

    def get(self):
        self.post()

    def post(self):
        products = [(n.id, n.product_name) for n in self.get_opr_products()]
        form = account_forms.account_change_form(products=products)
        try:
            request = self.parse_form_request()
            if not form.validates(**request):
                raise Exception(form.errors)
        except apiutils.SignError, err:
            return self.render_sign_err(err)
        except Exception as err:
            return self.render_parse_err(err)
        try:
            account_number = request.get('account_number')
            if not account_number:
                return self.render_verify_err(msg="account_number must")
            account = self.db.query(models.TrAccount).get(account_number)
            user = self.query_account(account_number)
            if account.status not in (1, 4):
                return self.render_verify_err(msg=u"无效用户状态")
            if not form.validates(**request):
                raise Exception(form.errors)

            product = self.db.query(models.TrProduct).get(form.d.product_id)

            accept_log = models.TrAcceptLog()
            accept_log.accept_type = 'change'
            accept_log.accept_source = 'api'
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
            self.render_success()
        except Exception, err:
            return self.render_unknow(err)

