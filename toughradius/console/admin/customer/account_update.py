#!/usr/bin/env python
#coding=utf-8

import cyclone.auth
import cyclone.escape
import cyclone.web
import decimal
from toughradius.console import models
from toughradius.console.admin.base import BaseHandler
from toughradius.console.admin.customer import account, account_forms
from toughradius.common.permit import permit
from toughradius.common import utils
from toughradius.common.settings import * 

@permit.route(r"/account/update", u"用户策略修改",MenuUser, order=2.2000)
class AccountUpdatetHandler(account.AccountHandler):

    @cyclone.web.authenticated
    def get(self):
        account_number = self.get_argument("account_number",None)
        account = self.db.query(models.TrAccount).get(account_number)
        form = account_forms.account_update_form()
        form.fill(account)
        self.render("base_form.html", form=form)

    @cyclone.web.authenticated
    def post(self):
        form = account_forms.account_update_form()
        if not form.validates(source=self.get_params()):
            return self.render("base_form.html", form=form)

        account = self.db.query(models.TrAccount).get(form.d.account_number)
        account.ip_address = form.d.ip_address
        account.install_address = form.d.install_address
        account.user_concur_number = form.d.user_concur_number
        account.bind_mac = form.d.bind_mac
        account.bind_vlan = form.d.bind_vlan
        account.account_desc = form.d.account_desc
        if form.d.new_password:
            account.password = self.aes.encrypt(form.d.new_password)

        self.add_oplog(u'修改上网账号信息:%s' % (account.account_number))
        self.db.commit()
        self.redirect(self.detail_url_fmt(account.account_number))




