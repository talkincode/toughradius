#!/usr/bin/env python
#coding:utf-8
import os
import cyclone.auth
import cyclone.escape
import cyclone.web
import datetime
import time
from toughradius.manage.base import BaseHandler
from toughlib.permit import permit
from toughlib import utils, logger, httpclient
from toughradius.manage import models
from toughradius.manage.settings import * 
from twisted.internet.threads import deferToThread
from toughradius import __version__ as sys_version
from twisted.internet import defer
import psutil


def get_uuid():
    fs = '/sys/class/dmi/id/product_uuid'
    if os.path.exists(fs):
        return open("/sys/class/dmi/id/product_uuid").read()
    return 'none'

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
    @cyclone.web.authenticated
    def get(self):
        log_name = "radius-manage"
        return self.render("logger.html", msg=log_query(log_name), title="%s logging" % log_name)

    @cyclone.web.authenticated
    def post(self):
        log_name = self.get_argument("log_name","radius-worker")
        return self.render("logger.html", msg=log_query(log_name), title="%s logging" % log_name)


@permit.route(r"/admin/feedback")
class FeedbackHandler(BaseHandler):

    last_send = 0

    def warp_content(self):
        warp_log =  u'<h2>{0}</h2><code>{1}</code><br/>'.format
        warp_attr = u"<em> {0}: {1}</em><br/>".format
        online_count = self.db.query(models.TrOnline.id).count()
        user_total = self.db.query(models.TrAccount.account_number).filter_by(status=1).count()
        _cpuuse = psutil.cpu_percent(interval=None, percpu=True)
        _memuse = psutil.virtual_memory()
        cpuuse = '; '.join([ 'cpu%s: %s/%%'%(_cpuuse.index(c),c)  for c in _cpuuse])
        memuse = "%s%%; %sMB/%sMB" % (_memuse.percent,
            int((_memuse.total-_memuse.available)/1024.0/1024.0),
            int(_memuse.total/1024.0/1024.0))

        attr_content = []
        attr_content.append(warp_attr("Version",sys_version))
        attr_content.append(warp_attr("Cpu use",cpuuse))
        attr_content.append(warp_attr("Memary use",memuse))
        attr_content.append(warp_attr("Online count",online_count))
        attr_content.append(warp_attr("User total",user_total))
        attr_content_str = "".join(attr_content)
        manage_content = warp_log('radius-manage',log_query("radius-manage"))
        radius_content = warp_log('radius-worker',log_query("radius-worker"))
        task_content = warp_log('radius-task',log_query("radius-task"))
        content = u'%s %s %s %s' % (attr_content_str, manage_content,radius_content,task_content)
        return content


    @defer.inlineCallbacks
    def post(self):

        # if FeedbackHandler.last_send == 0:
        #     FeedbackHandler.last_send = time.time()
        # elif (time.time() - FeedbackHandler.last_send) < 10:
        #     rsec = int(10 - (time.time() - FeedbackHandler.last_send))
        #     return self.render_json(code=0,msg=u"发送间隔为10秒，请再等待 %s 秒"% rsec)

        topic = self.get_argument("topic","")
        email = self.get_argument("email","")
        
        if not topic or len(topic.strip()) == 0:
            self.render_json(code=0,msg=u"描述不能为空")
            return
        if len(topic.strip()) > 90:
            self.render_json(code=0,msg=u"描述不能大于90个字符")
            return

        service_url = '%s/service/feedback'%self.settings.config.system.service_url
        param_data = dict(
            topic=utils.safestr(topic),
            email=email,
            uuid=get_uuid(),
            license=self.settings.config.system.license,
            content=utils.safestr(self.warp_content())
        )

        resp = yield httpclient.post(service_url.encode('utf-8'), data=param_data)
        content = yield resp.content()
        self.write(content)







