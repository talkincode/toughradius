#!/usr/bin/env python
#coding=utf-8

from toughlib import utils, apiutils
from toughlib.permit import permit
from toughradius.manage.api.apibase import ApiHandler
from toughradius.manage import models
from toughradius.manage.radius.radius_authorize import RadiusAuth

@permit.route(r"/api/authorize")
class AuthorizeHandler(ApiHandler):

    def post(self):
        try:
            req_msg = self.parse_request()
            if 'username' not in req_msg:
                raise ValueError('username is empty')
        except Exception as err:
            return self.render_result(msg=utils.safeunicode(err.message))
            
        self.render_result(**RadiusAuth(self.application.db_engine,
                                        self.application.mcache,
                                        self.application.aes,req_msg).authorize())


