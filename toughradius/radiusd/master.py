#!/usr/bin/env python
#coding:utf-8
from gevent.server import DatagramServer
import gevent

class RudiusServer(DatagramServer):

    def __init__(self, address, config):
        DatagramServer.__init__(self,address)
        self.config = config
        if self.config.radiusd.adapter == 'rest':
            from toughradius.radiusd.adapters.rest import RestAdapter
            self.adapter =  RestAdapter(self.config)
        elif self.config.radiusd.adapter == 'redis':
            from toughradius.radiusd.adapters.tredis import RedisAdapter
            self.adapter =  RedisAdapter(self.config)
        self.start()
        

class RudiusAuthServer(RudiusServer):

    def __init__(self, address, config):
        RudiusServer.__init__(self,address,config)

    def handle(self,data, address):
        gevent.spawn(self.adapter.handleAuth,self.socket,data,address)


class RudiusAcctServer(RudiusServer):

    def __init__(self, address, config):
        RudiusServer.__init__(self,address,config)

    def handle(self,data, address):
        gevent.spawn( self.adapter.handleAcct, self.socket, data, address)





















