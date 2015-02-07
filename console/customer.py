#!/usr/bin/env python
#coding:utf-8
import sys,os
sys.path.insert(0,os.path.split(__file__)[0])
sys.path.insert(0,os.path.abspath(os.path.pardir))
from twisted.python import log
from bottle import request
from bottle import response
from bottle import TEMPLATE_PATH,MakoTemplate
from bottle import run as runserver
from customer.customer import app as mainapp
from base import (
    get_cookie,
    set_cookie,
    get_param_value,
    get_member_by_name,
    get_account_by_number,
    get_online_status
)
from libs import sqla_plugin,utils
from websock import websock
import functools
import models
import base

def init_application(dbconf=None,cusconf=None,secret=None):
    log.startLogging(sys.stdout)  
    base.update_secret(secret)
    utils.update_secret(secret)
    log.msg("start init application...")
    TEMPLATE_PATH.append("./customer/views/")
    ''' install plugins'''
    engine,metadata = models.get_engine(dbconf)
    sqla_pg = sqla_plugin.Plugin(engine,metadata,keyword='db',create=False,commit=False,use_kwargs=False)
    session = sqla_pg.new_session()
    _sys_param_value = functools.partial(get_param_value,session)
    _get_member_by_name = functools.partial(get_member_by_name,session)
    _get_account_by_number = functools.partial(get_account_by_number,session)
    _get_online_status = functools.partial(get_online_status,session)
    MakoTemplate.defaults.update(**dict(
        get_cookie = get_cookie,
        fen2yuan = utils.fen2yuan,
        fmt_second = utils.fmt_second,
        request = request,
        sys_param_value = _sys_param_value,
        system_name = _sys_param_value("2_member_system_name"),
        get_member = _get_member_by_name,
        get_account = _get_account_by_number,
        is_online = _get_online_status
    ))

    websock.connect(
        _sys_param_value('3_radiusd_address'),
        _sys_param_value('4_radiusd_admin_port')
    )
    
    mainapp.install(sqla_pg)


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
    _customer = _config['customer']
    _secret = _config['secret']

    if args.httpport:_customer['httpport'] = args.httpport
    if args.debug:_customer['debug'] = bool(args.debug)

    init_application(dbconf=_database,cusconf=_customer,secret=_secret)
    
    runserver(
        mainapp, host='0.0.0.0', 
        port=_customer['httpport'] ,
        debug=bool(_customer['debug']),
        reloader=bool(_customer['debug']),
        server="twisted"
    )

if __name__ == "__main__":
    main()