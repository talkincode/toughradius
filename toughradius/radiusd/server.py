#!/usr/bin/env python
#coding:utf-8
import os
from gevent.server import DatagramServer
from gevent.threadpool import ThreadPool
from gevent import socket
from toughradius.txradius.radius import dictionary
from toughradius.common import cache
import gevent
from . import authenticate
from . import accounting
import click
import logging.config 


class RudiusServer(DatagramServer):

    def __init__(self, address, config):
        DatagramServer.__init__(self,address)
        self.config = config
        self.dictionary = dictionary.Dictionary(self.config.radiusd.dictionary)
        self.clients = self.config.clients
        self.modules = self.config.modules
        self.address = (self.config.radiusd.host,self.config.radiusd.auth_port)
        self.module_cache = cache.Mcache()
        self.cache = cache.Mcache()
        self.pool = ThreadPool(self.config.radiusd.pool_size)
        self.start()
        self.socket.setsockopt(socket.SOL_SOCKET,socket.SO_RCVBUF,10240000)

    def get_module(self,module_class):
        if module_class in self.module_cache:
            return self.server.module_cache[module_class]
        try:
            mdobj = importlib.import_module(module_class)
            if hasattr(robj, 'handle_radius'):
                self.server.module_cache[module_class] = mdobj
                return mdobj
        except:
            logging.exception("import module <%s> error" % module_class)



class RudiusAuthServer(RudiusServer):

    def __init__(self, address, config):
        RudiusServer.__init__(self,address,config)
        self.radius_handler = authenticate.Handler(self)

    def handle(self,data, address):
        self.pool.spawn(self.radius_handler.handle,data,address)


class RudiusAcctServer(RudiusServer):

    def __init__(self, address, config):
        RudiusServer.__init__(self,address,config)
        self.radius_handler = accounting.Handler(self)

    def handle(self,data, address):
        self.pool.spawn(self.radius_handler.handle,data,address)

@click.command()
@click.option('-c','--conf', default='/etc/toughradius/radiusd.json', help='toughradius config file')
@click.option('-d','--debug', is_flag=True)
@click.option('-auth-port','--auth-port', default=0,type=click.INT,help='auth port')
@click.option('-p','--pool-size', default=0,type=click.INT)
def auth(conf,debug,auth_port,pool_size):
    try:
        os.environ['CONFDIR'] = os.path.dirname(conf)
        from toughradius.common import config as iconfig
        config = iconfig.find_config(conf)

        logging.config.dictConfig(config.logger)

        if debug:
            config.radiusd['debug'] = True
        if auth_port > 0:
            config.radiusd['auth_port'] = auth_port
        if pool_size > 0:
            config.radiusd['pool_size'] = pool_size

        address = (config.radiusd.host,config.radiusd.port)
        server = RudiusAuthServer(address, config)
        gevent.run()
    except:
        import traceback
        traceback.print_exc()


@click.command()
@click.option('-c','--conf', default='/etc/toughradius/radiusd.json', help='toughradius config file')
@click.option('-d','--debug', is_flag=True)
@click.option('-acct-port','--acct-port', default=0,type=click.INT,help='acct port')
@click.option('-p','--pool-size', default=0,type=click.INT)
def acct(conf,debug,acct_port,pool_size):
    try:
        os.environ['CONFDIR'] = os.path.dirname(conf)
        from toughradius.common import config as iconfig
        config = iconfig.find_config(conf)

        logging.config.dictConfig(config.logger)

        if debug:
            config.radiusd['debug'] = True
        if acct_port > 0:
            config.radiusd['acct_port'] = acct_port
        if pool_size > 0:
            config.radiusd['pool_size'] = pool_size

        address = (config.radiusd.host,config.radiusd.port)
        server = RudiusAcctServer(address, config)
        gevent.run()
    except:
        import traceback
        traceback.print_exc()

















