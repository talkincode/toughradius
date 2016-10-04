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
from toughlib import logger, utils, dispatch
from toughradius.manage import models
from toughradius.manage import base
from toughradius.common import log_trace
from toughlib.dbengine import get_engine
from toughlib.permit import permit, load_events, load_handlers
from txzmq import ZmqEndpoint, ZmqFactory, ZmqSubConnection
from toughradius.manage.settings import *
from toughlib import db_session as session
from toughlib import db_cache as cache
from toughlib import redis_cache
from toughlib import redis_session
from toughlib import dispatch
from toughlib.dbutils import make_db
from toughlib.db_backup import DBBackup
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
            default_filters=['decode.utf8'],
            input_encoding='utf-8',
            output_encoding='utf-8',
            encoding_errors='ignore',
            module_directory="/tmp/toughradius_admin"
        )

        self.db_engine = dbengine or get_engine(config)
        self.db = scoped_session(sessionmaker(bind=self.db_engine, autocommit=False, autoflush=False))
        self.session_manager = session.SessionManager(settings["cookie_secret"], self.db_engine, 600)

        redisconf = redis_conf(config)
        self.session_manager = redis_session.SessionManager(redisconf,settings["cookie_secret"], 600)
        self.mcache = redis_cache.CacheManager(redisconf,cache_name='RadiusManageCache-%s'%os.getpid())
        self.mcache.print_hit_stat(180)
        
        self.db_backup = DBBackup(models.get_metadata(self.db_engine), excludes=[
            'tr_online','system_session','system_cache','tr_ticket','tr_billing','tr_online_stat',
            'tr_flow_stat'
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
        load_handlers(handler_path=os.path.join(os.path.abspath(os.path.dirname(__file__))),
            pkg_prefix="toughradius.manage", excludes=['views','webserver','radius'])

        # app event init
        event_params= dict(dbengine=self.db_engine, mcache=self.mcache, aes=self.aes)
        dispatch.load_events(os.path.join(os.path.abspath(os.path.dirname(toughradius.manage.events.__file__))),
            "toughradius.manage.events",event_params=event_params)

        permit.add_route(cyclone.web.StaticFileHandler, 
                            r"/admin/backup/download/(.*)",
                            u"下载数据",MenuSys, 
                            handle_params={"path": self.config.database.backup_path},
                            order=5.0005)

        cyclone.web.Application.__init__(self, permit.all_handlers, **settings)

def run(config, dbengine,**kwargs):
    app = HttpServer(config, dbengine)
    reactor.listenTCP(int(config.admin.port), app, interface=config.admin.host)

