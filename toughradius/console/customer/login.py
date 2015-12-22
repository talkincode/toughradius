#!/usr/bin/env python
#coding:utf-8
from hashlib import md5
from toughradius.common import utils
from toughradius.console.customer.base import BaseHandler
from toughradius.console.customer import forms
from toughradius.console import models
from toughradius.common.permit import permit
from toughradius.common import validate

vcache = validate.ValidateCache()

@permit.route(r"/customer/login")
class LoginHandler(BaseHandler):

    def get(self):
        form = forms.customer_login_form()
        form.next.set_value(self.get_argument('next', '/customer'))
        self.render("login.html", form=form)

    def post(self):
        next = self.get_argument("next", "/customer")
        form = forms.customer_login_form()
        if not form.validates(source=self.get_params()):
            return self.render("login.html", form=form)

        if vcache.is_over(form.d.username, '0'):
            return render_error(msg=u"用户一小时内登录错误超过5次，请一小时后再试")

        customer = self.db.query(models.TrCustomer).filter_by(
            customer_name=form.d.username
        ).first()

        if not customer:
            return self.render("login.html", form=form, msg=u"用户不存在")

        if customer.password != md5(form.d.password.encode()).hexdigest():
            vcache.incr(form.d.username, '0')
            return self.render("login.html", form=form, msg=u"用户名密码错误第%s次" % vcache.errs(form.d.username, '0'))

        vcache.clear(form.d.username, '0')

        self.set_session_user(customer, self.request.remote_ip, utils.get_currtime())
        self.redirect(next)


