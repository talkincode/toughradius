#!/usr/bin/env python
#coding:utf-8
import os
from gevent.server import DatagramServer
from gevent.threadpool import ThreadPool
from gevent import socket
import gevent
import click
import logging


class RudiusServer(DatagramServer):

    def __init__(self, address, config):
        DatagramServer.__init__(self,address)
        self.logger = logging.getLogger(__name__)
        self.config = config
        self.address = address
        self.pool = ThreadPool(self.config.radiusd.pool_size)
        self.start()


class RudiusAuthServer(RudiusServer):

    def __init__(self, address, config):
        RudiusServer.__init__(self,address,config)

    def handle(self,data, address):
        pass


class RudiusAcctServer(RudiusServer):

    def __init__(self, address, config):
        RudiusServer.__init__(self,address,config)

    def handle(self,data, address):
        pass




















