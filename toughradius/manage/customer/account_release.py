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
from toughlib import utils,dispatch,db_cache
from toughradius.manage.settings import * 

@permit.route(r"/admin/account/release", u"用户解绑",MenuUser, order=2.8000)
class AccountReleasetHandler(account.AccountHandler):

    @cyclone.web.authenticated
    def post(self):
        account_number = self.get_argument('account_number')  
        user = self.db.query(models.TrAccount).filter_by(account_number=account_number).first()
        user.mac_addr = ''
        user.vlan_id1 = 0
        user.vlan_id2 = 0
        self.add_oplog(u'释放用户账号（%s）绑定信息'%(account_number))
        self.db.commit()
        dispatch.pub(db_cache.CACHE_DELETE_EVENT,account_cache_key(account_number), async=True)
        return self.render_json(msg=u"解绑成功")