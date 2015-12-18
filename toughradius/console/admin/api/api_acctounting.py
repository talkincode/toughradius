#!/usr/bin/env python
# coding=utf-8

from toughradius.common import utils
from toughradius.common.permit import permit
from toughradius.console.admin.api import api_base
from toughradius.console import models


@permit.route(r"/api/acctounting")
class AcctountingHandler(api_base.ApiHandler):
    """ accounting handler"""

    def post(self):
        try:
            req_msg = self.parse_request()
            if 'username' not in req_msg:
                raise ValueError('username is empty')
        except Exception as err:
            self.render_json(code=1, msg=utils.safeunicode(err))
            return

        try:
            username = req_msg['username']
            result = dict(
                code=0,
                msg=u'success',
                username=username
            )

            sign = self.mksign(result.values())
            result['sign'] = sign
            self.render_json(**result)

        except Exception as err:
            self.syslog.error(u"api authorize error %s" % safeunicode(err))
            self.render_json(code=1, msg=u"api error")


