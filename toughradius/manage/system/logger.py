#!/usr/bin/env python
#coding:utf-8
import os
import datetime
import time
import cyclone.escape
import cyclone.web
from toughradius.manage.base import BaseHandler,authenticated
from toughlib.permit import permit
from toughlib import utils, logger
from toughradius.manage import models
from toughradius.manage.settings import * 
from twisted.internet.threads import deferToThread
from toughradius import __version__ as sys_version
from twisted.internet import defer



def log_query(log_name):
    logfile = "/var/toughradius/{0}.log".format(log_name)
    if not os.path.exists(logfile):
        logfile = "/tmp/{0}.log".format(log_name)

    if os.path.exists(logfile):
        with open(logfile) as f:
            f.seek(0, 2)
            if f.tell() > 64 * 1024:
                f.seek(f.tell() - 64 * 1024)
            else:
                f.seek(0)
            return cyclone.escape.xhtml_escape(f.read()).replace('\n', '<br/>')
    else:
        return "logfile %s not exist" % logfile

@permit.route(r"/admin/logger", u"系统日志查询", MenuSys, order=7.0000, is_menu=True)
class LoggerHandler(BaseHandler):
    @authenticated
    def get(self):
        log_name = "radius-manage"
        return self.render("logger.html", log_name=log_name, msg=log_query(log_name), title="%s logging" % log_name)

    @authenticated
    def post(self):
        log_name = self.get_argument("log_name","radius-worker")
        return self.render("logger.html", log_name=log_name, msg=log_query(log_name), title="%s logging" % log_name)

@permit.route(r"/admin/logger/download", u"系统日志下载", MenuSys, order=7.0001, is_menu=False)
class LoggerHandler(BaseHandler):
    @authenticated
    def get(self):
        log_name = self.get_argument("log_name","radius-worker")
        logfile = "/var/toughradius/{0}.log".format(log_name)
        if not os.path.exists(logfile):
            logfile = "/tmp/{0}.log".format(log_name)

        if os.path.exists(logfile):
            with open(logfile) as f:
                self.export_file("%s.log"%log_name,f.read())
        else:
            self.write("logfile %s not exists"  % logfile )
            self.finish()

    def export_file(self, filename, data):
        self.set_header ('Content-Type', 'application/octet-stream')
        self.set_header ('Content-Disposition', 'attachment; filename=' + filename)
        self.write(data)
        self.finish()


