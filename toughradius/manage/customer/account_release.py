#!/usr/bin/env python
#coding=utf-8
import cyclone.web
import decimal
from toughradius import models
from toughradius.manage.base import BaseHandler
from toughradius.manage.customer import account, account_forms
from toughradius.common.permit import permit
from toughradius.common import utils,dispatch
from toughradius import events
from toughradius import settings 

@permit.route(r"/admin/account/release", u"用户解绑",settings.MenuUser, order=2.8000)
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

        dispatch.pub(events.CACHE_DELETE_EVENT,
            settings.ACCOUNT_CACHE_KEY(account_number), async=True)
        
        return self.render_json(msg=u"解绑成功")