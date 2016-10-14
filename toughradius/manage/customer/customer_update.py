#!/usr/bin/env python
#coding=utf-8

import cyclone.auth
import cyclone.escape
import cyclone.web
import decimal
import datetime
from hashlib import md5
from toughradius import models
from toughradius.manage.customer import customer_forms
from toughradius.manage.customer.customer import CustomerHandler
from toughlib.permit import permit
from toughlib import utils
from toughradius.manage.settings import * 


@permit.route(r"/admin/customer/update", u"用户修改",MenuUser, order=1.4000)
class CustomerUpdateHandler(CustomerHandler):

    @cyclone.web.authenticated
    def get(self):
        customer_id = self.get_argument("customer_id")
        account_number = self.get_argument("account_number")
        customer = self.db.query(models.TrCustomer).get(customer_id)
        nodes = [(n.id, n.node_name) for n in self.get_opr_nodes()]
        form = customer_forms.customer_update_form(nodes)
        form.fill(customer)
        form.account_number.set_value(account_number)
        return self.render("base_form.html", form=form)

    @cyclone.web.authenticated
    def post(self):
        nodes = [(n.id, n.node_name) for n in self.get_opr_nodes()]
        form = customer_forms.customer_update_form(nodes)
        if not form.validates(source=self.get_params()):
            return self.render("base_form.html", form=form)

        customer = self.db.query(models.TrCustomer).get(form.d.customer_id)
        customer.realname = form.d.realname
        if form.d.new_password:
            customer.password = md5(form.d.new_password.encode()).hexdigest()
        customer.email = form.d.email
        customer.idcard = form.d.idcard
        customer.mobile = form.d.mobile
        customer.address = form.d.address
        customer.customer_desc = form.d.customer_desc
        self.add_oplog(u"修改用户信息 %s" % customer.customer_name)
        self.db.commit()
        self.redirect(self.detail_url_fmt(form.d.account_number))



