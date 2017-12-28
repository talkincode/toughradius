#!/usr/bin/env python
#coding:utf-8
from gevent.server import DatagramServer
from gevent.pool import Pool
import socket
import gevent
import logging

logger = logging.getLogger(__name__)

def setsockopt(_socket):
    try:
        _socket.setsockopt(socket.SOL_SOCKET, socket.SO_SNDBUF, 32 * 1024 * 1024)
        _socket.setsockopt(socket.SOL_SOCKET, socket.SO_RCVBUF, 32 * 1024 * 1024)
    except:
        pass

class RudiusAuthServer(DatagramServer):
    """Radius auth server"""

    def __init__(self,adapter, host="0.0.0.0", port=1812, pool_size=32):
        DatagramServer.__init__(self,(host,port))
        self.pool = Pool(pool_size)
        self.adapter = adapter
        self.init_socket()
        setsockopt(self.socket)


    def handle(self, data, address):
        if not self.pool.full():
            self.pool.spawn(self.adapter.handleAuth, self.socket, data, address)
            gevent.sleep(0)
        else:
            logger.error("radius auth workpool full")


class RudiusAcctServer(DatagramServer):
    """Radius acct server"""

    def __init__(self,adapter, host="0.0.0.0", port=1813, pool_size=32):
        DatagramServer.__init__(self,(host,port))
        self.pool = Pool(pool_size)
        self.adapter = adapter
        self.init_socket()
        setsockopt(self.socket)

    def handle(self, data, address):
        if not self.pool.full():
            self.pool.spawn(self.adapter.handleAcct, self.socket, data, address)
            gevent.sleep(0)
        else:
            logger.error("radius accounting workpool full")























