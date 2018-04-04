from gevent import monkey; monkey.patch_all(thread=False)
import gevent
from thrift.protocol import TCompactProtocol
from thrift.protocol import TMultiplexedProtocol
from thrift.transport import TSocket
import time
TSocket.socket = gevent.socket
from toughradius.adapters.libthrift import BrasService, AccessService, AccountingService, LoggerService
from toughradius.adapters.libthrift.ttypes import FindBrasRequest, AccessRequest, AccountingRequest, LoggerRequest




start = time.time()
gs = []
for i in range(100):
    transport = TSocket.TSocket(host="192.168.100.178", port=1815)
    protocol = TCompactProtocol.TCompactProtocol(transport)
    brasClient = BrasService.Client(TMultiplexedProtocol.TMultiplexedProtocol(protocol, "brasService"))
    transport.open()
    gs.append(gevent.spawn(brasClient.findBras,FindBrasRequest(nasid="radius-tester",nasip="0.0.0.0")))
gevent.joinall(gs)
end = time.time()
print "total second = %s" % (end-start)
print "%s per second " % ( 100/(end-start) )
