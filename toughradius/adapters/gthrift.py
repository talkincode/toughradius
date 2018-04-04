#!/usr/bin/env python
# coding:utf-8
import gevent
from .base import BasicAdapter
try:
    from thrift.protocol import TCompactProtocol
    from thrift.protocol import TMultiplexedProtocol
    from thrift.transport import TSocket
    TSocket.socket = gevent.socket
    from toughradius.adapters.libthrift import BrasService, AccessService, AccountingService, LoggerService
    from toughradius.adapters.libthrift.ttypes import FindBrasRequest, AccessRequest, AccountingRequest,LoggerRequest
except Exception as err:
    import traceback
    traceback.print_exc()
    print "need thrift"

class ThriftAdapterError(Exception):
    pass

class ThriftAdapter(BasicAdapter):
    def __init__(self, config):
        BasicAdapter.__init__(self, config)
        self.hostname = self.settings.RADIUSD["hostname"]
        gevent.spawn(self.init_rpc)

    def init_rpc(self):
        transport = TSocket.TSocket(host=self.settings.ADAPTERS['thrift']['host'], port=self.settings.ADAPTERS['thrift']['port'])
        protocol = TCompactProtocol.TCompactProtocol(transport)
        self.brasClient = BrasService.Client(TMultiplexedProtocol.TMultiplexedProtocol(protocol, "brasService"))
        self.accessClient = AccessService.Client(TMultiplexedProtocol.TMultiplexedProtocol(protocol, "accessService"))
        self.accountClient = AccountingService.Client(TMultiplexedProtocol.TMultiplexedProtocol(protocol, "accountingService"))
        self.logClient = LoggerService.Client(TMultiplexedProtocol.TMultiplexedProtocol(protocol, "loggerService"))
        transport.open()

    def getClient(self, nasip=None, nasid=None):
        resp = self.brasClient.findBras(FindBrasRequest(nasip=nasip,nasid=nasid))
        if resp.code > 0:
            raise ThriftAdapterError(resp.message)

        return dict(
            nasid = resp.nasid,
            nasip = resp.nasip,
            vendor = resp.vendor,
            secret = resp.secret,
            status = resp.status,
        )

    def processAuth(self, req):
        username = req.get_user_name()
        req = AccessRequest(**req.dict_message)
        resp = self.accessClient.radiusAccess(req)
        if resp.code > 0:
            raise ThriftAdapterError(resp.message)

        if int(self.settings.RADIUSD.get('ignore_password', 0)) == 0 and resp.check_pwd == 1 :
            if not req.is_valid_pwd(resp.password):
                errstr = 'user password error'
                self.logClient.writeLog(LoggerRequest(host="",submodule="radiusd",level="error",username=username,message=errstr))
                raise ThriftAdapterError(errstr)
        return resp

    def processAcct(self, req):
        billing = req.get_billing()
        billing['nas_paddr'] = req.source[0]
        req = AccountingRequest(**billing)
        resp = self.accountClient.radiusAccounting(req)
        return dict(code=0, msg="ok")



adapter = ThriftAdapter





