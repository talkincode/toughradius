#!/usr/bin/env python
#coding=utf-8

import cyclone.auth
import cyclone.escape
import cyclone.web
import decimal
import datetime
from toughradius.console import models
from toughradius.console.admin.base import BaseHandler
from toughradius.console.admin.customer import account
from toughradius.common.permit import permit
from toughradius.common import utils
from toughradius.common.settings import * 

@permit.route(r"/admin/account/resume", u"用户复机",MenuUser, order=2.1000)
class AccountResumetHandler(account.AccountHandler):

    @cyclone.web.authenticated
    def post(self):
        account_number = self.get_argument("account_number")
        account = self.db.query(models.TrAccount).get(account_number)
        if account.status != 2:
            return self.render_json(code=1, msg=u"用户当前状态不允许复机")

        account.status = 1
        _datetime = datetime.datetime.now()
        _pause_time = datetime.datetime.strptime(account.last_pause, "%Y-%m-%d %H:%M:%S")
        _expire_date = datetime.datetime.strptime(account.expire_date + ' 23:59:59', "%Y-%m-%d %H:%M:%S")
        days = (_expire_date - _pause_time).days
        new_expire = (_datetime + datetime.timedelta(days=int(days))).strftime("%Y-%m-%d")
        account.expire_date = new_expire

        accept_log = models.TrAcceptLog()
        accept_log.accept_type = 'resume'
        accept_log.accept_source = 'console'
        accept_log.accept_desc = u"用户复机：上网账号:%s" % (account_number)
        accept_log.account_number = account.account_number
        accept_log.accept_time = utils.get_currtime()
        accept_log.operator_name = self.current_user.username
        self.db.add(accept_log)

        self.db.commit()
        return self.render_json(msg=u"操作成功")




