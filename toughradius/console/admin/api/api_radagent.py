#!/usr/bin/env python
#coding=utf-8
import time
import traceback
from toughradius.common import utils
from toughradius.common.permit import permit
from toughradius.console.admin.api import api_base
from toughradius.console import models


@permit.route(r"/api/radagent/fetch")
class AgentFetchHandler(api_base.ApiHandler):

    def get(self):
        self.post()

    def post(self):
        try:
            req_msg = self.parse_request()
        except Exception as err:
            self.render_result(code=1, msg=utils.safeunicode(err.message))
            return

        try:
            auth_agents = self.db.query(models.TrRadAgent).filter_by(
                protocol='zeromq',
                radius_type='authorize'
            )

            acct_agents = self.db.query(models.TrRadAgent).filter_by(
                protocol='zeromq',
                radius_type='acctounting'
            )

            radius_agent_protocol = self.get_param_value('radius_agent_protocol', 'http')

            api_addr = "{0}://{1}".format(self.request.protocol, self.request.host)
            
            result = {
                'code'          : 0,
                'msg'           : 'ok',
                'api_auth_url'  : "{0}/api/authorize".format(api_addr),
                'api_acct_url'  : "{0}/api/acctounting".format(api_addr),
                'protocol'      : radius_agent_protocol,   
                'auth_endpoints': ",".join([ a.endpoint.replace('*', self.request.host) for a in auth_agents]),    
                'acct_endpoints': ",".join([ a.endpoint.replace('*', self.request.host) for a in acct_agents]), 
                'nonce'         : str(int(time.time())),
            }

            self.render_result(**result)
        except Exception as err:
            self.syslog.error(u"api fetch radagent error, %s" % utils.safeunicode(traceback.format_exc()))
            self.render_result(code=1, msg=u"api error")



