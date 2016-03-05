#!/usr/bin/env python
#coding=utf-8
import time
import traceback
from toughlib.btforms import dataform
from toughlib.btforms import rules
from toughlib import utils, apiutils, dispatch
from toughlib import db_cache as cache
from toughlib.permit import permit
from toughradius.manage.api.apibase import ApiHandler
from toughradius.manage import models
from toughradius.manage.events.settings import ACCOUNT_DELETE_EVENT
from toughradius.manage.settings import * 
from hashlib import md5

""" 客户账号删除，删除客户账号资料及相关数据，但不删除客户信息
"""

@permit.route(r"/api/account/delete")
class AccountDeleteHandler(ApiHandler):
    """ @param: 
        account_number: str,
    """

    def get(self):
        self.post()

    def post(self):
        try:
            request = self.parse_form_request()
            account_number = request.get('account_number')

            if not account_number:
                raise Exception("account_number is empty")

            account = self.db.query(models.TrAccount).filter_by(account_number=account_number).first()
            if not account:
                raise Exception("account is not exists")

            self.db.query(models.TrAcceptLog).filter_by(account_number=account.account_number).delete()
            self.db.query(models.TrAccountAttr).filter_by(account_number=account.account_number).delete()
            self.db.query(models.TrBilling).filter_by(account_number=account.account_number).delete()
            self.db.query(models.TrTicket).filter_by(account_number=account.account_number).delete()
            self.db.query(models.TrOnline).filter_by(account_number=account.account_number).delete()
            self.db.query(models.TrAccount).filter_by(account_number=account.account_number).delete()
            self.db.query(models.TrCustomerOrder).filter_by(account_number=account.account_number).delete()
            self.add_oplog(u'API删除用户账号%s' % (account.account_number))
            self.db.commit()
            dispatch.pub(ACCOUNT_DELETE_EVENT, account.account_number, async=True)
            dispatch.pub(cache.CACHE_DELETE_EVENT,account_cache_key(account.account_number), async=True) 
            return self.render_result(code=0, msg='success')
        except Exception as err:
            self.render_result(code=1, msg=utils.safeunicode(err.message))
            import traceback
            traceback.print_exc()
            return















