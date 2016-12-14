#!/usr/bin/env python
#coding:utf-8
import sys
import os
import time
import importlib
import cyclone.web
from twisted.python import log
from twisted.internet import reactor
from mako.lookup import TemplateLookup
from sqlalchemy.orm import scoped_session, sessionmaker
from toughradius.common import logger, utils, dispatch
from toughradius.common import log_trace
from toughradius.common.config import redis_conf
from toughradius.common.dbengine import get_engine
from toughradius.common.permit import permit, load_events, load_handlers
from toughradius.common import db_session as session
from toughradius.common import db_cache as cache
from toughradius.common import redis_cache
from toughradius.common import redis_session
from toughradius.common import dispatch
from toughradius.common.dbutils import make_db
from toughradius.common.db_backup import DBBackup
from toughradius import settings as sysconfig
from toughradius import models
import toughradius

class HttpServer(cyclone.web.Application):

    def __init__(self, config=None, dbengine=None, **kwargs):

        self.config = config

        settings = dict(
            cookie_secret="12oETzKXQAGaYdkL5gEmGeJJFuYh7EQnp2XdTP1o/Vo=",
            login_url="/admin/login",
            template_path=os.path.join(os.path.dirname(toughradius.__file__), "views"),
            static_path=os.path.join(os.path.dirname(toughradius.__file__), "static"),
            xsrf_cookies=True,
            config=self.config,
            debug=self.config.system.debug,
            xheaders=True,
        )

        self.tp_lookup = TemplateLookup(
            directories=[settings['template_path']],
            default_filters=['decode.utf8','h'],
            input_encoding='utf-8',
            output_encoding='utf-8',
            encoding_errors='ignore',
            module_directory="/tmp/toughradius.{}".format(int(time.time()))
        )

        self.db_engine = dbengine or get_engine(config)
        self.db = scoped_session(sessionmaker(bind=self.db_engine, autocommit=False, autoflush=False))
        redisconf = redis_conf(config)
        self.session_manager = redis_session.SessionManager(redisconf,settings["cookie_secret"], 600)
        self.mcache = redis_cache.CacheManager(redisconf,cache_name='RadiusManageCache-%s'%os.getpid())
        
        self.db_backup = DBBackup(models.get_metadata(self.db_engine), excludes=[
            'tr_online','system_session','system_cache','tr_ticket',
            'tr_billing','tr_online_stat','tr_flow_stat'
        ])

        self.aes = utils.AESCipher(key=self.config.system.secret)
        self.logtrace = log_trace.LogTrace(redisconf)

        self.superrpc = None
        if self.config.system.get("superrpc"):
            try:
                import xmlrpclib
                self.superrpc = xmlrpclib.Server(self.config.system.superrpc)
                os.environ['TOUGHEE_SUPER_RPC'] = 'true'
            except:
                logger.error(traceback.format_exc())        

        logger.info("start register httpd events")
        # cache event init
        dispatch.register(self.mcache)
        # logtrace event init
        dispatch.register(self.logtrace,check_exists=True)

        # app init_route
        handler_path = os.path.join(os.path.abspath(os.path.dirname(__file__)),"manage")
        load_handlers(handler_path=handler_path, pkg_prefix="toughradius.manage")

        # app event init
        event_path = os.path.join(os.path.abspath(os.path.dirname(__file__)),"events")
        dispatch.load_events(event_path,"toughradius.events",app=self)

        cyclone.web.Application.__init__(self, permit.all_handlers, **settings)

def run(config, dbengine,**kwargs):
    app = HttpServer(config, dbengine)
    reactor.listenTCP(int(config.admin.port), app, interface=config.admin.host)

