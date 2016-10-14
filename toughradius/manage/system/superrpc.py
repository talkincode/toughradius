#!/usr/bin/env python
# coding:utf-8
import os
from hashlib import md5
import cyclone.auth
import cyclone.escape
import cyclone.web
import traceback
from toughlib import utils,dispatch,logger
from toughradius.manage.base import BaseHandler,authenticated
from toughlib.permit import permit
from toughradius import models
from toughradius.manage.settings import * 
from toughradius.common import tools

if os.environ.get("TOUGHEE_SUPER_RPC") == 'true':

    state_descs = {
        0 : '<span class="label label-warning">未启动</span>',
        10 : '<span class="label label-info">启动中</span>',
        20 : '<span class="label label-success">运行中</span>',
        30 : '<span class="label label-danger">启动失败</span>',
        40 : '<span class="label label-warning">停止中</span>',
        100 : '<span class="label label-danger">已退出</span>',
        200 : '<span class="label label-danger">崩溃中</span>',
        1000 : '<span class="label label-warning">未知</span>',
    }

    name_desc = {
        'auth' : u"认证监听服务",
        'acct' : u"记账监听服务",
        'manage' : u"管理控制台",
        'task' : u"定时任务调度",
        'worker' : u"Radius消息处理器",
        'ssportal' : u"自助服务门户",
        'dbsync' : u"数据同步服务",
        'redis' : u"Redis缓存服务",
        'mongod' : u"Mongodb数据库",
    }

    @permit.route(r"/admin/superrpc", u"系统服务管理", MenuSys, order=9.0000, is_menu=True)
    class SuperProcsHandler(BaseHandler):

        def state_desc(self,statecode):
            return state_descs.get(int(statecode))

        def name_desc(self,name):
            if 'worker' in name:
                name = 'worker'
            return name_desc.get(str(name))

        @authenticated
        def get(self):
            procs = self.superrpc.supervisor.getAllProcessInfo()
            self.render("superrpc.html",procs=procs)

    @permit.route(r"/admin/superrpc/restart", u"服务进程重启", MenuSys, order=9.0001, is_menu=False)
    class SuperProcRestartHandler(BaseHandler):

        @authenticated
        def get(self):
            try:
                name = self.get_argument("name",None)
                if 'worker' in name:
                    name = 'worker:'+name
                ret = self.superrpc.system.multicall(
                    [ {'methodName':'supervisor.stopProcess',
                       'params': [name]},
                      {'methodName':'supervisor.startProcess',
                       'params': [name]},
                      ]
                    )
                logger.info(ret)
            except:
                logger.error(traceback.format_exc())

            self.render_json(code=0,msg=u"重启服务完成")


    @permit.route(r"/admin/superrpc/stop", u"服务进程停止", MenuSys, order=9.0002, is_menu=False)
    class SuperProcStopHandler(BaseHandler):

        @authenticated
        def get(self):
            try:
                name = self.get_argument("name",None)
                if 'worker' in name:
                    name = 'worker:'+name
                ret = self.superrpc.supervisor.stopProcess(name)
                logger.info(ret)
            except:
                logger.error(traceback.format_exc())

            self.render_json(code=0,msg=u"停止服务完成")


    @permit.route(r"/admin/superrpc/restartall", u"重启所有服务", MenuSys, order=9.0003, is_menu=False)
    class SuperProcRestartAllHandler(BaseHandler):

        @authenticated
        def get(self):
            try:
                ret = self.superrpc.supervisor.restart()
                logger.info(ret)
            except:
                logger.error(traceback.format_exc())

            self.render_json(code=0,msg=u"正在重启服务")


    @permit.route(r"/admin/superrpc/reloadconfig", u"重载服务配置", MenuSys, order=9.0004, is_menu=False)
    class SuperProcReloadConfigHandler(BaseHandler):

        @authenticated
        def get(self):
            try:
                ret = self.superrpc.supervisor.reloadConfig()
                logger.info(ret)
            except:
                logger.error(traceback.format_exc())

            self.render_json(code=0,msg=u"正在重载服务配置")


    @permit.route(r"/admin/superrpc/taillog", u"查询服务日志", MenuSys, order=9.0005, is_menu=False)
    class SuperProcTaillogHandler(BaseHandler):

        def log_query(self,logfile):
            if os.path.exists(logfile):
                with open(logfile) as f:
                    f.seek(0, 2)
                    if f.tell() >  16 * 1024:
                        f.seek(f.tell() - 16 * 1024)
                    else:
                        f.seek(0)
                    return cyclone.escape.xhtml_escape(utils.safeunicode(f.read())).replace('\n', '<br/>')
            else:
                return "logfile %s not exist" % logfile

        @authenticated
        def get(self):
            try:
                logfile = self.get_argument("logfile","/var/toughee/manage.log")
                self.render_json(code=0,msg=u"ok",log=self.log_query(logfile))
            except:
                logger.error(traceback.format_exc())
                self.render_json(code=1,msg=u"err",log=u"read logger error:<br><br>%s"%traceback.format_exc())
            

@permit.route(r"/admin/superrpc/logdownload", u"服务日志下载", MenuSys, order=9.0007, is_menu=False)
class LoggerDownloadHandler(BaseHandler):
    @authenticated
    def get(self):
        logfile = self.get_argument("logfile","/var/toughee/manage.log")
        if os.path.exists(logfile):
            with open(logfile) as f:
                self.export_file(os.path.basename(logfile),f.read())
        else:
            self.write("logfile %s not exists"  % logfile )
            self.finish()

    def export_file(self, filename, data):
        self.set_header ('Content-Type', 'application/octet-stream')
        self.set_header ('Content-Disposition', 'attachment; filename=' + filename)
        self.write(data)
        self.finish()







