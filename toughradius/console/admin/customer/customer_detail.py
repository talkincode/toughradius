#!/usr/bin/env python
#coding=utf-8

import cyclone.auth
import cyclone.escape
import cyclone.web
import decimal
import datetime
from tablib import Dataset
from toughradius.console import models
from toughradius.console.admin.customer import customer_forms
from toughradius.console.admin.customer.customer import CustomerHandler
from toughradius.common.permit import permit
from toughradius.common import utils
from toughradius.common.settings import * 


@permit.route(r"/customer/detail", u"用户详情",MenuUser, order=1.2000)
class CustomerDetailHandler(CustomerHandler):

    @cyclone.web.authenticated
    def get(self):
        account_number = self.get_argument('account_number')
        user = self.db.query(
            models.TrCustomer.realname,
            models.TrAccount.customer_id,
            models.TrAccount.account_number,
            models.TrAccount.password,
            models.TrAccount.expire_date,
            models.TrAccount.balance,
            models.TrAccount.time_length,
            models.TrAccount.flow_length,
            models.TrAccount.user_concur_number,
            models.TrAccount.status,
            models.TrAccount.mac_addr,
            models.TrAccount.vlan_id,
            models.TrAccount.vlan_id2,
            models.TrAccount.ip_address,
            models.TrAccount.bind_mac,
            models.TrAccount.bind_vlan,
            models.TrAccount.ip_address,
            models.TrAccount.install_address,
            models.TrAccount.last_pause,
            models.TrAccount.create_time,
            models.TrProduct.product_name,
            models.TrProduct.product_policy
        ).filter(
            models.TrProduct.id == models.TrAccount.product_id,
            models.TrCustomer.customer_id == models.TrAccount.customer_id,
            models.TrAccount.account_number == account_number
        ).first()

        customer = self.db.query(models.TrCustomer).get(user.customer_id)

        orders = self.db.query(
            models.TrCustomerOrder.order_id,
            models.TrCustomerOrder.order_id,
            models.TrCustomerOrder.product_id,
            models.TrCustomerOrder.account_number,
            models.TrCustomerOrder.order_fee,
            models.TrCustomerOrder.actual_fee,
            models.TrCustomerOrder.pay_status,
            models.TrCustomerOrder.create_time,
            models.TrCustomerOrder.order_desc,
            models.TrProduct.product_name
        ).filter(
            models.TrProduct.id == models.TrCustomerOrder.product_id,
            models.TrCustomerOrder.account_number == account_number
        ).order_by(models.TrCustomerOrder.create_time.desc())

        accepts = self.db.query(
            models.TrAcceptLog.id,
            models.TrAcceptLog.accept_type,
            models.TrAcceptLog.accept_time,
            models.TrAcceptLog.accept_desc,
            models.TrAcceptLog.operator_name,
            models.TrAcceptLog.accept_source,
            models.TrAcceptLog.account_number,
            models.TrCustomer.node_id,
            models.TrNode.node_name
        ).filter(
            models.TrAcceptLog.account_number == models.TrAccount.account_number,
            models.TrCustomer.customer_id == models.TrAccount.customer_id,
            models.TrNode.id == models.TrCustomer.node_id,
            models.TrAcceptLog.account_number == account_number
        ).order_by(models.TrAcceptLog.accept_time.desc())

        get_orderid = lambda aid: self.db.query(models.TrCustomerOrder.order_id).filter_by(accept_id=aid).scalar()

        type_map = ACCEPT_TYPES

        return self.render("customer_detail.html",
                          customer=customer,
                          user=user,
                          orders=orders,
                          accepts=accepts,
                          type_map=type_map,
                          get_orderid=get_orderid)


