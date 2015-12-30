#!/usr/bin/env python
#coding=utf-8

from toughlib import utils, apiutils
from toughlib.permit import permit
from toughradius.manage.api.apibase import ApiHandler
from toughradius.manage import models

@permit.route(r"/api/authorize")
class AuthorizeHandler(ApiHandler):

    def post(self):

        @self.cache.cache('get_account_by_username',expire=600)   
        def get_account_by_username(username):
            return self.db.query(models.TrAccount).filter_by(account_number=username).first()

        @self.cache.cache('get_product_by_id',expire=600)   
        def get_product_by_id(product_id):
            return self.db.query(models.TrProduct).filter_by(id=product_id).first()

        try:
            req_msg = self.parse_request()
            if 'username' not in req_msg:
                raise ValueError('username is empty')
        except Exception as err:
            self.render_result(msg=utils.safeunicode(err.message))
            return

        try:
            username = req_msg['username']
            account = get_account_by_username(username)
            if not account:
                self.render_result(code=1, msg=u'user  {0} not exists'.format(utils.safeunicode(username)))
                return

            passwd = self.aes.decrypt(account.password)
            product = get_product_by_id(account.product_id)

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
            self.syslog.error(u"api authorize error %s" % utils.safeunicode(err))
            self.render_result(code=1, msg=u"api error")


