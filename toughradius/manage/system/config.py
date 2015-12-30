#!/usr/bin/env python
# coding:utf-8
import ConfigParser

import cyclone.auth
import cyclone.escape
import cyclone.web

from toughradius.manage.base import BaseHandler
from toughlib.permit import permit
from toughradius.manage.system import config_forms
from toughradius.manage import models
from toughradius.manage.settings import * 


@permit.route(r"/admin/config", u"数据库配置", MenuSys, order=2.0000, is_menu=True)
class ConfigHandler(BaseHandler):
    @cyclone.web.authenticated
    def get(self):
        database_form = config_forms.database_form()
        database_form.fill(self.settings.config.database)
        self.render("config.html",database_form=database_form)


@permit.route(r"/admin/config/database/update", u"数据库配置", u"系统管理", order=2.0002, is_menu=False)
class DatabaseHandler(BaseHandler):
    @cyclone.web.authenticated
    def post(self):
        config = self.settings.config
        config.database.echo = self.get_argument("echo")
        config.database.dbtype = self.get_argument("dbtype")
        config.database.dburl = self.get_argument("dburl")
        config.database.pool_size = self.get_argument("pool_size")
        config.database.pool_recycle = self.get_argument("pool_recycle")
        config.database.backup_path = self.get_argument("backup_path")
        config.update()
        self.redirect("/config")


