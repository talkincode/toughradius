#!/usr/bin/env python
#coding=utf-8
import time
import traceback
from toughradius.common import utils
from toughradius.common.permit import permit
from toughradius.console.admin.api import api_base
from toughradius.console import models


@permit.route(r"/api/nas/fetch")
class NasFetchHandler(api_base.ApiHandler):

    def get(self):
        self.post()

    def post(self):
        try:
            req_msg = self.parse_request()
            if 'nasaddr' not in req_msg:
                raise ValueError(u"nasaddr is empty")
        except Exception as err:
            self.render_result(code=1, msg=utils.safeunicode(err.message))
            return

        try:
            nasaddr = req_msg['nasaddr']
            nas = self.db.query(models.TrBas).filter_by(ip_addr=nasaddr).first()
            if not nas:
                self.render_result(code=1, msg=u'nas {0} not exists'.format(nasaddr))
                return

            api_addr = "{0}://{1}".format(self.request.protocol, self.request.host)
            
            result = {
                'code'        : 0,
                'msg'         : 'ok',
                'ipaddr'      : nasaddr,
                'secret'      : nas.bas_secret,
                'vendor_id'   : nas.vendor_id,
                'coa_port'    : int(nas.coa_port or 3799),
                'api_auth_url': "{0}/api/authorize".format(api_addr),
                'api_acct_url': "{0}/api/acctounting".format(api_addr),
                'nonce'       : str(int(time.time())),
            }

            self.render_result(**result)
        except Exception as err:
            self.syslog.error(u"api authorize error, %s" % utils.safeunicode(traceback.format_exc()))
            self.render_result(code=1, msg=u"api error")



