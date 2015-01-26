#!/usr/bin/env python
#coding:utf-8
import sys,os
sys.path.insert(0,os.path.split(__file__)[0])
sys.path.insert(0,os.path.abspath(os.path.pardir))
from bottle import TEMPLATE_PATH,MakoTemplate
from bottle import run as runserver
from admin.admin import app as mainapp
from admin.ops import app as ops_app
from admin.business import app as bus_app
from base import *
from libs import sqla_plugin
from websock import websock
import functools
import models

def init_application(dbconf=None,consconf=None):
    log.startLogging(sys.stdout)  
    TEMPLATE_PATH.append("./admin/views/")
    ''' install plugins'''
    engine,metadata = models.get_engine(dbconf)
    sqla_pg = sqla_plugin.Plugin(engine,metadata,keyword='db',create=False,commit=False,use_kwargs=False)
    _sys_param_value = functools.partial(get_param_value,sqla_pg.new_session())
    MakoTemplate.defaults.update(**dict(
        get_cookie = get_cookie,
        fen2yuan = utils.fen2yuan,
        fmt_second = utils.fmt_second,
        request = request,
        sys_param_value = _sys_param_value,
        system_name = _sys_param_value("1_system_name"),
        radaddr = _sys_param_value('3_radiusd_address'),
        adminport = _sys_param_value('4_radiusd_admin_port')
    ))
    
    # connect radiusd websocket admin port 
    websock.connect(
        MakoTemplate.defaults['radaddr'],
        MakoTemplate.defaults['adminport'],
    )

    mainapp.install(sqla_pg)
    ops_app.install(sqla_pg)
    bus_app.install(sqla_pg)

    mainapp.mount("/ops",ops_app)
    mainapp.mount("/bus",bus_app)

    #create dir
    try:
        os.makedirs(os.path.join(APP_DIR,'static/xls'))
    except:pass


###############################################################################
# run server                                                                 
###############################################################################

def main():
    import argparse,json
    parser = argparse.ArgumentParser()
    parser.add_argument('-http','--httpport', type=int,default=0,dest='httpport',help='http port')
    parser.add_argument('-raddr','--radaddr', type=str,default=None,dest='radaddr',help='raduis address')
    parser.add_argument('-admin','--adminport', type=int,default=0,dest='adminport',help='admin port')
    parser.add_argument('-d','--debug', nargs='?',type=bool,default=False,dest='debug',help='debug')
    parser.add_argument('-c','--conf', type=str,default="../config.json",dest='conf',help='conf file')
    args =  parser.parse_args(sys.argv[1:])

    if not args.conf or not os.path.exists(args.conf):
        print 'no config file user -c or --conf cfgfile'
        return

    _config = json.loads(open(args.conf).read())
    _database = _config['database']
    _console = _config['console']

    if args.httpport:_console['httpport'] = args.httpport
    if args.radaddr:_console['radaddr'] = args.radaddr
    if args.adminport:_console['adminport'] = args.adminport
    if args.debug:_console['debug'] = bool(args.debug)

    init_application(dbconf=_database,consconf=_console)
    
    runserver(
        mainapp, host='0.0.0.0', 
        port=_console['httpport'] ,
        debug=bool(_console['debug']),
        reloader=bool(_console['debug']),
        server="twisted"
    )

if __name__ == "__main__":
    main()