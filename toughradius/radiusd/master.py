#!/usr/bin/env python
#coding:utf-8
from gevent.server import DatagramServer
from gevent.pool import Pool
from toughradius import settings
from gevent.queue import Queue, Empty, Full
import gevent
import logging

def get_adapter():
    adapter = settings.radiusd['adapter']
    if adapter == 'free':
        from toughradius.radiusd.adapters.free import FreeAdapter
        return FreeAdapter()
    if adapter == 'rest':
        from toughradius.radiusd.adapters.rest import RestAdapter
        return RestAdapter()

class RudiusAuthServer(DatagramServer):

    def __init__(self,):
        DatagramServer.__init__(self,(settings.radiusd['host'], int(settings.radiusd['auth_port'])))
        self.pool = Pool(settings.radiusd['pool_size'])
        self.adapter = get_adapter()
        self.start()


    def handle(self, data, address):
        self.pool.spawn(self.adapter.handleAuth, self.socket, data, address)



class RudiusAcctServer(DatagramServer):

    def __init__(self,):
        DatagramServer.__init__(self,(settings.radiusd['host'], int(settings.radiusd['acct_port'])))
        self.pool = Pool(settings.radiusd['pool_size'])
        self.adapter = get_adapter()
        self.start()

    def handle(self, data, address):
        self.pool.spawn(self.adapter.handleAcct, self.socket, data, address)























