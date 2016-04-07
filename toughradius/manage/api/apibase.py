#!/usr/bin/env python
#coding=utf-8

import json
import time
import traceback
from hashlib import md5
from cyclone.util import ObjectDict
from toughlib import utils, apiutils, dispatch, logger
from toughradius.manage.base import BaseHandler
from toughlib.apiutils import apistatus


class ApiHandler(BaseHandler):

    def check_xsrf_cookie(self):
        pass

    def render_result(self, **result):
        resp = apiutils.make_message(self.settings.config.system.secret, **result)
        if self.settings.debug:
            logger.debug("[api debug] :: %s response body: %s" % (self.request.path, utils.safeunicode(resp)))
        self.write(resp)

    def parse_form_request(self):
        try:
            print self.get_params()
            return apiutils.parse_form_request(self.settings.config.system.secret, self.get_params())
        except Exception as err:
            logger.error(u"api authorize parse error, %s" % utils.safeunicode(traceback.format_exc()))
            raise ValueError(u"Error: %s" % utils.safeunicode(err.message))

    def parse_request(self):
        try:
            return apiutils.parse_request(self.settings.config.system.secret, self.request.body)
        except Exception as err:
            logger.error(u"api authorize parse error, %s" % utils.safeunicode(traceback.format_exc()))
            raise ValueError(u"Error: %s" % utils.safeunicode(err.message))

    def get_current_user(self):
        session_opr = ObjectDict()
        session_opr.username = 'api'
        session_opr.ipaddr = self.request.remote_ip
        session_opr.opr_type = 0
        session_opr.login_time = utils.get_currtime()
        return session_opr


    def _decode_msg(self,err, msg):
        _msg = msg and utils.safeunicode(msg) or ''
        if issubclass(type(err),BaseException):
            return u'{0}, {1}'.format(utils.safeunicode(_msg),utils.safeunicode(err.message))
        else:
            return _msg

    def render_success(self, msg=None, **result):
        self.render_result(code=apistatus.success.code,
            msg=self._decode_msg(None,msg or apistatus.success.msg),**result)

    def render_sign_err(self, err=None, msg=None):
        self.render_result(code=apistatus.sign_err.code,
            msg=self._decode_msg(err,msg or apistatus.sign_err.msg))
 
    def render_parse_err(self, err=None, msg=None):
        self.render_result(code=apistatus.sign_err.code, 
            msg=self._decode_msg(err,msg or apistatus.sign_err.msg))
 
    def render_verify_err(self, err=None,msg=None):
        self.render_result(code=apistatus.verify_err.code, 
            msg=self._decode_msg(err,msg or apistatus.verify_err.msg))
 
    def render_server_err(self,err=None, msg=None):
        self.render_result(code=apistatus.server_err.code, 
            msg=self._decode_msg(err,msg or apistatus.server_err.msg))

    def render_timeout(self,err=None, msg=None):
        self.render_result(code=apistatus.timeout.code, 
            msg=self._decode_msg(err,msg or apistatus.timeout))

    def render_limit_err(self,err=None, msg=None):
        self.render_result(code=apistatus.limit_err.code, 
            msg=self._decode_msg(err,msg or apistatus.limit_err)) 

    def render_unknow(self,err=None, msg=None):
        self.render_result(code=apistatus.unknow.code, 
            msg=self._decode_msg(err,msg or apistatus.unknow))




