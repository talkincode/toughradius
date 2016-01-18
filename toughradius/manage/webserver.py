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
from toughlib.dbengine import get_engine
from toughlib.permit import permit, load_handlers
from toughradius.manage.settings import *
from toughlib import db_session as session
from toughlib import db_cache as cache
from toughlib import dispatch
from toughlib.db_backup import DBBackup
import toughradius

class WebManageServer(cyclone.web.Application):
    def __init__(self, config=None, dbengine=None, **kwargs):

        self.config = config

        settings = dict(
            cookie_secret="12oETzKXQAGaYdkL5gEmGeJJFuYh7EQnp2XdTP1o/Vo=",
            login_url="/admin/login",
            template_path=os.path.join(os.path.dirname(__file__), "views"),
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
            encoding_errors='replace',
            module_directory="/tmp/admin"
        )

        self.db_engine = dbengine or get_engine(config)
        self.db = scoped_session(sessionmaker(bind=self.db_engine, autocommit=False, autoflush=False))
        self.session_manager = session.SessionManager(settings["cookie_secret"], self.db_engine, 600)
        self.mcache = cache.CacheManager(self.db_engine)
        self.db_backup = DBBackup(models.get_metadata(self.db_engine), excludes=[
            'tr_online','system_session','system_cache','tr_ticket'])

        self.aes = utils.AESCipher(key=self.config.system.secret)

        permit.add_route(cyclone.web.StaticFileHandler,
                         r"/admin/backup/download/(.*)",
                         u"下载数据",
                         MenuSys,
                         handle_params={"path": self.config.database.backup_path},
                         order=1.0405)

        self.init_route()

        event_path = os.path.join(os.path.abspath(os.path.dirname(toughradius.manage.events.__file__)))
        pkg_prefix="toughradius.manage.events"
        self.load_events(event_path,pkg_prefix)

        cyclone.web.Application.__init__(self, permit.all_handlers, **settings)

    def init_route(self):
        handler_path = os.path.join(os.path.abspath(os.path.dirname(__file__)))
        load_handlers(handler_path=handler_path, pkg_prefix="toughradius.manage",
            excludes=['views','webserver','radius'])

        conn = self.db()
        try:
            oprs = conn.query(models.TrOperator)
            for opr in oprs:
                if opr.operator_type > 0:
                    for rule in self.db.query(models.TrOperatorRule).filter_by(operator_name=opr.operator_name):
                        permit.bind_opr(rule.operator_name, rule.rule_path)
                elif opr.operator_type == 0:  # 超级管理员授权所有
                    permit.bind_super(opr.operator_name)
        except Exception as err:
            dispatch.pub(logger.EVENT_ERROR,"init route error , %s" % str(err))
        finally:
            conn.close()

    def load_events(self,event_path=None,pkg_prefix=None):
        _excludes = ['__init__','settings'] 
        evs = set(os.path.splitext(it)[0] for it in os.listdir(event_path))
        evs = [it for it in evs if it not in _excludes]
        for ev in evs:
            try:
                sub_module = os.path.join(event_path, ev)
                if os.path.isdir(sub_module):
                    dispatch.pub(logger.EVENT_INFO,'load sub event %s' % ev)
                    self.load_events(
                        event_path=sub_module,
                        pkg_prefix="{0}.{1}".format(pkg_prefix, ev)
                    )
                _ev = "{0}.{1}".format(pkg_prefix, ev)
                dispatch.pub(logger.EVENT_INFO,'load_event %s' % _ev)
                dispatch.register(importlib.import_module(_ev).instance())
            except Exception as err:
                dispatch.pub(logger.EVENT_EXCEPTION,err)
                dispatch.pub(logger.EVENT_ERROR,"%s, skip event %s.%s" % (str(err),pkg_prefix,ev))
                continue


def run(config, dbengine):
    app = WebManageServer(config, dbengine)
    reactor.listenTCP(int(config.admin.port), app, interface=config.admin.host)

