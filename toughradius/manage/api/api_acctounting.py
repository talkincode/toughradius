#!/usr/bin/env python
# coding=utf-8

from toughlib import utils,apiutils
from toughlib.permit import permit
from toughradius.manage.api.apibase import ApiHandler
from toughradius.manage import models


@permit.route(r"/api/acctounting")
class AcctountingHandler(ApiHandler):

    def post(self):
        try:
            req_msg = self.parse_request()
            if 'username' not in req_msg:
                raise ValueError('username is empty')
        except Exception as err:
            self.render_result(code=1, msg=utils.safeunicode(err))
            return

        try:
            username = req_msg['username']
            result = dict(
                code=0,
                msg=u'success',
                username=username
            )
            self.render_result(**result)
        except Exception as err:
            self.syslog.error(u"api authorize error %s" % utils.safeunicode(err))
            self.render_result(code=1, msg=u"api error")


