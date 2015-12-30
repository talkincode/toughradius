#!/usr/bin/env python
#coding=utf-8

import cyclone.auth
import cyclone.escape
import cyclone.web

from toughradius.manage import models
from toughradius.manage.base import BaseHandler
from toughlib.permit import permit
from toughlib import utils
from toughradius.manage.settings import * 

@permit.route(r"/admin/radagent", u"认证代理",MenuRes, order=4.0000, is_menu=True)
class RadAgentListHandler(BaseHandler):
    @cyclone.web.authenticated
    def get(self):
        agents = self.db.query(models.TrRadAgent)
        return self.render('radagent_list.html',agents=agents)


@permit.route(r"/admin/radagent/delete", u"删除认证代理", MenuRes, order=4.0001)
class RadAgentDeleteHandler(BaseHandler):
    @cyclone.web.authenticated
    def get(self):
        agent_id = self.get_argument("agent_id")
        self.db.query(models.TrRadAgent).filter_by(id=agent_id).delete()
        self.add_oplog(u'删除认证接入代理信息:%s' % agent_id)
        self.db.commit()
        self.redirect("/admin/radagent",permanent=False)


