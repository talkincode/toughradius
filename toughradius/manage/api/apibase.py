#!/usr/bin/env python
#coding=utf-8

#!/usr/bin/env python
# coding:utf-8
import json
import time
import traceback
from hashlib import md5
from cyclone.util import ObjectDict
from toughlib import utils, apiutils, dispatch, logger
from toughradius.manage.base import BaseHandler


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


