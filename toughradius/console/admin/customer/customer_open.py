#!/usr/bin/env python
#coding=utf-8

import cyclone.auth
import cyclone.escape
import cyclone.web
import decimal
import datetime
from hashlib import md5
from toughradius.console import models
from toughradius.console.admin.customer import customer_forms
from toughradius.console.admin.customer.customer import CustomerHandler
from toughradius.common.permit import permit
from toughradius.common import utils
from toughradius.common.settings import * 

@permit.route(r"/customer/open", u"用户快速开户",MenuUser, order=1.1000, is_menu=True)
class CustomerOpenHandler(CustomerHandler):

    @cyclone.web.authenticated
    def get(self):
        nodes = [(n.id, n.node_desc) for n in self.get_opr_nodes()]
        products = [(n.id, n.product_name) for n in self.get_opr_products()]
        form = customer_forms.customer_open_form(nodes, products)
        return self.render("account_open_form.html", form=form)

    @cyclone.web.authenticated
    def post(self):
        nodes = [(n.id, n.node_desc) for n in self.get_opr_nodes()]
        products = [(n.id, n.product_name) for n in self.get_opr_products()]
        form = customer_forms.customer_open_form(nodes, products)
        if not form.validates(source=self.get_params()):
            return self.render("account_open_form.html", form=form)

        if self.db.query(models.TrAccount).filter_by(account_number=form.d.account_number).count() > 0:
            return self.render("account_open_form.html", form=form, msg=u"账号%s已经存在" % form.d.account_number)

        if form.d.ip_address and self.db.query(models.TrAccount).filter_by(ip_address=form.d.ip_address).count() > 0:
            return self.render("account_open_form.html", form=form, msg=u"ip%s已经被使用" % form.d.ip_address)

        if self.db.query(models.TrCustomer).filter_by(
            customer_name=form.d.customer_name).count() > 0:
            return self.render("account_open_form.html", form=form, msg=u"用户名%s已经存在" % form.d.customer_name)

        customer = models.TrCustomer()
        customer.node_id = form.d.node_id
        customer.realname = form.d.realname
        customer.customer_name = form.d.customer_name or form.d.account_number
        mpwd = form.d.customer_password or form.d.password
        customer.password = md5(mpwd.encode()).hexdigest()
        customer.idcard = form.d.idcard
        customer.sex = '1'
        customer.age = '0'
        customer.email = ''
        customer.mobile = form.d.mobile
        customer.address = form.d.address
        customer.create_time = utils.get_currtime()
        customer.update_time = utils.get_currtime()
        customer.email_active = 0
        customer.mobile_active = 0
        customer.active_code = utils.get_uuid()
        customer.customer_desc = form.d.customer_desc
        self.db.add(customer)
        self.db.flush()
        self.db.refresh(customer)

        accept_log = models.TrAcceptLog()
        accept_log.accept_type = 'open'
        accept_log.accept_source = 'console'
        accept_log.account_number = form.d.account_number
        accept_log.accept_time = customer.create_time
        accept_log.operator_name = self.current_user.username
        accept_log.accept_desc = u"用户新开户：(%s)%s" % (customer.customer_name, customer.realname)
        self.db.add(accept_log)
        self.db.flush()
        self.db.refresh(accept_log)

        order_fee = 0
        balance = 0
        expire_date = form.d.expire_date
        product = self.db.query(models.TrProduct).get(form.d.product_id)

        # 预付费包月
        if product.product_policy == BOMonth:
            order_fee = decimal.Decimal(product.fee_price) * decimal.Decimal(form.d.months)
            order_fee = int(order_fee.to_integral_value())

        # 买断包月,买断流量
        elif product.product_policy in (BOMonth, BOFlows):
            order_fee = int(product.fee_price)

        # 预付费时长,预付费流量
        elif product.product_policy in (PPTimes, PPFlow):
            balance = utils.yuan2fen(form.d.fee_value)
            expire_date = MAX_EXPIRE_DATE

        order = models.TrCustomerOrder()
        order.order_id = utils.gen_order_id()
        order.customer_id = customer.customer_id
        order.product_id = product.id
        order.account_number = form.d.account_number
        order.order_fee = order_fee
        order.actual_fee = utils.yuan2fen(form.d.fee_value)
        order.pay_status = 1
        order.accept_id = accept_log.id
        order.order_source = 'console'
        order.create_time = customer.create_time
        order.order_desc = u"用户新开账号"
        self.db.add(order)

        account = models.TrAccount()
        account.account_number = form.d.account_number
        account.ip_address = form.d.ip_address
        account.customer_id = customer.customer_id
        account.product_id = order.product_id
        account.install_address = customer.address
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
        account.vlan_id = 0
        account.vlan_id2 = 0
        account.create_time = customer.create_time
        account.update_time = customer.create_time
        account.account_desc = customer.customer_desc
        self.db.add(account)

        self.add_oplog(u"用户新开账号 %s" % account.account_number)

        self.db.commit()
        self.redirect(self.detail_url_fmt(account.account_number))


