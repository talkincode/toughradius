# !/usr/bin/env python
# -*- coding:utf-8 -*-

import time
from urllib import urlencode
from cyclone import httpclient
from toughlib import utils, logger
from toughlib import apiutils
from twisted.internet import defer
from toughradius.events.event_basic import BasicEvent
from toughradius.manage.settings import TOUGHCLOUD as toughcloud
from toughradius.common import tools
from toughlib.mail import send_mail as sendmail


class AccountOpenNotifyEvent(BasicEvent):

    def event_account_open(self, userinfo):

        open_notify = u"""尊敬的 %customer% 您好：
        欢迎使用产品 %product%, 您的账号已经开通，账号名是 %username%, 服务截止 %expire%。"""

        ctx = open_notify.replace('%customer%', userinfo.get('realname'))
        ctx = ctx.replace('%product%', userinfo.get('product_name'))
        ctx = ctx.replace('%username%', userinfo.get('account_number'))
        ctx = ctx.replace('%expire%', userinfo.get('expire_date'))
        ctx = ctx.replace('%product%', userinfo.get('product_name'))
        topic = ctx[:ctx.find('\n')]
        smtp_server = self.get_param_value("smtp_server", '127.0.0.1')
        from_addr = self.get_param_value("smtp_from")
        smtp_port = int(self.get_param_value("smtp_port", 25))
        smtp_sender = self.get_param_value("smtp_sender", None)
        smtp_user = self.get_param_value("smtp_user", None)
        smtp_pwd = self.get_param_value("smtp_pwd", None)
        return sendmail(
                server=smtp_server,
                port=smtp_port,
                user=smtp_user,
                password=smtp_pwd,
                from_addr=from_addr, mailto=userinfo.email,
                topic=utils.safeunicode(topic),
                content=utils.safeunicode(ctx),
                tls=False)


def __call__(dbengine=None, mcache=None, **kwargs):
    return AccountOpenNotifyEvent(dbengine=dbengine, mcache=mcache, **kwargs)