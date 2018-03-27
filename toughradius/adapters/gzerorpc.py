#!/usr/bin/env python
# coding:utf-8
try:
    import zerorpc
except:
    pass
import gevent
from .base import BasicAdapter
from toughradius.pyrad import message
from toughradius.common import tools

class RedisAdapterError(Exception):
    pass

class RedisAdapter(BasicAdapter):
    def __init__(self, config):
        BasicAdapter.__init__(self, config)
        gevent.spawn(self.init_rpc)
        gevent.spawn(self.init_coarpc)

    def init_rpc(self):
        self.zrpc = zerorpc.Client(connect_to=self.settings.ADAPTERS['zerorpc']['connect'])

    def init_coarpc(self):
        self.coazrpc = zerorpc.Server(self.coaservice)
        self.coazrpc.connect(self.settings.ADAPTERS['zerorpc']['coa_bind_connect'])
        self.coazrpc.run()

    def getClient(self, nasip=None, nasid=None):
        return self.zrpc.radius_find_nas(nasip,nasid)

    def processAuth(self, req):
        # check exists
        username = req.get_user_name()
        resp = self.zrpc.radius_authenticat(req.dict_message, async=True).get(timeout=3)
        password = resp.get('password')

        if int(self.settings.RADIUSD.get('ignore_password', 0)) == 0 and password :
            if not req.is_valid_pwd(password):
                errstr = 'user password error'
                self.zrpc.send_radius_userlog(username,errstr, async=True)
                raise RedisAdapterError(errstr)
        return resp

    def processAcct(self, req):
        status_type = req.get_acct_status_type()
        username = req.get_user_name()
        billing = req.get_billing()
        billing['nas_paddr'] = req.source[0]

        if status_type == message.STATUS_TYPE_START:
            self.zrpc.radius_start_accounting(billing, async=True)
            self.logger.info(u'user {0} start accounting'.format(tools.safeunicode(username)))
            return dict(code=0, msg='ok')
        elif status_type == message.STATUS_TYPE_UPDATE:
            self.zrpc.radius_update_accounting(billing, async=True)
            self.logger.info(u'user {0} update accounting'.format(tools.safeunicode(username)))
            return dict(code=0, msg='ok')
        elif status_type == message.STATUS_TYPE_STOP:
            self.zrpc.radius_stop_accounting(billing, async=True)
            self.logger.info(u'user {0} stop accounting'.format(tools.safeunicode(username)))
            return dict(code=0, msg='ok')
        elif status_type in (message.STATUS_TYPE_ACCT_ON, message.STATUS_TYPE_ACCT_OFF):
            nasid, nasip, nas_pub_addr = billing['nas_id'],billing['nas_addr'],billing['nas_paddr']
            self.zrpc.radius_onoff_accounting(nasid, nasip, nas_pub_addr, async=True)
            self.logger.info(u'on/off accounting nasid={0},nasip={1},nas_pub_addr={2}'.format(nasid, nasip, nas_pub_addr))
            return dict(code=0, msg='ok')


adapter = RedisAdapter





