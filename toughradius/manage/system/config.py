#!/usr/bin/env python
# coding:utf-8
import os
import cyclone.web
from toughlib import utils, logger, dispatch
from toughradius.manage.base import BaseHandler
from toughlib.permit import permit
from toughradius.manage.system import config_forms
from toughradius.manage import models
from toughradius.manage.settings import * 

@permit.route(r"/admin/config", u"系统配置管理", MenuSys, order=2.0000, is_menu=True)
class ConfigHandler(BaseHandler):
    @cyclone.web.authenticated
    def get(self):
        active = self.get_argument("active", "system")
        system_form = config_forms.system_form()
        system_form.fill(self.settings.config.system)
        database_form = config_forms.database_form()
        if 'DB_TYPE' in os.environ and 'DB_URL' in os.environ:
            self.settings.config['database']['dbtype'] = os.environ.get('DB_TYPE')
            self.settings.config['database']['dburl'] = os.environ.get('DB_URL')

        database_form.fill(self.settings.config.database)        
        syslog_form = config_forms.syslog_form()
        syslog_form.fill(self.settings.config.syslog)
        self.render("config.html",
            active=active,
            system_form=system_form,
            database_form=database_form,
            syslog_form=syslog_form)


@permit.route(r"/admin/config/system/update", u"系统配置", u"系统管理", order=2.0001, is_menu=False)
class DefaultHandler(BaseHandler):
    @cyclone.web.authenticated
    def post(self):
        config = self.settings.config
        config['system']['debug'] = int(self.get_argument("debug"))
        config['system']['tz'] = self.get_argument("tz")
        config['system']['license'] = self.get_argument("license")
        config['system']['secret'] = self.get_argument("secret")
        config.save()
        self.redirect("/admin/config?active=system")

@permit.route(r"/admin/config/database/update", u"数据库配置", u"系统管理", order=2.0002, is_menu=False)
class DatabaseHandler(BaseHandler):
    @cyclone.web.authenticated
    def post(self):
        config = self.settings.config
        config['database']['echo'] = int(self.get_argument("echo"))
        config['database']['dbtype'] = self.get_argument("dbtype")
        config['database']['dburl'] = self.get_argument("dburl")
        config['database']['pool_size'] = int(self.get_argument("pool_size"))
        config['database']['pool_recycle'] = int(self.get_argument("pool_recycle"))
        # config['database']['backup_path'] = self.get_argument("backup_path")
        config.save()
        self.redirect("/admin/config?active=database")


@permit.route(r"/admin/config/syslog/update", u"syslog 配置", u"系统管理", order=2.0003, is_menu=False)
class SyslogHandler(BaseHandler):
    @cyclone.web.authenticated
    def post(self):
        self.settings.config['syslog']['enable'] = int(self.get_argument("enable"))
        self.settings.config['syslog']['server'] = self.get_argument("server")
        self.settings.config['syslog']['port'] = int(self.get_argument("port",514))
        self.settings.config['syslog']['level'] = self.get_argument("level")
        self.settings.config.save()
        dispatch.pub(logger.EVENT_SETUP,self.settings.config)
        self.redirect("/admin/config?active=syslog")

@permit.route(r"/admin/config/secret/update", u"系统密钥更新", u"系统管理", order=2.0004, is_menu=False)
class SecretHandler(BaseHandler):

    @cyclone.web.authenticated
    def post(self):
        new_secret = utils.gen_secret(32)
        new_aes = utils.AESCipher(key=new_secret)
        users = self.db.query(models.TrAccount)
        for user in users:
            oldpwd = self.aes.decrypt(user.password)
            user.password = new_aes.encrypt(oldpwd)

        self.application.aes = new_aes
        self.db.commit()
        self.settings.config['system']['secret'] = new_secret
        self.settings.config.save()
        self.render_json(code=0)









