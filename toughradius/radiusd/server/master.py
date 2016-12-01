#!/usr/bin/env python
# coding=utf-8
import msgpack
from txzmq import ZmqEndpoint, ZmqFactory, ZmqREQConnection
from twisted.internet import protocol
from twisted.internet import reactor
from twisted.internet import defer
from toughradius.common import logger


class RADIUSMaster(protocol.DatagramProtocol):
    """ 
     Radius protocol listen on main process, 
     itself does not deal with any business logic, only to do the message routing and forwarding.
     auth_master and acct_master can be run independently of the process to ensure that the 
     authentication and billing business does not affect each other.
     Message through the msgpack binary package forward, to ensure the performance and compatibility.
    """
    def __init__(self, config, service='auth'):
        self.config = config
        self.service = service
        self.zmqreq = ZmqREQConnection(ZmqFactory(), ZmqEndpoint('bind', config.mqproxy[service+'_bind']))
        logger.info("%s master bind @ " % self.zmqreq)

    def datagramReceived(self, datagram, (host, port)):
        message = msgpack.packb([datagram, host, port])
        d = self.zmqreq.sendMsg(message)
        d.addCallback(self.reply)
        d.addErrback(logger.exception)        
        
    def reply(self, result):
        data, host, port = msgpack.unpackb(result[0])
        self.transport.write(data, (host, int(port)))









