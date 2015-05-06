#!/usr/bin/env python
#coding:utf-8
import sys
import os.path
import cyclone.auth
import cyclone.escape
import cyclone.web
from toughradius.console.portal.base import BaseHandler
from toughradius.console.portal.login import LoginHandler
from toughradius.console.portal.weixin_login import MpLoginHandler
from toughradius.console.portal.logout import LogoutHandler



class HomeHandler(BaseHandler):
    @cyclone.web.authenticated
    def get(self):
        self.render(self.get_index_template())
        