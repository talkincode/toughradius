#!/usr/bin/env python
#coding:utf-8
import cyclone.auth
import cyclone.escape
import cyclone.web
from beaker.cache import cache_managers
from toughradius.console.admin.base import BaseHandler
from toughradius.common.permit import permit
from toughradius.common.settings import * 

@permit.route(r"/admin")
class HomeHandler(BaseHandler):
    @cyclone.web.authenticated
    def get(self):
        self.render("index.html")


@permit.route(r"/")
class HomeHandler(BaseHandler):
    @cyclone.web.authenticated
    def get(self):
        self.redirect("/admin")


