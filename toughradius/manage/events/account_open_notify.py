# !/usr/bin/env python
# -*- coding:utf-8 -*-

import time
from urllib import urlencode
from cyclone import httpclient
from toughlib import utils, logger
from toughlib import apiutils
from twisted.internet import defer
from toughradius.manage.events.event_basic import BasicEvent
from toughradius.manage.settings import TOUGHCLOUD as toughcloud
from toughradius.common import tools
from toughlib.mail import send_mail as sendmail


class AccountOpenNotifyEvent(BasicEvent):

    """用户开户CLOUD通知服务EVENT"""

    MAIL_TPLNAME = 'tr_open_notify'
    MAIL_TPLNAME_WITH_PASSWD = 'tr_open_notify_withup'
    MAIL_APIURL = "%s/sendmail" % toughcloud.apiurl

    SMS_TPLNAME = 'tr_open_notify'
    SMS_APIURL = "%s/sendsms" % toughcloud.apiurl

    @defer.inlineCallbacks
    def event_toughcloud_sms_account_open(self, userinfo):
        """ toughCloud sms api open notify event """
        if not userinfo:
            return

        api_secret = self.get_param_value("toughcloud_license")
        api_token = yield tools.get_sys_token()
        params = dict(
            token=api_token.strip(),
            action='sms',
            tplname=self.SMS_TPLNAME,
            phone=userinfo.get('phone'),
            customer=utils.safestr(userinfo.get('realname')),
            username=userinfo.get('account_number'),
            product=utils.safestr(userinfo.get('product_name')),
            password=userinfo.get('password'),
            expire=userinfo.get('expire_date'),
            nonce=str(int(time.time()))
        )
        params['sign'] = apiutils.make_sign(api_secret.strip(), params.values())
        try:
            resp = yield httpclient.fetch(self.SMS_APIURL, postdata=urlencode(params))
            logger.info(resp.body)
            logger.info('open account send short message success')
        except Exception as err:
            logger.exception(err)

    @defer.inlineCallbacks
    def event_toughcloud_mail_account_open(self, userinfo):
        """ toughCloud mail api open notify without password event """
        if not userinfo:
            return
        try:
            api_secret = self.get_param_value("toughcloud_license")
            service_mail = self.get_param_value("toughcloud_service_mail")
            if not service_mail:
                return
            api_token = yield tools.get_sys_token()
            params = dict(
                token=api_token.strip(),
                action='email',
                mailto=userinfo.get('email'),
                tplname=self.MAIL_TPLNAME,
                customer=utils.safestr(userinfo.get('realname')),
                username=userinfo.get('account_number'),
                product=utils.safestr(userinfo.get('product_name')),
                expire=userinfo.get('expire_date'),
                service_call=self.get_param_value("toughcloud_service_call", ''),
                service_mail=service_mail,
                nonce=str(int(time.time()))
            )
            params['sign'] = apiutils.make_sign(api_secret.strip(), params.values())
            resp = yield httpclient.fetch(self.MAIL_APIURL, postdata=urlencode(params))
            logger.info(resp.body)
            logger.info('open account send email without password success')
        except Exception as err:
            logger.exception(err)


    @defer.inlineCallbacks
    def event_toughcloud_mail_account_open_wp(self, userinfo):

        """ toughCloud mail api open notify with password event """

        if not userinfo:
            return

        api_secret = self.get_param_value("toughcloud_license")
        service_mail = self.get_param_value("toughcloud_service_mail")
        if not service_mail:
            return
        api_token = yield tools.get_sys_token()
        params = dict(
            token=api_token.strip(),
            action='email',
            mailto=userinfo.get('email'),
            tplname=self.MAIL_TPLNAME_WITH_PASSWD,
            customer=utils.safestr(userinfo.get('realname')),
            username=userinfo.get('account_number'),
            product=utils.safestr(userinfo.get('product_name')),
            password=userinfo.get('password'),
            expire=userinfo.get('expire_date'),
            service_call=self.get_param_value("toughcloud_service_call", ''),
            service_mail=service_mail,
            nonce=str(int(time.time()))
        )
        params['sign'] = apiutils.make_sign(api_secret.strip(), params.values())
        try:
            resp = yield httpclient.fetch(self.MAIL_APIURL, postdata=urlencode(params))
            logger.info(resp.body)
            logger.info('open account send email with password success')
        except Exception as err:
            logger.exception(err)

    def event_smtp_account_open(self, userinfo):

        tr_open_notify = u"""尊敬的 %customer% 您好：
        欢迎使用产品 %product%, 您的账号已经开通，账号名是 %username%, 服务截止 %expire%。
        如有疑问，请联系我们: %service_call%, %service_mail%"""
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