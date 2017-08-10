#!/usr/bin/env python
#coding:utf-8

import gevent
import logging
import six
from toughradius.txradius.radius import dictionary
from toughradius.common import cache
from toughradius.txradius import message
from toughradius.txradius.radius import packet
import importlib

class BasicWorker(object):

    def __init__(self,config):
        self.config = config
        self.logger = logging.getLogger(__name__)
        self.dictionary = dictionary.Dictionary(self.config.radiusd.dictionary)
        self.clients = self.config.clients
        self.modules = self.config.modules
        self.module_cache = cache.Mcache()
        self.cache = cache.Mcache()
        self.pool = ThreadPool(self.config.radiusd.pool_size)       


    def load_auth_modules(self):
        self.logger.info('starting load authentication modules')
        for module_cls in self.config.modules.authentication:
            self.logger.info("enable module %s" % module_cls)
            mod = self.get_module(module_cls)

        self.logger.info('starting load authorization modules')
        for module_cls in self.config.modules.authorization:
            self.logger.info("enable module %s" % module_cls)
            mod = self.get_module(module_cls)

    def load_acct_modules(self):
        self.logger.info('starting load acctounting modules')
        for name in ['parse','start','stop','update','on','off','acct_post']:
            for module_cls in self.config.modules.acctounting[name]:
                self.logger.info("enable module %s" % module_cls)
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


