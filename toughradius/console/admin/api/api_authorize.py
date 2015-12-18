#!/usr/bin/env python
#coding=utf-8

from toughradius.common import utils
from toughradius.common.permit import permit
from toughradius.console.admin.api import api_base
from toughradius.console import models

@permit.route(r"/api/authorize")
class AuthorizeHandler(api_base.ApiHandler):
    """ authorize handler"""

    def post(self):
        try:
            req_msg = self.parse_request()
            if 'username' not in req_msg:
                raise ValueError('username is empty')
        except Exception as err:
            self.render_result(msg=utils.safeunicode(err.message))
            return

        try:
            username = req_msg['username']
            account = self.db.query(models.TrAccount).filter_by(account_number=username).first()
            if not account:
                self.render_result(code=1, msg=u'user  {0} not exists'.format(utils.safeunicode(username)))
                return

            passwd = self.aes.decrypt(account.password)
            product = self.db.query(models.TrProduct).filter_by(id=account.product_id).first()


            result = dict(
                code=0,
                msg='success',
                username=username,
                passwd=passwd,
                input_rate=product.input_max_limit,
                output_rate=product.output_max_limit,
                attrs={
                    "Session-Timeout"      : 86400,
                    "Acct-Interim-Interval": 300
                }
            )

            self.render_result(**result)

        except Exception as err:
            self.syslog.error(u"api authorize error %s" % safeunicode(err))
            self.render_result(code=1, msg=u"api error")


