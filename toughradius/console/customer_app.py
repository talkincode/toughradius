#!/usr/bin/env python
#coding:utf-8
import sys,os
from twisted.python import log
from twisted.internet import reactor
from twisted.web import server, wsgi
from twisted.python.logfile import DailyLogFile
from bottle import request
from bottle import response
from toughradius.console.customer.customer import app as mainapp
from toughradius.console.libs import sqla_plugin,mako_plugin,utils
from toughradius.console.libs.smail import mail
from toughradius.console.websock import websock
from toughradius.console import base
from toughradius.console import models
from toughradius.tools.dbengine import get_engine
from toughradius.console.base import (
    get_cookie,
    set_cookie,
    get_param_value,
    get_member_by_name,
    get_account_by_number,
    get_online_status,
    get_product_name,
    Connect,
)
import functools
import time
import bottle

reactor.suggestThreadPoolSize(30)


###############################################################################
# run server                                                                 
###############################################################################

class CustomerServer(object):
    
    def __init__(self,config,db_engine=None,daemon=False,app=None,subapps=[]):
        self.config = config
        self.app = app
        self.subapps = subapps
        self.db_engine = db_engine
        self.daemon = daemon
        self.viewpath = os.path.join(os.path.split(__file__)[0],"customer/views/")
        self.init_config()
        self.init_timezone()
        self.init_db_engine()
        self.init_application()
        self.init_websock()
        self.init_mail()
        self.init_protocol()
    
    def init_config(self):
        self.logfile = self.config.get('customer','logfile')
        self.standalone = self.config.has_option('DEFAULT','standalone') and \
            self.config.getboolean('DEFAULT','standalone') or False
        self.secret = self.config.get('DEFAULT','secret')
        self.timezone = self.config.has_option('DEFAULT','tz') and self.config.get('DEFAULT','tz') or "CST-8"
        self.debug = self.config.getboolean('DEFAULT','debug')
        self.port = self.config.getint('customer','port')
        self.host = self.config.has_option('customer','host') \
            and self.config.get('customer','host') or  '0.0.0.0'
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
    
    def init_db_engine(self):
        if not self.db_engine:
            self.db_engine = get_engine(self.config)
        metadata = models.get_metadata(self.db_engine)
        self.sqla_pg = sqla_plugin.Plugin(
            self.db_engine,
            metadata,
            keyword='db',
            create=False,
            commit=False,
            use_kwargs=False
        )
        
    def init_protocol(self):
        self.web_factory = server.Site(wsgi.WSGIResource(reactor, reactor.getThreadPool(), self.app))
        
    def _sys_param_value(self,pname):
        with Connect(self.sqla_pg.new_session) as db:
            return get_param_value(db,pname)
     
    def _get_product_name(self,pid):
        with Connect(self.sqla_pg.new_session) as db:
            return get_product_name(db,pid)
            
    def _get_member_by_name(self,mname):
        with Connect(self.sqla_pg.new_session) as db:
            return get_member_by_name(db,mname)
    
    def _get_account_by_number(self,account):
        with Connect(self.sqla_pg.new_session) as db:
            return get_account_by_number(db,account)
            
    def _get_online_status(self,account):
        with Connect(self.sqla_pg.new_session) as db:
            return get_online_status(db,account)
            
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
            use_ssl = self.use_ssl,
            get_cookie = get_cookie,
            fen2yuan = utils.fen2yuan,
            fmt_second = utils.fmt_second,
            bb2mb = utils.bb2mb,
            bbgb2mb = utils.bbgb2mb,
            kb2mb = utils.kb2mb,
            mb2kb = utils.mb2kb,
            sec2hour = utils.sec2hour,
            is_expire=utils.is_expire,
            request = request,
            sys_param_value = self._sys_param_value,
            get_member = self._get_member_by_name,
            get_account = self._get_account_by_number,
            is_online = self._get_online_status
        )
        self.render_plugin = mako_plugin.Plugin('customer', _lookup, _context)
        log.msg("mount app and install plugins...")
        for _app in [self.app]+self.subapps:
            for _class in ['DEFAULT', 'database', 'radiusd', 'admin', 'customer','control']:
                for _key, _val in self.config.items(_class):
                    _app.config['%s.%s' % (_class, _key)] = _val
            _app.error_handler[403] = self.error403
            _app.error_handler[404] = self.error404
            _app.error_handler[500] = self.error500
            _app.install(self.sqla_pg)
            _app.install(self.render_plugin)
            if _app is not self.app:
                self.app.mount(_app.config['__prefix__'],_app)
        
    def init_mail(self):
        log.msg("init sendmail..")
        mail.setup(
            server=self._sys_param_value('smtp_server'),
            user=self._sys_param_value('smtp_user'),
            pwd=self._sys_param_value('smtp_pwd'),
            fromaddr=self._sys_param_value('smtp_user'),
            sender=self._sys_param_value('smtp_sender')
        )
        
    def init_websock(self):
        # connect radiusd websocket admin port 
        log.msg("init websocket client...")
        websock.use_ssl = self.use_ssl
        wsparam = (
                self._sys_param_value('radiusd_address'),
                self._sys_param_value('radiusd_admin_port')
        )
        reactor.callLater(1, websock.connect,*wsparam)
    
    def run_normal(self):
        if self.debug:
            log.startLogging(sys.stdout)
        else:
            log.startLogging(DailyLogFile.fromFullPath(self.logfile))
        log.msg('server listen %s'%self.host)
        if self.use_ssl:
            log.msg('Customer SSL Enable!')
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
            log.msg('Admin SSL Enable!')
            from twisted.internet import ssl
            sslContext = ssl.DefaultOpenSSLContextFactory(self.privatekey, self.certificate)
            return internet.SSLServer(
                self.port,
                self.web_factory,
                contextFactory = sslContext,
                interface = self.host
            )
        else: 
            log.msg('Admin SSL Disable!')       
            return internet.TCPServer(self.port,self.web_factory,interface = self.host)

def import_sub_app():
    from toughradius.console import customer
    for name in customer.__all__:
        __import__('toughradius.console.customer', globals(), locals(), [name])
    return [getattr(customer, name).app for name in customer.__all__]

def run(config,db_engine=None,is_service=False):
    print 'running customer server...'
    subapps = import_sub_app()
    admin = CustomerServer(config,db_engine,daemon=is_service,app=mainapp, subapps=subapps)
    if is_service:
        return admin.get_service()
    else:
        admin.run_normal()
            
