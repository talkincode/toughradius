#!/usr/bin/env python
#coding=utf-8
import time
import traceback
from toughlib import utils, apiutils, dispatch
from toughlib.permit import permit
from toughradius.manage.api.apibase import ApiHandler
from toughradius import models


@permit.route(r"/api/v1/nas/fetch")
class NasFetchHandler(ApiHandler):

    def get(self):
        self.post()

    def post(self):

        @self.cache.cache(expire=60)   
        def get_bas_by_addr(ip_addr):
            return self.db.query(models.TrBas).filter_by(ip_addr=ip_addr).first()

        try:
            req_msg = self.parse_form_request()
            if 'ip_addr' not in req_msg:
                raise ValueError(u"ip_addr is empty")
        except Exception as err:
            self.render_result(code=1, msg=utils.safeunicode(err.message))
            return

        try:
            ip_addr = req_msg['ip_addr']
            nas = get_bas_by_addr(ip_addr)
            if not nas:
                self.render_result(code=1, msg=u'nas {0} not exists'.format(ip_addr))
                return

            api_addr = "{0}://{1}".format(self.request.protocol, self.request.host)
            
            result = {
                'code'          : 0,
                'msg'           : 'ok',
                'ipaddr'        : ip_addr,
                'secret'        : nas.bas_secret,
                'vendor_id'     : nas.vendor_id,
                'coa_port'      : int(nas.coa_port or 3799),
                'nonce'         : str(int(time.time())),
            }

            self.render_result(**result)
        except Exception as err:
            self.logger.error(u"api fetch nas error, %s" % utils.safeunicode(traceback.format_exc()))
            self.render_result(code=1, msg=u"api error")



