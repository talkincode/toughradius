#!/usr/bin/env python
#coding=utf-8

from toughlib import utils, apiutils
from toughlib.permit import permit
from toughradius.manage.api.apibase import ApiHandler
from toughradius import models
from toughradius.radius.radius_authorize import RadiusAuth

@permit.route(r"/api/v1/authorize")
class AuthorizeHandler(ApiHandler):

    def post(self):
        try:
            req_msg = self.parse_request()
            app = self.application
            auth = RadiusAuth(app.db_engine,app.mcache,app.aes,req_msg)
            self.render_result(**auth.authorize())
        except Exception as err:
            return self.render_result(code=1,msg=utils.safeunicode(err.message))
            
