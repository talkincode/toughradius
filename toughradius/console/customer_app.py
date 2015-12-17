#!/usr/bin/env python
#coding:utf-8
import sys
import os
import time

import cyclone.web
from twisted.python import log
from twisted.internet import reactor
from beaker.cache import CacheManager
from beaker.util import parse_cache_config_options
from mako.lookup import TemplateLookup
from sqlalchemy.orm import scoped_session, sessionmaker
from toughadmin.common import utils
from toughadmin.common import logger
from toughadmin.console import models
from toughadmin.common.dbengine import get_engine
from toughadmin.common.permit import permit, load_handlers



class Application(cyclone.web.Application):
    def __init__(self, config=None, **kwargs):

        self.config = config

        try:
            if 'TZ' not in os.environ:
                os.environ["TZ"] = config.defaults.tz
            time.tzset()
        except:
            pass

        settings = dict(
            cookie_secret="12oETzKXQAGaYdkL5gEmGeJJFuYh7EQnp2XdTP1o/Vo=",
            login_url="/login",
            template_path=os.path.join(os.path.dirname(__file__), "views"),
            static_path=os.path.join(os.path.dirname(__file__), "static"),
            xsrf_cookies=True,
            config=config,
            debug=self.config.defaults.debug,
            xheaders=True,
        )

        self.cache = CacheManager(**parse_cache_config_options({
            'cache.type': 'file',
            'cache.data_dir': '/tmp/cache/data',
            'cache.lock_dir': '/tmp/cache/lock'
        }))

        self.tp_lookup = TemplateLookup(directories=[settings['template_path']],
                                        default_filters=['decode.utf8'],
                                        input_encoding='utf-8',
                                        output_encoding='utf-8',
                                        encoding_errors='replace',
                                        module_directory="/tmp")

        self.db_engine = get_engine(config)
        self.db = scoped_session(sessionmaker(bind=self.db_engine, autocommit=False, autoflush=False))

        self.syslog = logger.Logger(config)

        aescipher = utils.AESCipher(key=self.config.defaults.secret)
        self.encrypt = aescipher.encrypt
        self.decrypt = aescipher.decrypt


        permit.add_route(cyclone.web.StaticFileHandler,
                         r"/backup/download/(.*)",
                         u"下载数据",
                         u"系统管理",
                         handle_params={"path": self.config.database.backup_path},
                         order=1.0405)

        self.init_route()
        cyclone.web.Application.__init__(self, permit.all_handlers, **settings)

    def init_route(self):
        handler_path = os.path.join(os.path.abspath(os.path.dirname(__file__)), "handlers")
        load_handlers(handler_path=handler_path, pkg_prefix="toughadmin.console.handlers")

        conn = self.db()
        oprs = conn.query(models.TraOperator)
        for opr in oprs:
            if opr.operator_type > 0:
                for rule in self.db.query(models.TraOperatorRule).filter_by(operator_name=opr.operator_name):
                    permit.bind_opr(rule.operator_name, rule.rule_path)
            elif opr.operator_type == 0:  # 超级管理员授权所有
                permit.bind_super(opr.operator_name)


def run(config):
    log.startLogging(sys.stdout)
    log.msg('admin web server listen %s' % config.admin.host)
    app = Application(config)
    reactor.listenTCP(int(config.admin.port), app, interface=config.admin.host)
    reactor.run()

