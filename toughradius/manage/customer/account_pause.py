#!/usr/bin/env python
#coding=utf-8

import cyclone.auth
import cyclone.escape
import cyclone.web
import decimal
import datetime
from toughradius import models
from toughradius.manage.base import BaseHandler
from toughradius.manage.customer import account, account_forms
from toughlib.permit import permit
from toughlib import utils, dispatch
from toughradius.manage.settings import * 
from toughradius.events import settings
from toughradius.events.settings import ACCOUNT_PAUSE_EVENT
from toughradius.events.settings import UNLOCK_ONLINE_EVENT


@permit.route(r"/admin/account/pause", u"用户停机",MenuUser, order=2.1000)
class AccountPausetHandler(account.AccountHandler):

    @cyclone.web.authenticated
    def post(self):
        account_number = self.get_argument("account_number")
        account = self.db.query(models.TrAccount).get(account_number)

        if account.status != 1:
            return self.render_json(code=1, msg=u"用户当前状态不允许停机")

        _datetime = utils.get_currtime()
        account.last_pause = _datetime
        account.status = 2

        accept_log = models.TrAcceptLog()
        accept_log.accept_type = 'pause'
        accept_log.accept_source = 'console'
        accept_log.accept_desc = u"用户停机：上网账号:%s" % (account_number)
        accept_log.account_number = account.account_number
        accept_log.accept_time = _datetime
        accept_log.operator_name = self.current_user.username
        self.db.add(accept_log)

        self.db.commit()

        dispatch.pub(ACCOUNT_PAUSE_EVENT, account.account_number, async=True)
        dispatch.pub(settings.CACHE_DELETE_EVENT,account_cache_key(account.account_number), async=True)

        for online in self.db.query(models.TrOnline).filter_by(account_number=account_number):
            dispatch.pub(UNLOCK_ONLINE_EVENT,account_number,online.nas_addr, online.acct_session_id,async=True)

        return self.render_json(msg=u"操作成功")




