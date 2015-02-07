#!/usr/bin/env python
#coding:utf-8
import sys,os
sys.path.insert(0,os.path.split(__file__)[0])
sys.path.insert(0,os.path.abspath(os.path.pardir))
from twisted.internet import reactor
from bottle import TEMPLATE_PATH,MakoTemplate
from bottle import run as runserver
from admin.admin import app as mainapp
from admin.ops import app as ops_app
from admin.business import app as bus_app
from admin.card import app as card_app
from base import *
from libs import sqla_plugin,utils
from websock import websock
import tasks
import functools
import models
import base

def init_application(dbconf=None,consconf=None,secret=None):
    log.startLogging(sys.stdout)  
    log.msg("start init application...")
    base.update_secret(secret)
    utils.update_secret(secret)
    TEMPLATE_PATH.append("./admin/views/")
    ''' install plugins'''
    log.msg("init plugins..")
    engine,metadata = models.get_engine(dbconf)
    sqla_pg = sqla_plugin.Plugin(engine,metadata,keyword='db',create=False,commit=False,use_kwargs=False)
    session = sqla_pg.new_session()
    _sys_param_value = functools.partial(get_param_value,session)
    _get_product_name = functools.partial(get_product_name,session)
    log.msg("init template context...")
    MakoTemplate.defaults.update(**dict(
        get_cookie = get_cookie,
        fen2yuan = utils.fen2yuan,
        fmt_second = utils.fmt_second,
        currdate = utils.get_currdate,
        request = request,
        sys_param_value = _sys_param_value,
        get_product_name = _get_product_name,
        system_name = _sys_param_value("1_system_name"),
        radaddr = _sys_param_value('3_radiusd_address'),
        adminport = _sys_param_value('4_radiusd_admin_port'),
        permit = permit,
        all_menus = permit.build_menus(order_cats=[u"系统管理",u"营业管理",u"运维管理"])
    ))
    
    # connect radiusd websocket admin port 
    log.msg("init websocket client...")
    wsparam = (MakoTemplate.defaults['radaddr'],MakoTemplate.defaults['adminport'],)
    reactor.callLater(3, websock.connect,*wsparam)
    log.msg("init tasks...")
    reactor.callLater(5, tasks.start_online_stat_job, sqla_pg.new_session)
   
    log.msg("init operator rules...")
    for _super in session.query(models.SlcOperator.operator_name).filter_by(operator_type=0):
        permit.bind_super(_super[0])

    log.msg("install plugins...")
    mainapp.install(sqla_pg)
    ops_app.install(sqla_pg)
    bus_app.install(sqla_pg)
    card_app.install(sqla_pg)

    mainapp.mount("/ops",ops_app)
    mainapp.mount("/bus",bus_app)
    mainapp.mount("/card",card_app)
    
    #create dir
    try:os.makedirs(os.path.join(APP_DIR,'static/xls'))
    except:pass


###############################################################################
# run server                                                                 
###############################################################################

def main():
    import argparse,json
    parser = argparse.ArgumentParser()
    parser.add_argument('-http','--httpport', type=int,default=0,dest='httpport',help='http port')
    parser.add_argument('-d','--debug', nargs='?',type=bool,default=False,dest='debug',help='debug')
    parser.add_argument('-c','--conf', type=str,default="../config.json",dest='conf',help='conf file')
    args =  parser.parse_args(sys.argv[1:])

    if not args.conf or not os.path.exists(args.conf):
        print 'no config file user -c or --conf cfgfile'
        return

    _config = json.loads(open(args.conf).read())
    _database = _config['database']
    _admin = _config['admin']
    _secret = _config['secret']

    if args.httpport:_admin['httpport'] = args.httpport
    if args.debug:_admin['debug'] = bool(args.debug)

    init_application(dbconf=_database,consconf=_admin,secret=_secret)
    
    runserver(
        mainapp, host='0.0.0.0', 
        port=_admin['httpport'] ,
        debug=bool(_admin['debug']),
        reloader=bool(_admin['debug']),
        server="twisted"
    )

if __name__ == "__main__":
    main()