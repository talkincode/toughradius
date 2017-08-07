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
import logging
import logging.config 
import importlib



class RudiusServer(DatagramServer):

    def __init__(self, address, config):
        DatagramServer.__init__(self,address)
        self.logger = logging.getLogger(__name__)
        self.config = config
        self.dictionary = dictionary.Dictionary(self.config.radiusd.dictionary)
        self.clients = self.config.clients
        self.modules = self.config.modules
        self.address = (self.config.radiusd.host,self.config.radiusd.auth_port)
        self.module_cache = cache.Mcache()
        self.cache = cache.Mcache()
        self.load_modules()
        self.pool = ThreadPool(self.config.radiusd.pool_size)
        self.start()

    def load_modules(self):
        self.logger.info('starting load authentication modules')
        for module_cls in self.config.modules.authentication:
            self.logger.info("load module %s" % module_cls)
            mod = self.get_module(module_cls)

        self.logger.info('starting load authorization modules')
        for module_cls in self.config.modules.authorization:
            self.logger.info("load module %s" % module_cls)
            mod = self.get_module(module_cls)

        self.logger.info('starting load acctounting modules')
        for name in ['parse','start','stop','update','on','off','acct_post']:
            for module_cls in self.config.modules.acctounting[name]:
                self.logger.info("load module %s" % module_cls)
                mod = self.get_module(module_cls)        
    

    def get_module(self,module_class):
        modobj = self.module_cache.get(module_class)
        if modobj:
            return modobj
        try:
            mdobj = importlib.import_module(module_class)
            if hasattr(mdobj, 'handle_radius'):
                self.module_cache.set(module_class,mdobj)
                return mdobj
        except:
            self.logger.exception("load module <%s> error" % module_class)



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
        logging.info(server)
        server.serve_forever()
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
        server.serve_forever()
    except:
        import traceback
        traceback.print_exc()

















