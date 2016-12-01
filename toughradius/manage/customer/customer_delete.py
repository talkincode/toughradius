#!/usr/bin/env python
#coding=utf-8

import cyclone.auth
import cyclone.escape
import cyclone.web
import decimal
import datetime
from toughradius import models
from toughradius.manage.customer import customer_forms
from toughradius.manage.customer.customer import CustomerHandler
from toughradius.common.permit import permit
from toughradius.common import utils,logger,dispatch,redis_cache
from toughradius.manage.settings import * 
from toughradius.events import settings
from toughradius.events.settings import ACCOUNT_DELETE_EVENT


@permit.route(r"/admin/customer/delete", u"用户资料删除",MenuUser, order=1.5000)
class CustomerDeleteHandler(CustomerHandler):

    @cyclone.web.authenticated
    def get(self):
        customer_id = self.get_argument("customer_id")
        if not customer_id:
            return self.render_error(msg=u'无效的客户ID')

        for account in self.db.query(models.TrAccount).filter_by(customer_id=customer_id):
            self.db.query(models.TrAcceptLog).filter_by(account_number=account.account_number).delete()
            self.db.query(models.TrAccountAttr).filter_by(account_number=account.account_number).delete()
            self.db.query(models.TrBilling).filter_by(account_number=account.account_number).delete()
            self.db.query(models.TrTicket).filter_by(account_number=account.account_number).delete()
            self.db.query(models.TrOnline).filter_by(account_number=account.account_number).delete()
            self.db.query(models.TrAccount).filter_by(account_number=account.account_number).delete()
            self.db.query(models.TrCustomerOrder).filter_by(account_number=account.account_number).delete()
            self.add_oplog(u'删除用户账号%s' % (account.account_number))
            dispatch.pub(ACCOUNT_DELETE_EVENT, account.account_number, async=True)
            dispatch.pub(settings.CACHE_DELETE_EVENT,account_cache_key(account.account_number), async=True)

        self.db.query(models.TrCustomer).filter_by(customer_id=customer_id).delete()
        self.add_oplog(u'删除用户资料 %s' % (customer_id))    
        self.db.commit()

        
        return self.redirect("/admin/customer")

