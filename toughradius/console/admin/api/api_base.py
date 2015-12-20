#!/usr/bin/env python
# coding:utf-8
import json
import time
import traceback
from hashlib import md5
from toughradius.common import utils
from toughradius.console.admin.base import BaseHandler


class ApiHandler(BaseHandler):

    def check_xsrf_cookie(self):
        pass

    def mksign(self, params=[], debug=True):
        _params = [utils.safestr(p) for p in params if p is not None]
        _params.sort()
        _params.insert(0, self.settings.config.defaults.secret)
        strs = ''.join(_params)
        mds = md5(strs.encode()).hexdigest()
        return mds.upper()


    def check_sign(self, msg, debug=True):
        if "sign" not in msg:
            return False
        sign = msg['sign']
        params = [utils.safestr(msg[k]) for k in msg if k != 'sign']
        local_sign = self.mksign(params)
        return sign == local_sign

    def render_result(self, **result):
        if 'code' not in result:
            result["code"] = 0
        if 'nonce' not in result:
            result['nonce' ] = str(int(time.time()))
        result['sign'] = self.mksign(result.values())
        resp = json.dumps(result, ensure_ascii=False)
        if self.settings.debug:
            self.syslog.debug("[api debug] :: %s response body: %s" % (self.request.path, utils.safeunicode(resp)))
        self.write(resp)

    def parse_request(self):
        try:
            # import pdb;pdb.set_trace()
            msg_src = self.request.body
            if self.settings.debug:
                self.syslog.debug(u"[api debug] :: (%s) request body : %s" % (
                    self.request.path, utils.safeunicode(msg_src)))
            req_msg = json.loads(msg_src)
        except Exception as err:
            self.syslog.error(u"api authorize parse error, %s" % utils.safeunicode(traceback.format_exc()))
            raise ValueError(u"parse params error")

        if not self.check_sign(req_msg):
            raise ValueError(u"message sign error")

        return req_msg


    


