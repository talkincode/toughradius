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
from toughlib import utils, dispatch
from toughradius.manage.settings import * 
from toughradius.manage.events.settings import ACCOUNT_DELETE_EVENT

@permit.route(r"/admin/account/delete", u"用户账号删除",MenuUser, order=2.6000)
class AccountDeleteHandler(account.AccountHandler):

    @cyclone.web.authenticated
    def get(self):
        account_number = self.get_argument("account_number")
        if not account_number:
            self.render_error(msg=u'account_number is empty')
        account = self.db.query(models.TrAccount).get(account_number)
        customer_id = account.customer_id

        self.db.query(models.TrAcceptLog).filter_by(account_number=account.account_number).delete()
        self.db.query(models.TrAccountAttr).filter_by(account_number=account.account_number).delete()
        self.db.query(models.TrBilling).filter_by(account_number=account.account_number).delete()
        self.db.query(models.TrTicket).filter_by(account_number=account.account_number).delete()
        self.db.query(models.TrOnline).filter_by(account_number=account.account_number).delete()
        self.db.query(models.TrAccount).filter_by(account_number=account.account_number).delete()
        self.db.query(models.TrCustomerOrder).filter_by(account_number=account.account_number).delete()
        self.add_oplog(u'删除用户账号%s' % (account_number))
        self.db.commit()
        dispatch.pub(ACCOUNT_DELETE_EVENT, account.account_number, async=True)
        return self.redirect("/admin/customer")





