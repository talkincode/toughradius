#!/usr/bin/env python
#coding:utf-8
import sys,os
from autobahn.twisted import choosereactor
choosereactor.install_optimal_reactor(True)
sys.path.insert(0,os.path.split(__file__)[0])
sys.path.insert(0,os.path.abspath(os.path.pardir))
from twisted.internet import reactor
from bottle import TEMPLATE_PATH,MakoTemplate
from bottle import mako_template as render
from bottle import run as runserver
from admin.admin import app as mainapp
from admin.ops import app as ops_app
from admin.business import app as bus_app
from admin.card import app as card_app
from admin.product import app as product_app
from base import *
from libs import sqla_plugin,utils
from libs.smail import mail
from websock import websock
import bottle
import tasks
import functools
import models
import base
import time

__version__ = 'v0.7'

reactor.suggestThreadPoolSize(30)

subapps = [ops_app,bus_app,card_app,product_app]

def error403(error):
    return render("error",msg=u"Unauthorized access %s"%error.exception)
    
def error404(error):
    return render("error",msg=u"Not found %s"%error.exception)

def error500(error):
    return render("error",msg=u"Server Internal error %s"%error.exception)

def init_application(config):
    log.startLogging(sys.stdout)  
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
    
    log.msg("init template context...")
    MakoTemplate.defaults.update(**dict(
        sys_version = __version__,
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

def main():
    import argparse,ConfigParser,traceback
    parser = argparse.ArgumentParser()
    parser.add_argument('-http','--httpport', type=int,default=0,dest='httpport',help='http port')
    parser.add_argument('-d','--debug', action='store_true',default=False,dest='debug',help='debug')
    parser.add_argument('-c','--conf', type=str,default="../radiusd.conf",dest='conf',help='conf file')
    args =  parser.parse_args(sys.argv[1:])

    if not args.conf or not os.path.exists(args.conf):
        print 'no config file use -c or --conf cfgfile'
        return
        
    # read config file
    config = ConfigParser.ConfigParser()
    config.read(args.conf)
    
    # update aescipher,timezone
    utils.aescipher.setup(config.get('DEFAULT','secret'))
    base.scookie.setup(config.get('DEFAULT','secret'))
    utils.update_tz(config.get('DEFAULT','tz'))

    try:
        init_application(config)
        runserver(
            mainapp, host='0.0.0.0', 
            port=args.httpport or config.getint('admin','port') ,
            debug=config.getboolean('DEFAULT','debug')  ,
            reloader=False,
            server="twisted"
        )
    except:
        log.err()
        
        
if __name__ == "__main__":
    main()