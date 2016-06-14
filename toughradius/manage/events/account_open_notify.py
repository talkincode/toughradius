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
            mailto=userinfo.email,
            tplname=self.MAIL_TPLNAME_WITH_PASSWD,
            customer=utils.safestr(userinfo.realname),
            product=utils.safestr(userinfo.product_name),
            username=userinfo.account_number,
            password=userinfo.password,
            expire=userinfo.expire_date,
            service_call=self.get_param_value("toughcloud_service_call", ''),
            service_mail=service_mail,
            nonce=str(int(time.time()))
        )
        params['sign'] = apiutils.make_sign(api_secret.strip(), params.values())
        try:
            resp = yield httpclient.fetch(self.MAIL_APIURL, postdata=urlencode(params))
            logger.info(resp.body)
        except Exception as err:
            logger.exception(err)


def __call__(dbengine=None, mcache=None, **kwargs):
    return AccountOpenNotifyEvent(dbengine=dbengine, mcache=mcache, **kwargs)