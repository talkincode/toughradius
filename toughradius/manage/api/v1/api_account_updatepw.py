#!/usr/bin/env python
#coding=utf-8

import traceback
from toughradius.common import apiutils, dispatch
from toughradius.common import db_cache as cache
from toughradius.common.permit import permit
from toughradius.manage.api.apibase import ApiHandler
from toughradius import models
from toughradius.events.settings import ACCOUNT_CHANGE_EVENT
from toughradius.manage.settings import * 

""" 客户上网账号密码修改
"""

@permit.route(r"/api/v1/account/pw/update")
class AccountPwUpdateHandler(ApiHandler):
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

            password = request.get("password")

            account.password = self.aes.encrypt(password)

            self.add_oplog(u'修改用户上网账号%s密码' % account.account_number)
            self.db.commit()
            dispatch.pub(ACCOUNT_CHANGE_EVENT, account.account_number, async=True)
            dispatch.pub(cache.CACHE_UPDATE_EVENT, account_cache_key(account.account_number), async=True)
            self.render_success()
        except Exception as err:
            self.render_unknow(err)
            import traceback
            traceback.print_exc()















