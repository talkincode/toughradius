#!/usr/bin/env python
#coding=utf-8
import time
import traceback
from toughradius.common import utils
from toughradius.common.permit import permit
from toughradius.manage.api.apibase import ApiHandler
from toughradius import models


@permit.route(r"/api/v1/nas/delete")
class NasDeleteHandler(ApiHandler):

    def get(self):
        self.post()

    def post(self):
        try:
            req_msg = self.parse_form_request()
            if 'ip_addr' not in req_msg:
                return self.render_verify_err(u"nas ip_addr is empty")
        except Exception as err:
            self.render_result(code=1, msg=utils.safeunicode(err.message))
            return

        try:
            ip_addr = req_msg['ip_addr']
            self.db.query(models.TrBas).filter_by(ip_addr=ip_addr).delete()
            self.db.commit()
            self.render_success(msg=u'API delete bas:%s success' % ip_addr)
        except Exception as err:
            self.logger.error(u"api delete nas error, %s" % utils.safeunicode(traceback.format_exc()))
            self.render_result(code=1, msg=u"api error")



