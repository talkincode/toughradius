#!/usr/bin/env python
#coding=utf-8

#!/usr/bin/env python
# coding:utf-8
import json
import time
import traceback
from hashlib import md5
from toughlib import utils, apiutils
from toughradius.manage.base import BaseHandler


class ApiHandler(BaseHandler):

    def check_xsrf_cookie(self):
        pass

    def render_result(self, **result):
        resp = apiutils.make_message(self.settings.config.system.secret, **result)
        if self.settings.debug:
            self.syslog.debug("[api debug] :: %s response body: %s" % (self.request.path, utils.safeunicode(resp)))
        self.write(resp)

    def parse_request(self):
        try:
            return apiutils.parse_request(self.settings.config.system.secret, self.request.body)
        except Exception as err:
            self.syslog.error(u"api authorize parse error, %s" % utils.safeunicode(traceback.format_exc()))
            raise ValueError(u"parse params error")


