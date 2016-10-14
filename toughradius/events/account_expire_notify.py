#!/usr/bin/env python
#coding:utf-8
import os
import time
import datetime
from urllib import urlencode
from cyclone import httpclient
from toughlib import utils,dispatch,logger
from toughlib import apiutils
from twisted.internet import reactor,defer
from toughradius.events.event_basic import BasicEvent
from toughradius.manage.settings import TOUGHCLOUD as toughcloud
from toughradius.common import tools
from toughlib.mail import send_mail as sendmail
from email.mime.text import MIMEText
from email import Header
from urllib import quote

class AccountExpireNotifyEvent(BasicEvent):

    def event_webhook_account_expire(self, userinfo):
        """webhook notify event """
        notify_url = self.get_param_value("expire_notify_url")
        if not notify_url:
            return
        url = notify_url.replace('{account}',userinfo.account_number)
        url = url.replace('{customer}',utils.safestr(userinfo.realname))
        url = url.replace('{expire}',userinfo.expire_date)
        url = url.replace('{email}',userinfo.email)
        url = url.replace('{mobile}',userinfo.mobile)
        url = url.replace('{product}',utils.safestr(userinfo.product_name))
        url = url.encode('utf-8')
        url = quote(url,":?=/&")
        return httpclient.fetch(url).addCallbacks(lambda r: logger.info(r.body),logger.exception)


    def event_smtp_account_expire(self, userinfo):
        notify_tpl = self.get_param_value("smtp_notify_tpl")
        ctx = notify_tpl.replace('#account#',userinfo.account_number)
        ctx = ctx.replace('#expire#',userinfo.expire_date)
        topic = ctx[:ctx.find('\n')]
        smtp_server = self.get_param_value("smtp_server",'127.0.0.1')
        from_addr = self.get_param_value("smtp_from")
        smtp_port = int(self.get_param_value("smtp_port",25))
        smtp_sender = self.get_param_value("smtp_sender",None)
        smtp_user = self.get_param_value("smtp_user",None)
        smtp_pwd = self.get_param_value("smtp_pwd",None)
        return  sendmail(
                server=smtp_server, 
                port=smtp_port,
                user=smtp_user, 
                password=smtp_pwd, 
                from_addr=from_addr, mailto=userinfo.email, 
                topic=utils.safeunicode(topic), 
                content=utils.safeunicode(ctx),
                tls=False)


def __call__(dbengine=None, mcache=None, **kwargs):
    return AccountExpireNotifyEvent(dbengine=dbengine, mcache=mcache, **kwargs)
