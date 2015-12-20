#!/usr/bin/env python
#coding:utf-8
from toughradius.console.admin.base import BaseHandler
from toughradius.common.permit import permit

@permit.route(r"/logout")
class LogoutHandler(BaseHandler):

    def get(self):
        if not self.current_user:
            self.redirect("/login")
            return
        self.clear_session()
        self.redirect("/login",permanent=False)


