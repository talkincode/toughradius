#!/usr/bin/env python
#coding:utf-8
import sys,os
import cyclone.web
import logging
from twisted.python import log
from twisted.internet import reactor
from twisted.web import server, wsgi
from twisted.python.logfile import DailyLogFile
from toughradius.console.libs import utils
from toughradius.console.libs.smail import mail
from toughradius.console import models
from toughradius.console.portal import handlers
from toughradius.tools.dbengine import get_engine
from sqlalchemy.orm import scoped_session, sessionmaker
from beaker.cache import CacheManager
from beaker.util import parse_cache_config_options
from mako.lookup import TemplateLookup
import functools
import time

###############################################################################
# portal web application                                                                 
###############################################################################
class Application(cyclone.web.Application):
    def __init__(self,**kwargs):
        _handlers = [
            (r"/", handlers.HomeHandler),
            (r"/login", handlers.LoginHandler),
            (r"/mplogin", handlers.MpLoginHandler),
            (r"/logout", handlers.LogoutHandler),
        ]
        
        server = kwargs.pop("server")
        
        settings = dict(
            cookie_secret="12oETzKXQAGaYdkL5gEmGeJJFuYh7EQnp2XdTP1o/Vo=",
            login_url="/login",
            template_path=os.path.join(os.path.dirname(__file__), "portal/views"),
            static_path=os.path.join(os.path.dirname(__file__), "static"),
            xsrf_cookies=True,
            debug=kwargs.get("debug",False),
            share_secret=server.share_secret,
            ac_addr=(server.ac1[0],int(server.ac1[1]))
        )

        self.db = scoped_session(sessionmaker(bind=server.db_engine, autocommit=False, autoflush=False))

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

        self.logging = logging.getLogger('app:logging')
        
        cyclone.web.Application.__init__(self, _handlers, **settings)

###############################################################################
# portal web server                                                                 
###############################################################################
class PortalServer(object):
    
    def __init__(self,config,db_engine=None,daemon=False):
        self.config = config
        self.db_engine = db_engine
        self.daemon = daemon
        self.init_config()
        self.init_timezone()
        self.web_factory = Application(server=self,debug=self.debug)
        
    def init_config(self):
        self.logfile = self.config.get('portal','logfile')
        self.standalone = self.config.has_option('DEFAULT','standalone') and \
            self.config.getboolean('DEFAULT','standalone') or False
        self.secret = self.config.get('DEFAULT','secret')
        self.timezone = self.config.has_option('DEFAULT','tz') and self.config.get('DEFAULT','tz') or "CST-8"
        self.debug = self.config.getboolean('DEFAULT','debug')
        self.port = self.config.getint('portal','port')
        self.host = self.config.has_option('portal','host') \
            and self.config.get('portal','host') or  '0.0.0.0'
        self.share_secret = self.config.get('portal','secret')
        self.ac1 = self.config.get('portal','ac1').split(':')
        self.ac2 = self.config.has_option('portal','ac2') and \
            self.config.get('portal','ac2').split(':') or None
        # update aescipher
        utils.aescipher.setup(self.secret)
        self.encrypt = utils.aescipher.encrypt
        self.decrypt = utils.aescipher.decrypt
        self._check_ssl_config()
    
    def init_timezone(self):
        try:
            os.environ["TZ"] = self.timezone
            time.tzset()
        except:pass
    
    
    def _check_ssl_config(self):
        self.use_ssl = False
        self.privatekey = None
        self.certificate = None
        if self.config.has_option('DEFAULT','ssl') and self.config.getboolean('DEFAULT','ssl'):
            self.privatekey = self.config.get('DEFAULT','privatekey')
            self.certificate = self.config.get('DEFAULT','certificate')
            if os.path.exists(self.privatekey) and os.path.exists(self.certificate):
                self.use_ssl = True

    def run_normal(self):
        if self.debug:
            log.startLogging(sys.stdout)
        else:
            log.startLogging(DailyLogFile.fromFullPath(self.logfile))
        log.msg('portal web server listen %s'%self.host)
        if self.use_ssl:
            log.msg('Portal SSL Enable!')
            from twisted.internet import ssl
            sslContext = ssl.DefaultOpenSSLContextFactory(self.privatekey, self.certificate)
            reactor.listenSSL(
                self.port,
                self.web_factory,
                contextFactory = sslContext,
                interface=self.host
            )
        else:
            reactor.listenTCP(self.port, self.web_factory,interface=self.host)
        if not self.standalone:
            reactor.run()
        
    def get_service(self):
        from twisted.application import service, internet
        if self.use_ssl:
            log.msg('Portal SSL Enable!')
            from twisted.internet import ssl
            sslContext = ssl.DefaultOpenSSLContextFactory(self.privatekey, self.certificate)
            return internet.SSLServer(
                self.port,
                self.web_factory,
                contextFactory = sslContext,
                interface = self.host
            )
        else: 
            log.msg('Portal SSL Disable!')       
            return internet.TCPServer(self.port,self.web_factory,interface = self.host)    
 
def run(config,db_engine=None,is_service=False):
    print 'running portal server...'
    portal = PortalServer(config,db_engine,daemon=is_service)
    if is_service:
        return portal.get_service()
    else:
        portal.run_normal()
            
