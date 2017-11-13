#!/usr/bin/env python
#coding:utf-8
from gevent.server import DatagramServer
from gevent.pool import Pool
from gevent.queue import Queue, Empty, Full
import gevent
import logging


class RudiusServer(DatagramServer):

    def __init__(self, address, config):
        DatagramServer.__init__(self,address)
        self.config = config
        self.pool = Pool(self.config.pool_size)
        if self.config.radiusd.adapter == 'free':
            from toughradius.radiusd.adapters.free import FreeAdapter
            self.adapter = FreeAdapter(self.config)
        if self.config.radiusd.adapter == 'rest':
            from toughradius.radiusd.adapters.rest import RestAdapter
            self.adapter = RestAdapter(self.config)
        self.start()

        

class RudiusAuthServer(RudiusServer):

    def __init__(self, address, config):
        RudiusServer.__init__(self, address, config)

    def handle(self, data, address):
        self.pool.spawn(self.adapter.handleAuth, self.socket, data, address)



class RudiusAcctServer(RudiusServer):

    def __init__(self, address, config):
        RudiusServer.__init__(self,address,config)

    def handle(self, data, address):
        self.pool.spawn(self.adapter.handleAcct, self.socket, data, address)























