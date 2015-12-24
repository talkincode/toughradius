#!/usr/bin/env python
#coding:utf-8
import sys
import os
import time
import cyclone.web
from twisted.python import log
from twisted.internet import reactor
from mako.lookup import TemplateLookup
from sqlalchemy.orm import scoped_session, sessionmaker
from toughradius.common import utils
from toughradius.common import logger
from toughradius.console import models
from toughradius.common.dbengine import get_engine
from toughradius.common.permit import permit, load_handlers
from toughradius.common.settings import *
from toughradius.common import db_session as session
from toughradius.common import db_cache as cache
from toughradius.console.zagent import authorize, acctounting


class Application(cyclone.web.Application):
    def __init__(self, config=None, **kwargs):

        self.config = config

        self.syslog = logger.Logger(config)

        try:
            if 'TZ' not in os.environ:
                os.environ["TZ"] = self.config.defaults.tz
            time.tzset()
        except:
            pass

        settings = dict(
            cookie_secret="12oETzKXQAGaYdkL5gEmGeJJFuYh7EQnp2XdTP1o/Vo=",
            session_secret="12oETzKXQAGaYdkL5gEmGeJJFuYh7EQnp2XdTP1o/Vo=",
            login_url="/admin/login",
            template_path=os.path.join(os.path.dirname(__file__), "views"),
            static_path=os.path.join(os.path.dirname(__file__), "static"),
            xsrf_cookies=True,
            config=config,
            debug=self.config.defaults.debug,
            xheaders=True,
        )

        self.tp_lookup = TemplateLookup(directories=[settings['template_path']],
                                        default_filters=['decode.utf8'],
                                        input_encoding='utf-8',
                                        output_encoding='utf-8',
                                        encoding_errors='replace',
                                        module_directory="/tmp/admin")

        self.db_engine = get_engine(config)
        self.db = scoped_session(sessionmaker(bind=self.db_engine, autocommit=False, autoflush=False))
        self.session_manager = session.SessionManager(settings["session_secret"], self.db_engine, 600)
        self.mcache = cache.CacheManager(self.db_engine)

        self.aes = utils.AESCipher(key=self.config.defaults.secret)

        permit.add_route(cyclone.web.StaticFileHandler,
                         r"/backup/download/(.*)",
                         u"下载数据",
                         MenuSys,
                         handle_params={"path": self.config.database.backup_path},
                         order=1.0405)

        self.init_route()

        self.zauth_agent = authorize.ZAuthAgent(self)
        self.zacct_agent = acctounting.ZAcctAgent(self)

        cyclone.web.Application.__init__(self, permit.all_handlers, **settings)

    def init_route(self):
        handler_path = os.path.join(os.path.abspath(os.path.dirname(__file__)), "admin")
        load_handlers(handler_path=handler_path, pkg_prefix="toughradius.console.admin")

        chandler_path = os.path.join(os.path.abspath(os.path.dirname(__file__)), "customer")
        load_handlers(handler_path=chandler_path, pkg_prefix="toughradius.console.customer")

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
            self.syslog.error("init route error , %s" % str(err))
        finally:
            conn.close()


def run(config):
    log.startLogging(sys.stderr)
    log.msg('admin web server listen %s' % config.admin.host)
    app = Application(config)
    reactor.listenTCP(int(config.admin.port), app, interface=config.admin.host)
    reactor.run()

