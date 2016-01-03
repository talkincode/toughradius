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

@permit.route(r"/admin/account/open", u"用户开户",MenuUser, order=2.0000)
class AccountOpentHandler(account.AccountHandler):

    @cyclone.web.authenticated
    def get(self):
        customer_id = self.get_argument('customer_id')
        customer = self.db.query(models.TrCustomer).get(customer_id)
        products = [(n.id, n.product_name) for n in self.get_opr_products()]
        form = account_forms.account_open_form(products)
        form.customer_id.set_value(customer_id)
        form.realname.set_value(customer.realname)
        form.node_id.set_value(customer.node_id)
        return self.render("account_open_form.html", form=form)

    @cyclone.web.authenticated
    def post(self):
        products = [(n.id, n.product_name) for n in self.get_opr_products()]
        form = account_forms.account_open_form(products)
        if not form.validates(source=self.get_params()):
            return self.render("account_open_form.html", form=form)

        if self.db.query(models.TrAccount).filter_by(
            account_number=form.d.account_number).count() > 0:
            return self.render("account_open_form.html", form=form, msg=u"上网账号已经存在")

        if form.d.ip_address and self.db.query(models.TrAccount).filter_by(ip_address=form.d.ip_address).count() > 0:
            return self.render("account_open_form.html", form=form, msg=u"ip%s已经被使用" % form.d.ip_address)

        accept_log = models.TrAcceptLog()
        accept_log.accept_type = 'open'
        accept_log.accept_source = 'console'
        accept_log.account_number = form.d.account_number
        accept_log.accept_time = utils.get_currtime()
        accept_log.operator_name = self.current_user.username
        accept_log.accept_desc = u"用户新增账号：上网账号:%s" % (form.d.account_number)
        self.db.add(accept_log)
        self.db.flush()
        self.db.refresh(accept_log)

        _datetime = utils.get_currtime()
        order_fee = 0
        balance = 0
        expire_date = form.d.expire_date
        product = self.db.query(models.TrProduct).get(form.d.product_id)

        # 预付费包月
        if product.product_policy == PPMonth:
            order_fee = decimal.Decimal(product.fee_price) * decimal.Decimal(form.d.months)
            order_fee = int(order_fee.to_integral_value())
        # 买断包月,买断时长,买断流量
        elif product.product_policy in (BOMonth, BOTimes, BOFlows):
            order_fee = int(product.fee_price)
        # 预付费时长,预付费流量
        elif product.product_policy in (PPTimes, PPFlow):
            balance = utils.yuan2fen(form.d.fee_value)
            expire_date = MAX_EXPIRE_DATE

        order = models.TrCustomerOrder()
        order.order_id = utils.gen_order_id()
        order.customer_id = form.d.customer_id
        order.product_id = product.id
        order.account_number = form.d.account_number
        order.order_fee = order_fee
        order.actual_fee = utils.yuan2fen(form.d.fee_value)
        order.pay_status = 1
        order.accept_id = accept_log.id
        order.order_source = 'console'
        order.create_time = _datetime
        order.order_desc = u"用户增开账号"
        self.db.add(order)

        account = models.TrAccount()
        account.account_number = form.d.account_number
        account.ip_address = form.d.ip_address
        account.customer_id = int(form.d.customer_id)
        account.product_id = order.product_id
        account.install_address = form.d.address
        account.mac_addr = ''
        account.password = self.aes.encrypt(form.d.password)
        account.status = form.d.status
        account.balance = balance
        account.time_length = int(product.fee_times)
        account.flow_length = int(product.fee_flows)
        account.expire_date = expire_date
        account.user_concur_number = product.concur_number
        account.bind_mac = product.bind_mac
        account.bind_vlan = product.bind_vlan
        account.vlan_id1 = 0
        account.vlan_id2 = 0
        account.create_time = _datetime
        account.update_time = _datetime
        account.account_desc = form.d.account_desc
        self.db.add(account)
        self.add_oplog(u"用户增开子账号 %s" % account.account_number)
        self.db.commit()
        self.redirect(self.detail_url_fmt(account.account_number))

