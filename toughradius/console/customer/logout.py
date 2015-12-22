#!/usr/bin/env python
#coding:utf-8
from toughradius.console.customer.base import BaseHandler
from toughradius.common.permit import permit

@permit.route(r"/customer/logout")
class LogoutHandler(BaseHandler):

    def get(self):
        if not self.current_user:
            self.redirect("/customer/login")
            return
        self.clear_session()
        self.redirect("/customer/login",permanent=False)
