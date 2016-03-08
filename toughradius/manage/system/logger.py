#!/usr/bin/env python
#coding:utf-8
import os
import cyclone.auth
import cyclone.escape
import cyclone.web
import datetime
import time
import evernote.edam.userstore.constants as UserStoreConstants
import evernote.edam.type.ttypes as Types
import evernote.edam.error.ttypes as Errors
import evernote.edam.notestore.NoteStore as NoteStore
from evernote.api.client import EvernoteClient,Store
from toughradius.manage.base import BaseHandler
from toughlib.permit import permit
from toughlib import utils, logger
from toughradius.manage import models
from toughradius.manage.settings import * 
from twisted.internet.threads import deferToThread
from toughradius import __version__ as sys_version
import psutil

__token__ = "S=s60:U=b95550:E=15aac617869:C=15354b04ae8:P=1cd:A=en-devtoken:V=2:H=bbc10177481aa5c4ca19c848ed2ced74"
__store_url__ = "https://app.yinxiang.com/shard/s60/notestore"
__book_guid__ = "53cf724e-8ba4-4e28-9093-f07f08274792"


def get_uuid():
    fs = '/sys/class/dmi/id/product_uuid'
    if os.path.exists(fs):
        return open("/sys/class/dmi/id/product_uuid").read()
    return 'none'


def create_note(usermail,topic,content, 
        online_count,user_total,cpuuse,memuse):
    note_store = Store(__token__, NoteStore.Client, __store_url__)
    title = u"ToughRADIUS Feedback： %s" % (topic)
    note = Types.Note()
    note.notebookGuid = __book_guid__
    note.title = utils.safestr(title)
    note.content = '<?xml version="1.0" encoding="UTF-8"?>'
    note.content += '<!DOCTYPE en-note SYSTEM ' \
                    '"http://xml.evernote.com/pub/enml2.dtd">'
    note.content += '<en-note>'
    note.content +=  utils.safestr(u"<em> Version: %s</em><br/>" % sys_version)
    note.content +=  utils.safestr(u"<em> UUID: %s</em><br/>" % get_uuid())
    note.content +=  utils.safestr(u"<em> Sendtime: %s</em><br/>" % utils.get_currtime())
    note.content +=  utils.safestr(u"<em> Cpu use: %s</em><br/>" % cpuuse)
    note.content +=  utils.safestr(u"<em> Memary use: %s</em><br/>" % memuse)
    note.content +=  utils.safestr(u"<em> Online count: %s</em><br/>" % online_count)
    note.content +=  utils.safestr(u"<em> User total: %s</em><br/>" % user_total)
    note.content +=  utils.safestr(u"<em> Email: %s</em><br/>" % usermail)
    note.content +=  utils.safestr(content)
    note.content += '</en-note>'
    now = int(round(time.time() * 1000)) 
    then = now + 1200000
    note.attributes = Types.NoteAttributes()
    note.attributes.reminderOrder = now
    note.attributes.reminderTime = then
    created_note = note_store.createNote(note)
    return 'note guid: %s' % created_note.guid

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

    def warp_content(self,title,content):
        return u'<h2>%s</h2><code>%s</code><br/>' % (title,content)

    @cyclone.web.authenticated
    def post(self):

        if FeedbackHandler.last_send == 0:
            FeedbackHandler.last_send = time.time()
        elif (time.time() - FeedbackHandler.last_send) < 60:
            rsec = int(60 - (time.time() - FeedbackHandler.last_send))
            return self.render_json(code=0,msg=u"发送间隔为60秒，请再等待 %s 秒"% rsec)

        topic = self.get_argument("topic","")
        email = self.get_argument("email","")

        if not topic or len(topic.strip()) == 0:
            return self.render_json(code=0,msg=u"描述不能为空")
        if len(topic.strip()) > 90:
            return self.render_json(code=0,msg=u"描述不能大于90个字符")

        online_count = self.db.query(models.TrOnline.id).count()
        user_total = self.db.query(models.TrAccount.account_number).filter_by(status=1).count()
        _cpuuse = psutil.cpu_percent(interval=None, percpu=True)
        _memuse = psutil.virtual_memory()
        cpuuse = '; '.join([ 'cpu%s: %s/%%'%(_cpuuse.index(c),c)  for c in _cpuuse])
        memuse = "%s%%; %sMB/%sMB" % (_memuse.percent,
            int((_memuse.total-_memuse.available)/1024.0/1024.0),
            int(_memuse.total/1024.0/1024.0))
        manage_content = self.warp_content('radius-manage',log_query("radius-manage"))
        radius_content = self.warp_content('radius-worker',log_query("radius-worker"))
        task_content = self.warp_content('radius-task',log_query("radius-task"))
        content = u'%s %s %s' % (manage_content,radius_content,task_content)
        deferd = deferToThread(create_note,email,topic,content,
                                online_count,user_total,cpuuse,memuse)
        deferd.addCallbacks(logger.info,logger.error)
        return self.render_json(code=0,msg=u"感谢您的反馈")








