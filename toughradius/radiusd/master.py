#!/usr/bin/env python
#coding:utf-8
from gevent.server import DatagramServer
from gevent.pool import Pool
import logging

logger = logging.getLogger(__name__)

class RudiusAuthServer(DatagramServer):
    """Radius auth server"""

    def __init__(self,adapter, host="0.0.0.0", port=1812, pool_size=32):
        DatagramServer.__init__(self,(host,port))
        self.pool = Pool(pool_size)
        self.adapter = adapter

    def handle(self, data, address):
        if not self.pool.full():
            self.pool.spawn(self.adapter.handleAuth, self.socket, data, address)
        else:
            logger.error("radius auth workpool full")


class RudiusAcctServer(DatagramServer):
    """Radius acct server"""

    def __init__(self,adapter, host="0.0.0.0", port=1813, pool_size=32):
        DatagramServer.__init__(self,(host,port))
        self.pool = Pool(pool_size)
        self.adapter = adapter

    def handle(self, data, address):
        if not self.pool.full():
            self.pool.spawn(self.adapter.handleAcct, self.socket, data, address)
        else:
            logger.error("radius accounting workpool full")























