#!/usr/bin/env python
#coding:utf-8
import sys,os
from twisted.python import log
from twisted.internet import reactor
from twisted.web import server, wsgi
from twisted.python.logfile import DailyLogFile
from bottle import request
from bottle import response
from toughradius.console.control.control import app as mainapp
from toughradius.console.libs import mako_plugin,utils
from toughradius.console import base
from toughradius.console.base import *
import toughradius
import functools
import time
import bottle

reactor.suggestThreadPoolSize(30)


###############################################################################
# run server                                                                 
###############################################################################

class ControlServer(object):
    
    def __init__(self,config,daemon=False,app=None,subapps=[]):
        self.config = config
        self.app = app
        self.subapps = subapps
        self.daemon = daemon
        self.viewpath = os.path.join(os.path.split(__file__)[0],"control/views/")
        self.init_config()
        self.init_timezone()
        self.init_application()
        self.init_protocol()
    
    def init_config(self):
        self.logfile = self.config.get('control','logfile')
        self.standalone = self.config.has_option('DEFAULT','standalone') and \
            self.config.getboolean('DEFAULT','standalone') or False
        self.secret = self.config.get('DEFAULT','secret')
        self.timezone = self.config.has_option('DEFAULT','tz') and self.config.get('DEFAULT','tz') or "CST-8"
        self.debug = self.config.getboolean('DEFAULT','debug')
        self.port = self.config.getint('control','port')
        self.host = self.config.has_option('control','host') \
            and self.config.get('control','host') or  '0.0.0.0'
        bottle.debug(self.debug)
        base.scookie.setup(self.secret)
        # update aescipher
        utils.aescipher.setup(self.secret)
        self.encrypt = utils.aescipher.encrypt
        self.decrypt = utils.aescipher.decrypt
        #parse ssl
        self._check_ssl_config()
        
    def _check_ssl_config(self):
        self.use_ssl = False
        self.privatekey = None
        self.certificate = None
        if self.config.has_option('DEFAULT','ssl') and self.config.getboolean('DEFAULT','ssl'):
            self.privatekey = self.config.get('DEFAULT','privatekey')
            self.certificate = self.config.get('DEFAULT','certificate')
            if os.path.exists(self.privatekey) and os.path.exists(self.certificate):
                self.use_ssl = True
                
    def init_timezone(self):
        try:
            os.environ["TZ"] = self.timezone
            time.tzset()
        except:pass

        
    def init_protocol(self):
        self.web_factory = server.Site(wsgi.WSGIResource(reactor, reactor.getThreadPool(), self.app))

            
    def error403(self,error):
        return self.render_plugin.render("error",msg=u"Unauthorized access %s"%error.exception)
    
    def error404(self,error):
        return self.render_plugin.render("error",msg=u"Not found %s"%error.exception)

    def error500(self,error):
        return self.render_plugin.render("error",msg=u"Server Internal error %s"%error.exception)
        
    def init_application(self):
        log.msg("start init application...")
        _lookup = [self.viewpath]
        _context = dict(
            sys_version=toughradius.__version__,
            use_ssl = self.use_ssl,
            get_cookie = get_cookie,
            request = request,
        )
        self.render_plugin = mako_plugin.Plugin('control', _lookup, _context)
        log.msg("mount app and install plugins...")
        for _app in [self.app]+self.subapps:
            for _class in ['DEFAULT', 'database', 'radiusd', 'admin', 'customer','control']:
                for _key, _val in self.config.items(_class):
                    _app.config['%s.%s' % (_class, _key)] = _val
            _app.error_handler[403] = self.error403
            _app.error_handler[404] = self.error404
            _app.error_handler[500] = self.error500
            _app.install(self.render_plugin)
            if _app is not self.app:
                self.app.mount(_app.config['__prefix__'],_app)

    
    def run_normal(self):
        if self.debug:
            log.startLogging(sys.stdout)
        else:
            log.startLogging(DailyLogFile.fromFullPath(self.logfile))
        log.msg('server listen %s'%self.host)
        if self.use_ssl:
            log.msg('Control SSL Enable!')
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
            log.msg('Control SSL Enable!')
            from twisted.internet import ssl
            sslContext = ssl.DefaultOpenSSLContextFactory(self.privatekey, self.certificate)
            return internet.SSLServer(
                self.port,
                self.web_factory,
                contextFactory = sslContext,
                interface = self.host
            )
        else: 
            log.msg('Control SSL Disable!')
            return internet.TCPServer(self.port,self.web_factory,interface = self.host)

def import_sub_app():
    from toughradius.console import control
    for name in control.__all__:
        __import__('toughradius.console.control', globals(), locals(), [name])
    return [getattr(control, name).app for name in control.__all__]

def run(config,is_service=False):
    print 'running control server...'
    subapps = import_sub_app()
    _server = ControlServer(config,daemon=is_service,app=mainapp, subapps=subapps)
    if is_service:
        return _server.get_service()
    else:
        _server.run_normal()
            
