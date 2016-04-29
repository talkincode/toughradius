#!/usr/bin/env python
#coding=utf-8
import traceback
import decimal
from toughlib import utils, apiutils,dispatch
from hashlib import md5
from toughlib.permit import permit
from toughradius.manage.api.apibase import ApiHandler
from toughradius.manage import models
from toughradius.manage.customer.account import AccountCalc
from toughradius.manage.settings import * 
from toughradius.manage.events.settings import *

""" 客户续费
"""

@permit.route(r"/api/v1/account/renew")
class AccountRenewHandler(ApiHandler,AccountCalc):
    """ @param: 
        account_number: str,
    """

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
        try:
            request = self.parse_form_request()
        except apiutils.SignError, err:
            return self.render_sign_err(err)
        except Exception as err:
            return self.render_parse_err(err)

        try:
            pay_status = int(request.get('pay_status',0))
            pay_status_desc = pay_status == 0 and u'未支付' or u"已支付"

            if pay_status == 0:
                self.cache.set("renew_order_%s"%order_id, request, 24 * 60 * 60)
                self.render_success()
            elif pay_status == 1:
                request = self.cache.get("renew_order_%s"%order_id)
                if not request:
                    return self.render_verify_err(msg=u"订单不存在")
            else:
                return self.render_verify_err(msg=u"支付状态不正确")

            account_number = request.get('account_number')
            order_id = request.get('order_id')
            expire_date = request.get('expire_date')
            months = int(request.get('months',0))
            giftdays = int(request.get('giftdays',0))
            fee_value = request.get("fee_value",0)

            if not account_number:
                return self.render_verify_err(msg=u"账号不能为空")

            if utils.yuan2fen(request.get("fee_value",0)) < 0:
                return self.render_verify_err(msg=u"无效续费金额 %s"%fee_value)


            account = self.db.query(models.TrAccount).get(account_number)
            user = self.query_account(account_number)

            if account.status not in (1, 4):
                return self.render_verify_err(msg=u"无效用户状态")

            accept_log = models.TrAcceptLog()
            accept_log.accept_type = 'next'
            accept_log.accept_source = 'api'
            accept_log.accept_desc = u"用户续费：上网账号:%s，续费%s元" % (account_number, fee_value)
            accept_log.account_number = account_number
            accept_log.accept_time = utils.get_currtime()
            accept_log.operator_name = 'api'
            self.db.add(accept_log)
            self.db.flush()
            self.db.refresh(accept_log)

            order_fee = 0
            product = self.db.query(models.TrProduct).get(user.product_id)

            # 预付费包月
            if product.product_policy == PPMonth:
                order_fee = decimal.Decimal(product.fee_price) * decimal.Decimal(months)
                order_fee = int(order_fee.to_integral_value())

            # 买断包月,买断流量,买断时长
            elif product.product_policy in (BOMonth, BOTimes, BOFlows):
                order_fee = int(product.fee_price)

            order = models.TrCustomerOrder()
            order.order_id = order_id
            order.customer_id = user.customer_id
            order.product_id = user.product_id
            order.account_number = account_number
            order.order_fee = order_fee
            order.actual_fee = utils.yuan2fen(fee_value)
            order.pay_status = 1
            order.accept_id = accept_log.id
            order.order_source = 'api'
            order.create_time = utils.get_currtime()

            old_expire_date = account.expire_date
            old_time_length = account.time_length
            old_flow_length = account.flow_length

            ### 计算新的到期时间
            
            new_expire_date = expire_date
            if not new_expire_date:
                calc_result = self.calc(months, user.product_id, old_expire_date, giftdays)
                new_expire_date = calc_result.get("expire_date")

            account.status = 1
            account.expire_date = new_expire_date
            if product.product_policy == BOTimes:
                account.time_length += product.fee_times
            elif product.product_policy == BOFlows:
                account.flow_length += product.fee_flows

            if product.product_policy in (PPMonth,BOMonth):
                order.order_desc = u"用户续费,续费前到期:%s,续费后到期:%s, 赠送天数: %s" % (
                    old_expire_date, account.expire_date, giftdays)
            elif product.product_policy == BOTimes:
                order.order_desc = u"用户续费,续费前时长:%s小时,续费后时长:%s小时" % (
                    utils.sec2hour(old_time_length), utils.sec2hour(account.time_length))
            elif product.product_policy == BOFlows:
                order.order_desc = u"用户续费,续费前流量:%sMB,续费后流量:%sMB" % (
                    utils.kb2mb(old_flow_length), utils.kb2mb(account.flow_length))

            self.db.add(order)
            self.add_oplog(order.order_desc)
            self.db.commit()
            self.cache.delete("renew_order_%s"%order_id)
            self.render_success()
            dispatch.pub(ACCOUNT_NEXT_EVENT,order.account_number, async=True)
            dispatch.pub(CACHE_DELETE_EVENT,account_cache_key(account.account_number), async=True)
        except Exception as err:
            self.render_unknow(err)
            import traceback
            traceback.print_exc()
















