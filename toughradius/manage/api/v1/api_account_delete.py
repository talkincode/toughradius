#!/usr/bin/env python
#coding=utf-8
import time
import traceback
from toughradius.common.btforms import dataform
from toughradius.common.btforms import rules
from toughradius.common import utils, apiutils, dispatch
from toughradius.common import db_cache as cache
from toughradius.common.permit import permit
from toughradius.manage.api.apibase import ApiHandler
from toughradius import models
from toughradius.events.settings import ACCOUNT_DELETE_EVENT
from toughradius import settings 
from hashlib import md5

""" 客户账号删除，删除客户账号资料及相关数据，但不删除客户信息
"""

@permit.route(r"/api/v1/account/delete")
class AccountDeleteHandler(ApiHandler):
    """ @param: 
        account_number: str,
    """

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
            account_number = request.get('account_number')

            if not account_number:
                return self.render_verify_err(msg="account_number is empty")

            account = self.db.query(models.TrAccount).filter_by(account_number=account_number).first()
            if not account:
                return self.render_verify_err(msg="account is not exists")

            self.db.query(models.TrAcceptLog).filter_by(account_number=account.account_number).delete()
            self.db.query(models.TrAccountAttr).filter_by(account_number=account.account_number).delete()
            self.db.query(models.TrBilling).filter_by(account_number=account.account_number).delete()
            self.db.query(models.TrTicket).filter_by(account_number=account.account_number).delete()
            self.db.query(models.TrOnline).filter_by(account_number=account.account_number).delete()
            self.db.query(models.TrAccount).filter_by(account_number=account.account_number).delete()
            self.db.query(models.TrCustomerOrder).filter_by(account_number=account.account_number).delete()
            self.add_oplog(u'删除用户账号%s' % (account.account_number))
            self.db.commit()
            dispatch.pub(ACCOUNT_DELETE_EVENT, account.account_number, async=True)
            dispatch.pub(cache.CACHE_DELETE_EVENT,ACCOUNT_CACHE_KEY(account.account_number), async=True) 
            self.render_success()
        except Exception as err:
            self.render_unknow(err)
            import traceback
            traceback.print_exc()















