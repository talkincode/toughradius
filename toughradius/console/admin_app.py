#!/usr/bin/env python
#coding:utf-8
from autobahn.twisted import choosereactor
choosereactor.install_optimal_reactor(True)
import sys,os
from twisted.internet import reactor
from twisted.python.logfile import DailyLogFile
from bottle import TEMPLATE_PATH,MakoTemplate
from bottle import mako_template as render
from bottle import run as runserver
from toughradius.console.admin.admin import app as mainapp
from toughradius.console.admin.ops import app as ops_app
from toughradius.console.admin.business import app as bus_app
from toughradius.console.admin.card import app as card_app
from toughradius.console.admin.product import app as product_app
from toughradius.console.base import *
from toughradius.console.libs import sqla_plugin,utils
from toughradius.console.libs.smail import mail
from toughradius.console.websock import websock
from toughradius.console import tasks
from toughradius.console import models
from toughradius.console import base
import toughradius
import bottle
import functools
import time

reactor.suggestThreadPoolSize(30)

subapps = [ops_app,bus_app,card_app,product_app]

def error403(error):
    return render("error",msg=u"Unauthorized access %s"%error.exception)
    
def error404(error):
    return render("error",msg=u"Not found %s"%error.exception)

def error500(error):
    return render("error",msg=u"Server Internal error %s"%error.exception)

def init_application(config,use_ssl=False):
    log.msg("start init application...")
    TEMPLATE_PATH.append(os.path.join(os.path.split(__file__)[0],"admin/views/"))
    for _app in [mainapp]+subapps:
        _app.error_handler[403] = error403
        _app.error_handler[404] = error404
        _app.error_handler[500] = error500
        
    log.msg("init plugins..")
    engine,metadata = models.get_engine(config)
    sqla_pg = sqla_plugin.Plugin(engine,metadata,keyword='db',create=False,commit=False,use_kwargs=False)
    session = sqla_pg.new_session()
    _sys_param_value = functools.partial(get_param_value,session)
    _get_product_name = functools.partial(get_product_name,session)
    
    bottle.debug(_sys_param_value('radiusd_address')=='1')
    websock.use_ssl = use_ssl
    
    log.msg("init template context...")
    MakoTemplate.defaults.update(**dict(
        sys_version = toughradius.__version__,
        use_ssl = use_ssl,
        get_cookie = get_cookie,
        fen2yuan = utils.fen2yuan,
        fmt_second = utils.fmt_second,
        currdate = utils.get_currdate,
        bb2mb = utils.bb2mb,
        bbgb2mb = utils.bbgb2mb,
        kb2mb = utils.kb2mb,
        mb2kb = utils.mb2kb,
        sec2hour = utils.sec2hour,
        request = request,
        sys_param_value = _sys_param_value,
        get_product_name = _get_product_name,
        permit = permit,
        all_menus = permit.build_menus(order_cats=[u"系统管理",u"营业管理",u"运维管理"])
    ))
    
    # connect radiusd websocket admin port 
    log.msg("init websocket client...")
    wsparam = (
        _sys_param_value('radiusd_address'),
        _sys_param_value('radiusd_admin_port')
    )
    reactor.callLater(1, websock.connect,*wsparam)
    log.msg("init tasks...")
    reactor.callLater(2, tasks.start_online_stat_job, sqla_pg.new_session)
    reactor.callLater(3, tasks.start_flow_stat_job, sqla_pg.new_session)
    reactor.callLater(4, tasks.start_expire_notify_job, sqla_pg.new_session)
   
    log.msg("init operator rules...")
    for _super in session.query(models.SlcOperator.operator_name).filter_by(operator_type=0):
        permit.bind_super(_super[0])
        
    log.msg("init sendmail..")
    mail.setup(
        server=_sys_param_value('smtp_server'),
        user=_sys_param_value('smtp_user'),
        pwd=_sys_param_value('smtp_pwd'),
        fromaddr=_sys_param_value('smtp_user'),
        sender=_sys_param_value('smtp_sender')
    )

    log.msg("mount app and install plugins...")
    mainapp.install(sqla_pg)
    for _app in subapps:
        _app.install(sqla_pg)
        mainapp.mount(_app.config['__prefix__'],_app)
   

###############################################################################
# run server                                                                 
###############################################################################
from bottle import ServerAdapter
class TwistedService(ServerAdapter):
    def run(self, handler):
        from twisted.web import server, wsgi
        from twisted.python.threadpool import ThreadPool
        from twisted.internet import reactor
        thread_pool = ThreadPool()
        thread_pool.start()
        reactor.addSystemEventTrigger('after', 'shutdown', thread_pool.stop)
        factory = server.Site(wsgi.WSGIResource(reactor, thread_pool, handler))
        reactor.listenTCP(self.port, factory, interface=self.host)


def run(config,is_service=False):
    logfile = config.get('admin','logfile')
    log.startLogging(DailyLogFile.fromFullPath(logfile))
    # update aescipher,timezone
    utils.aescipher.setup(config.get('DEFAULT','secret'))
    base.scookie.setup(config.get('DEFAULT','secret'))
    utils.update_tz(config.get('DEFAULT','tz'))
    use_ssl,privatekey,certificate = utils.check_ssl(config)
    admin_host = config.has_option('admin','host') and config.get('admin','host') or  '0.0.0.0'
    log.msg('server listen %s'%admin_host)
    init_application(config,use_ssl)
    if not is_service:
        runserver(
            mainapp, host=admin_host, 
            port=config.getint('admin','port') ,
            debug=config.getboolean('DEFAULT','debug')  ,
            reloader=False,
            server="twisted"
        )
    else:
        from twisted.web import server, wsgi
        from twisted.python.threadpool import ThreadPool
        from twisted.internet import reactor
        from twisted.application import service, internet
        website = server.Site(wsgi.WSGIResource(reactor, reactor.getThreadPool(), mainapp))
        if use_ssl:
            log.msg('Admin SSL Enable!')
            from twisted.internet import ssl
            sslContext = ssl.DefaultOpenSSLContextFactory(privatekey, certificate)
            return internet.SSLServer(
                config.getint('admin','port'),
                website,
                contextFactory = sslContext,
                interface = admin_host
            )
        else: 
            log.msg('Admin SSL Disable!')       
            return internet.TCPServer(config.getint('admin','port'),website,interface = admin_host)

