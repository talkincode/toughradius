#!/usr/bin/env python
#coding:utf-8
import cyclone.auth
import cyclone.escape
import cyclone.web
import datetime
import time
import os
from beaker.cache import cache_managers
from toughradius.manage.base import BaseHandler
from toughlib.permit import permit
from toughradius.manage import models
from toughradius.manage.settings import * 
from toughradius.common import tools
import psutil

@permit.route(r"/admin")
class HomeHandler(BaseHandler):
    @cyclone.web.authenticated
    def get(self):
        # cpuuse = psutil.cpu_percent(interval=None, percpu=True)
        # memuse = psutil.virtual_memory()
        # online_count = self.db.query(models.TrOnline.id).count()
        # user_total = self.db.query(models.TrAccount.account_number).filter_by(status=1).count()
        # self.render("index.html",config=self.settings.config,
        #     cpuuse=cpuuse,memuse=memuse,online_count=online_count,user_total=user_total)
        self.redirect("/admin/dashboard")


@permit.route(r"/")
class HomeHandler(BaseHandler):
    @cyclone.web.authenticated
    def get(self):
        self.redirect("/admin/dashboard")

@permit.route(r"/about")
class HomeHandler(BaseHandler):
    @cyclone.web.authenticated
    def get(self):
        self.render("about.html")

@permit.route(r"/toughcloud/service/register")
class ToughcloudRegisterHandler(BaseHandler):

    def get_toughcloud_url(self):
        if os.environ.get("TR_DEV"):
            return 'http://127.0.0.1:9079/customer/license/request?sid=%s'%tools.get_sys_uuid()
        else:
            return 'https://www.toughcloud.net/customer/license/request?sid=%s'%tools.get_sys_uuid()

    @cyclone.web.authenticated
    def get(self):
        self.redirect(self.get_toughcloud_url())





