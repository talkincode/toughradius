#!/usr/bin/env python
#coding=utf-8
import time
import traceback
from toughlib import utils, apiutils
from toughlib.permit import permit
from toughradius.manage.api.apibase import ApiHandler
from toughradius.manage import models


@permit.route(r"/api/radagent/fetch")
class AgentFetchHandler(ApiHandler):

    def get(self):
        self.post()

    def post(self):
        secret = self.settings.config.system.secret
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

            _host = self.request.host

            api_addr = "{0}://{1}".format(self.request.protocol, _host)

            agent_addr = ':' in _host and _host[:_host.index(':')] or _host
            
            result = {
                'code'          : 0,
                'msg'           : 'ok',
                'api_auth_url'  : "{0}/api/authorize".format(api_addr),
                'api_acct_url'  : "{0}/api/acctounting".format(api_addr),
                'protocol'      : radius_agent_protocol,   
                'auth_endpoints': ",".join([ a.endpoint.replace('*', agent_addr) for a in auth_agents]),    
                'acct_endpoints': ",".join([ a.endpoint.replace('*', agent_addr) for a in acct_agents]), 
                'nonce'         : str(int(time.time())),
            }

            self.render_result(**result)
        except Exception as err:
            self.syslog.error(u"api fetch radagent error, %s" % utils.safeunicode(traceback.format_exc()))
            self.render_result(code=1, msg="api error")



