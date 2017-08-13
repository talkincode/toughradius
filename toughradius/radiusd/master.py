#!/usr/bin/env python
#coding:utf-8
import os
from gevent.server import DatagramServer
import gevent
import logging


class RudiusServer(DatagramServer):

    def __init__(self, address, config):
        DatagramServer.__init__(self,address)
        self.logger = logging.getLogger(__name__)
        self.config = config
        self.init_adapter()
        self.start()

    def init_adapters(self):
        if self.config.radiusd.adapter == 'rest':
            from toughradius.radiusd.adapters import rest
            self.adapter =  rest


class RudiusAuthServer(RudiusServer):

    def __init__(self, address, config):
        RudiusServer.__init__(self,address,config)

    def handle(self,data, address):
        reply = self.adapter.handleAuth(data,address)
        gevent.spawn(sendReply,self.socket,reply,address)


class RudiusAcctServer(RudiusServer):

    def __init__(self, address, config):
        RudiusServer.__init__(self,address,config)

    def handle(self,data, address):
        reply = self.adapter.handleAcct(data,address)
        gevent.spawn(sendReply,self.socket,reply,address)




















