#!/usr/bin/env python
# -*- coding: utf-8 -*-
from autobahn.twisted import choosereactor
choosereactor.install_optimal_reactor(False)
import sys,os,signal
import tempfile
import time
import argparse,ConfigParser
from toughradius.tools import config as iconfig
from toughradius.tools.shell import shell
from toughradius.tools.dbengine import get_engine

def get_service_tac(app):
    return '%s/%s_service.tac'%(tempfile.gettempdir(),app)

def get_service_pid(app):
    return '%s/%s_service.pid'%(tempfile.gettempdir(),app)
    
def get_service_log(config,app):
    if app == 'standalone':
        return config.get('radiusd','logfile')
    return config.get(app,'logfile')
    
def _run_daemon(config,app):
    tac = get_service_tac(app)
    pidfile = get_service_pid(app)
    with open(tac,'wb') as tfs:
        tfs.write(iconfig.echo_app_tac(app))
    shell.run('twistd -y %s -l %s --pidfile=%s'%(tac,get_service_log(config,app),pidfile),
                raise_on_fail=True)

def _kill_daemon(app):
    ''' kill daemons'''
    pidfile = get_service_pid(app)
    if os.path.exists(pidfile):
        os.kill(int(open(pidfile).read()),signal.SIGTERM)

def run_radiusd(config,daemon=False):
    if not daemon:
        from toughradius.radiusd import server 
        server.run(config,db_engine=get_engine(config))
    else:
        _run_daemon(config,'radiusd')


def run_admin(config,daemon=False):
    if not daemon:
        from toughradius.console import admin_app
        admin_app.run(config,db_engine=get_engine(config))
    else:
        _run_daemon(config,'admin')
    

def run_customer(config,daemon=False):
    if not daemon:
        from toughradius.console import customer_app
        customer_app.run(config,db_engine=get_engine(config))
    else:
        _run_daemon(config,'customer')
        
def run_standalone(config,daemon=False):
    from twisted.internet import reactor
    from toughradius.console import admin_app
    from toughradius.console import customer_app
    from toughradius.radiusd import server
    logfile = config.get('radiusd','logfile')
    config.set('DEFAULT','standalone','true')
    config.set('admin','logfile',logfile)
    config.set('customer','logfile',logfile)
    if not daemon:
        db_engine = get_engine(config)
        server.run(config,db_engine,False)
        admin_app.run(config,db_engine,False)
        customer_app.run(config,db_engine,False)
        reactor.run()
    else:
        _run_daemon(config,'standalone')
    

def start_server(config,app):
    apps = app == 'all' and ['radiusd','admin','customer'] or [app]
    for _app in apps:
        shell.info('start %s'%_app)
        _run_daemon(config,_app)
        time.sleep(0.1)
    
    
def stop_server(app):
    apps = (app == 'all' and ['radiusd','admin','customer'] or [app])
    for _app in apps:
        shell.info('stop %s'%_app)
        _kill_daemon(_app)
        time.sleep(0.1)
        
def restart_server(config,app):
    apps = (app == 'all' and ['radiusd','admin','customer'] or [app])
    for _app in apps:
        shell.info('stop %s'%_app)
        _kill_daemon(_app)
        time.sleep(0.1)
        shell.info('start %s'%_app)
        _run_daemon(config,_app)
        time.sleep(0.1)
    

def run_secret_update(config,cfgfile):
    from toughradius.tools import secret 
    secret.update(config,cfgfile)
    
def run_initdb(config):
    from toughradius.console import models
    models.update(get_engine(config))
        
def run_config():
    from toughradius.tools.config import setup_config
    setup_config()
    
def run_echo_radiusd_cnf():
    from toughradius.tools.config import echo_radiusd_cnf
    print echo_radiusd_cnf()
    
def run_execute_sqls(config,sqlstr):
    from toughradius.tools.sqlexec import execute_sqls
    execute_sqls(config,sqlstr)
    
