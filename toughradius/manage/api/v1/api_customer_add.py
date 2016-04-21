#!/usr/bin/env python
#coding=utf-8
import time
import traceback
import decimal
from toughlib.btforms import dataform
from toughlib.btforms import rules
from toughlib import utils, apiutils, dispatch
from toughlib.permit import permit
from toughradius.manage.api.apibase import ApiHandler
from toughradius.manage import models
from toughradius.manage.settings import *
from hashlib import md5

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
    dataform.Item("fee_value", rules.is_rmb, description=u"用户缴费金额"),
    dataform.Item("pay_status", rules.is_number, description=u"支付状态"),
    dataform.Item("time_length", rules.is_number, default='0', description=u"用户时长"),
    dataform.Item("flow_length", rules.is_number, default='0', description=u"用户流量"),
    dataform.Item("bind_mac", description=u"用户是否绑定 MAC"),
    dataform.Item("bind_vlan", description=u"用户是否绑定 vlan"),
    dataform.Item("concur_number", description=u"用户并发数"),
    dataform.Item("ip_address",  description=u"用户IP地址"),
    dataform.Item("input_max_limit", description=u"用户上行速度 Mbps"),
    dataform.Item("output_max_limit", description=u"用户下行速度 Mbps"),
    dataform.Item("free_bill_type", description=u"用户计费类型"),
    title="api customer add"
)

@permit.route(r"/api/v1/customer/add")
class CustomerAddHandler(ApiHandler):

    def get_account_attr(self, account_number, attr_name):
        pass

    def set_account_attr(self, account_number, attr_name, attr_value, attr_desc=None):
        pass

    def get(self):
        self.post()

    def post(self):
        form = customer_add_vform()
        try:
            request = self.parse_form_request()
        except apiutils.SignError, err:
            return self.render_sign_err(err)
        except Exception as err:
            return self.render_parse_err(err)

        pay_status = int(form.d.pay_status)
        pay_status_desc = pay_status == 0 and u'未支付' or u"已支付"

        try:
            if not form.validates(**request):
                raise Exception(form.errors)
            if pay_status not in (0,1):
                raise Exception("pay_status must 0 or 1")
            if self.db.query(models.TrAccount).filter_by(account_number=form.d.account_number).count() > 0:
                raise Exception("account already exists")
        except Exception, err:
            return self.render_verify_err(err)

        try:
            customer = models.TrCustomer()
            customer.node_id = form.d.node_id
            customer.realname = form.d.realname
            customer.idcard = form.d.idcard
            customer.customer_name = form.d.customer_name or form.d.account_number
            customer.password = md5(form.d.password.encode()).hexdigest()
            customer.sex = '1'
            customer.age = '0'
            customer.email = form.d.email
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
            accept_log.accept_desc =  u"开通账号：%s, %s" % (form.d.account_number,pay_status_desc)
            accept_log.account_number = form.d.account_number
            accept_log.accept_time = customer.update_time
            accept_log.operator_name = 'api'
            self.db.add(accept_log)
            self.db.flush()
            self.db.refresh(accept_log)

            product = self.db.query(models.TrProduct).get(form.d.product_id)

            order_fee = 0
            actual_fee = utils.yuan2fen(form.d.fee_value or 0)
            balance = 0
            time_length = 0
            flow_length = 0
            expire_date = form.d.expire_date
            user_concur_number = product.concur_number
            user_bind_mac = product.bind_mac
            user_bind_vlan = product.bind_vlan

            if product.product_policy == BOTimes:
                # 买断时长
                time_length = int(form.d.time_length)
            elif product.product_policy == BOFlows:
                # 买断流量
                flow_length = int(form.d.flow_length)
            elif product.product_policy in (PPTimes, PPFlow):
                # 预付费时长,预付费流量
                balance = utils.yuan2fen(form.d.balance)
                expire_date = MAX_EXPIRE_DATE
            elif product.product_policy == FreeFee:
                # 自由资费
                if int(form.d.free_bill_type or 9999) not in (FreeFeeDate,FreeFeeFlow,FreeFeeTimeLen):
                    return self.render_verify_err(msg=u"free_bill_type in (0,1,2)")
                time_length = int(form.d.time_length or 0)
                flow_length = int(form.d.flow_length or 0)
                balance = utils.yuan2fen(form.d.balance or 0)
                user_concur_number = int(form.d.concur_number or 0)
                user_bind_mac = int(form.d.bind_mac or 0)
                user_bind_vlan = int(form.d.bind_vlan or 0)
                user_input_max_limit = utils.mbps2bps(form.d.input_max_limit or 0)
                user_output_max_limit = utils.mbps2bps(form.d.output_max_limit or 0)
                self.set_account_attr(form.d.account_number,'bill_type',form.d.free_bill_type)
                self.set_account_attr(form.d.account_number,'input_max_limit',user_input_max_limit)
                self.set_account_attr(form.d.account_number,'output_max_limit',user_output_max_limit)

            order = models.TrCustomerOrder()
            order.order_id = utils.gen_order_id()
            order.customer_id = customer.customer_id
            order.product_id = product.id
            order.account_number = form.d.account_number
            order.order_fee = order_fee
            order.actual_fee = actual_fee
            order.pay_status = pay_status
            order.accept_id = accept_log.id
            order.order_source = 'api'
            order.create_time = customer.update_time
            order.order_desc = u"开通账号 %s" % pay_status_desc
            self.db.add(order)

            account = models.TrAccount()
            account.account_number = form.d.account_number
            account.customer_id = customer.customer_id
            account.product_id = order.product_id
            account.install_address = customer.address
            account.ip_address = form.d.ip_address
            account.mac_addr = ''
            account.password = self.aes.encrypt(form.d.password)
            account.status = pay_status
            account.balance = balance
            account.time_length = utils.hour2sec(time_length)
            account.flow_length = utils.mb2kb(flow_length)
            account.expire_date = expire_date
            account.user_concur_number = user_concur_number
            account.bind_mac = user_bind_mac
            account.bind_vlan = user_bind_vlan
            account.vlan_id1 = 0
            account.vlan_id2 = 0
            account.create_time = customer.create_time
            account.update_time = customer.update_time
            self.db.add(account)
            self.add_oplog(u"新用户开户，%s， %s" % form.d.account_number,pay_status_desc)

            self.db.commit()
            self.render_success()
        except Exception as e:
            self.render_unknow(err)
            import traceback
            traceback.print_exc()


