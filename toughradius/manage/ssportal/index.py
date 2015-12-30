#!/usr/bin/env python
#coding:utf-8
import cyclone.auth
import cyclone.escape
import cyclone.web
from beaker.cache import cache_managers
from toughradius.manage.ssportal.base import BaseHandler, authenticated
from toughradius.manage import models
from toughradius.manage.ssportal import forms
from toughlib.permit import permit
from toughradius.manage.settings import * 

@permit.route(r"/customer")
class HomeHandler(BaseHandler):
    @authenticated
    def get(self):
        customer = self.db.query(models.TrCustomer).filter_by(
            customer_name=self.current_user.username).first()
        accounts = self.db.query(
            models.TrCustomer.realname,
            models.TrAccount.customer_id,
            models.TrAccount.account_number,
            models.TrAccount.expire_date,
            models.TrAccount.balance,
            models.TrAccount.time_length,
            models.TrAccount.flow_length,
            models.TrAccount.status,
            models.TrAccount.last_pause,
            models.TrAccount.create_time,
            models.TrProduct.product_name,
            models.TrProduct.product_policy
        ).filter(
            models.TrProduct.id == models.TrAccount.product_id,
            models.TrCustomer.customer_id == models.TrAccount.customer_id,
            models.TrAccount.customer_id == customer.customer_id
        )
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
            models.TrCustomerOrder.customer_id==customer.customer_id
        ).order_by(models.TrCustomerOrder.create_time.desc())

        status_colors = {0:'',1:'',2:'class="warning"',3:'class="danger"',4:'class="warning"'}
        online_colors = lambda a : self.get_online_status(a) and 'class="success"' or ''
        return  self.render("index.html",
            customer=customer,
            accounts=accounts,
            orders=orders,
            status_colors=status_colors,
            online_colors = online_colors
        )    