def run_execute_sqlf(config,sqlfile):
    from toughradius.tools.sqlexec import execute_sqlf
    execute_sqlf(config,sqlfile)
    

def run_radius_tester(config):
    from toughradius.tools.radtest import Tester
    Tester(config).start()
        
    
def run():
    parser = argparse.ArgumentParser()
    parser.add_argument('-radiusd','--radiusd', action='store_true',default=False,dest='radiusd',help='run radiusd')
    parser.add_argument('-admin','--admin', action='store_true',default=False,dest='admin',help='run admin')
    parser.add_argument('-customer','--customer', action='store_true',default=False,dest='customer',help='run customer')
    parser.add_argument('-standalone','--standalone', action='store_true',default=False,dest='standalone',help='run standalone')
    parser.add_argument('-d','--daemon', action='store_true',default=False,dest='daemon',help='daemon mode')
    parser.add_argument('-start','--start', type=str,default=None,dest='start',help='start server all|radiusd|admin|customer')
    parser.add_argument('-restart','--restart', type=str,default=None,dest='restart',help='restart server all|radiusd|admin|customer')
    parser.add_argument('-stop','--stop', type=str,default=None,dest='stop',help='stop server all|radiusd|admin|customer')
    parser.add_argument('-initdb','--initdb', action='store_true',default=False,dest='initdb',help='run initdb')
    parser.add_argument('-config','--config', action='store_true',default=False,dest='config',help='setup config')
    parser.add_argument('-echo_radiusd_cnf','--echo_radiusd_cnf', action='store_true',default=False,dest='echo_radiusd_cnf',help='echo radiusd_cnf')
    parser.add_argument('-secret','--secret', action='store_true',default=False,dest='secret',help='secret update')
    parser.add_argument('-sqls','--sqls', type=str,default=None,dest='sqls',help='execute sql string')
    parser.add_argument('-sqlf','--sqlf', type=str,default=None,dest='sqlf',help='execute sql script file')
    parser.add_argument('-debug','--debug', action='store_true',default=False,dest='debug',help='debug option')
    parser.add_argument('-radtest','--radtest', action='store_true',default=False,dest='radtest',help='start radius tester')
    parser.add_argument('-c','--conf', type=str,default="/etc/radiusd.conf",dest='conf',help='config file')
    args =  parser.parse_args(sys.argv[1:])  
    
    if args.config:
        return run_config()
        
    if args.echo_radiusd_cnf:
        return run_echo_radiusd_cnf()
        
    if args.stop:
        if not args.stop in ('all','radiusd','admin','customer','standalone'):
            print 'usage %s --stop [all|radiusd|admin|customer|standalone]'%sys.argv[0]
            return
        return stop_server(args.stop)
    
    config = iconfig.find_config(args.conf)
    
    if args.debug:
        config.set('DEFAULT','debug','true')
        
    if args.radtest:
        run_radius_tester(config) 

    if args.sqls:
        return run_execute_sqls(config,args.sqls)
    
    if args.sqlf:
        return run_execute_sqlf(config,args.sqlf)
        
    if args.start:
        if not args.start in ('all','radiusd','admin','customer','standalone'):
            print 'usage %s --start [all|radiusd|admin|customer|standalone]'%sys.argv[0]
            return
        return start_server(config,args.start)
    
    if args.restart:
        if not args.restart in ('all','radiusd','admin','customer','standalone'):
            print 'usage %s --restart [all|radiusd|admin|customer|standalone]'%sys.argv[0]
            return
        return restart_server(config,args.restart)

    if args.radiusd:run_radiusd(config,args.daemon)
    elif args.admin:run_admin(config,args.daemon)
    elif args.customer:run_customer(config,args.daemon)
    elif args.standalone:run_standalone(config,args.daemon)
    elif args.secret:run_secret_update(config,args.conf)
    elif args.initdb:run_initdb(config)
    else: print 'do nothing'
    
        

    
    
    


