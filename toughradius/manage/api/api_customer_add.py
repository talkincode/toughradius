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

""" 客户新开户
"""

customer_add_vform = dataform.Form(
    dataform.Item("realname", rules.not_null, description=u"用户姓名"),
    dataform.Item("node_id", rules.not_null, description=u"区域id"),
    dataform.Item("idcard", rules.len_of(0, 32), description=u"证件号码"),
    dataform.Item("mobile", rules.len_of(0, 32), description=u"用户手机号码"),
    dataform.Item("email", rules.is_email, description=u"用户Email"),
    dataform.Item("address", description=u"用户地址"),
    dataform.Item("customer_name",description=u"客户自助服务账号"),
    dataform.Item("account_number", rules.not_null, description=u"用户认证账号"),
    dataform.Item("product_id", rules.not_null, description=u"资费id"),
    dataform.Item("password", rules.not_null, description=u"用户密码"),
    dataform.Item("begin_date", rules.is_date, description=u"开通日期"),
    dataform.Item("expire_date", rules.is_date, description=u"过期日期"),
    dataform.Item("balance", rules.is_rmb, description=u"用户余额"),
    dataform.Item("time_length", description=u"用户时长"),
    dataform.Item("flow_length", description=u"用户流量"),
    title="api customer add"
)

@permit.route(r"/api/customer/add")
class CustomerAddHandler(ApiHandler):

    def get(self):
        self.post()

    def post(self):
        form = customer_add_vform()
        try:
            request = self.parse_form_request()
            if not vform.validates(**request):
                raise Exception(vform.errors)
            if self.db.query(models.TrAccount).filter_by(account_number=form.d.account_number).count() > 0:
                raise Exception("account already exists")
        except Exception as err:
            self.render_result(code=1, msg=safeunicode(err.message))
            return

        try:
            customer = models.TrCustomer()
            customer.node_id = form.d.node_id
            customer.realname = form.d.realname
            customer.idcard = form.d.idcard
            customer.customer_name = form.d.customer_name or form.d.account_number
            customer.password = md5(form.d.password.encode()).hexdigest()
            customer.sex = '1'
            customer.age = '0'
            customer.email = ''
            customer.mobile = form.d.mobile
            customer.address = form.d.address
            customer.create_time = form.d.begin_date + ' 00:00:00'
            customer.update_time = utils.get_currtime()
            customer.email_active = 1
            customer.mobile_active = 1
            customer.active_code = utils.get_uuid()
            self.db.add(customer)
            self.db.flush()
            self.db.refresh(customer)

            accept_log = models.TrAcceptLog()
            accept_log.accept_type = 'open'
            accept_log.accept_source = 'api'
            accept_log.accept_desc =  u"API开通账号：%s" % form.d.account_number
            accept_log.account_number = form.d.account_number
            accept_log.accept_time = customer.update_time
            accept_log.operator_name = 'api'
            self.db.add(accept_log)
            self.db.flush()
            self.db.refresh(accept_log)

            order_fee = 0
            actual_fee = 0
            balance = 0
            time_length = 0
            flow_length = 0
            expire_date = form.d.expire_date
            product = self.db.query(models.TrProduct).get(form.d.product_id)
            # 买断时长
            if product.product_policy == BOTimes:
                time_length = int(form.d.time_length)
            # 买断流量
            elif product.product_policy == BOFlows:
                flow_length = int(form.d.flow_length)
            # 预付费时长,预付费流量
            elif product.product_policy in (PPTimes, PPFlow):
                balance = utils.yuan2fen(form.d.balance)
                expire_date = MAX_EXPIRE_DATE

            order = models.TrCustomerOrder()
            order.order_id = utils.gen_order_id()
            order.customer_id = customer.customer_id
            order.product_id = form.d.product.id
            order.account_number = form.d.account_number
            order.order_fee = order_fee
            order.actual_fee = actual_fee
            order.pay_status = 1
            order.accept_id = accept_log.id
            order.order_source = 'console'
            order.create_time = customer.update_time
            order.order_desc = u"用户导入开户"
            self.db.add(order)

            account = models.TrAccount()
            account.account_number = form.d.account_number
            account.customer_id = customer.customer_id
            account.product_id = order.product_id
            account.install_address = customer.address
            account.ip_address = ''
            account.mac_addr = ''
            account.password = self.aes.encrypt(form.d.password)
            account.status = 1
            account.balance = balance
            account.time_length = time_length
            account.flow_length = flow_length
            account.expire_date = expire_date
            account.user_concur_number = product.concur_number
            account.bind_mac = product.bind_mac
            account.bind_vlan = product.bind_vlan
            account.vlan_id1 = 0
            account.vlan_id2 = 0
            account.create_time = customer.create_time
            account.update_time = customer.update_time
            self.db.add(account)
            self.add_oplog(u"API开户，%s" % form.d.account_number)
            self.db.commit()
            self.render_result(code=0, msg='success')
        except Exception as e:
            self.render_result(code=1, msg=safeunicode(e.message))


