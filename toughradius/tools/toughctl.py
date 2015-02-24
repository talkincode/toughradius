#!/usr/bin/env python
# -*- coding: utf-8 -*-
import sys,os,signal
import tempfile
import time
import argparse,ConfigParser
from toughradius.tools import config as iconfig
from toughradius.tools.shell import shell


def get_service_tac(app):
    return '%s/%s_service.tac'%(tempfile.gettempdir(),app)

def get_service_pid(app):
    return '%s/%s_service.pid'%(tempfile.gettempdir(),app)
    
def _run_daemon(config,app):
    tac = get_service_tac(app)
    pidfile = get_service_pid(app)
    if not os.path.exists(tac):
        with open(tac,'wb') as tfs:
            tfs.write(iconfig.echo_app_tac(app))
    shell.run('twistd -y %s -l %s --pidfile=%s'%(tac,config.get(app,'logfile'),pidfile),
                raise_on_fail=True)

def _kill_daemon(app):
    ''' kill daemons'''
    pidfile = get_service_pid(app)
    if os.path.exists(pidfile):
        os.kill(int(open(pidfile).read()),signal.SIGTERM)

def run_radiusd(config,daemon=False):
    if not daemon:
        from toughradius.radiusd import server 
        server.run(config)
    else:
        _run_daemon(config,'radiusd')


def run_admin(config,daemon=False):
    if not daemon:
        from toughradius.console import admin_app
        admin_app.run(config)
    else:
        _run_daemon(config,'admin')
    

def run_customer(config,daemon=False):
    if not daemon:
        from toughradius.console import customer_app
        customer_app.run(config)
    else:
        _run_daemon(config,'customer')
    

def start_server(config,app):
    apps = app == 'all' and ['radiusd','admin','customer'] or [app]
    for _app in apps:
        shell.debug('start %s'%_app)
        _run_daemon(config,_app)
        time.sleep(0.5)
    
    
def stop_server(app):
    apps = (app == 'all' and ['radiusd','admin','customer'] or [app])
    for _app in apps:
        shell.debug('stop %s'%_app)
        _kill_daemon(_app)
        time.sleep(0.5)
        


def run_secret_update(config,cfgfile):
    from toughradius.tools import secret 
    secret.update(config,cfgfile)
    
def run_initdb(config,level='0'):
    from toughradius.console import models
    if level == '1':
        models.install(config)
    elif level == '2':
        models.install2(config)
    elif level == '3':
        models.update(config)
        
def run_config():
    from toughradius.tools.config import setup_config
    setup_config()
        
def run_dbdict(config):
    from toughradius.tools import dbdictgen 
    dbdictgen.main()   
    
def run_backup(config):
    from toughradius.tools import backupdb 
    backupdb.backup(config)
    
def run_echo_my_cnf():
    from toughradius.tools.config import echo_my_cnf
    print echo_my_cnf()
    
def run_echo_radiusd_cnf():
    from toughradius.tools.config import echo_radiusd_cnf
    print echo_radiusd_cnf()
    
def run_echo_supervisord_cnf():
    from toughradius.tools.config import echo_supervisord_cnf
    print echo_supervisord_cnf()
    
def run_echo_centos7_service():
    from toughradius.tools.config import echo_centos7_service
    print echo_centos7_service()
    
def run():
    parser = argparse.ArgumentParser()
    parser.add_argument('-radiusd','--radiusd', action='store_true',default=False,dest='radiusd',help='run radiusd')
    parser.add_argument('-admin','--admin', action='store_true',default=False,dest='admin',help='run admin')
    parser.add_argument('-customer','--customer', action='store_true',default=False,dest='customer',help='run customer')
    parser.add_argument('-d','--daemon', action='store_true',default=False,dest='daemon',help='daemon mode')
    parser.add_argument('-start','--start', type=str,default=None,dest='start',help='start server all|radiusd|admin|customer')
    parser.add_argument('-stop','--stop', type=str,default=None,dest='stop',help='stop server all|radiusd|admin|customer')
    parser.add_argument('-initdb','--initdb', type=str,default='0',dest='initdb',help='run initdb 1,2,3')
    parser.add_argument('-config','--config', action='store_true',default=False,dest='config',help='setup config')
    parser.add_argument('-echo_my_cnf','--echo_my_cnf', action='store_true',default=False,dest='echo_my_cnf',help='echo my_cnf')
    parser.add_argument('-echo_radiusd_cnf','--echo_radiusd_cnf', action='store_true',default=False,dest='echo_radiusd_cnf',help='echo radiusd_cnf')
    parser.add_argument('-echo_supervisord_cnf','--echo_supervisord_cnf', action='store_true',default=False,dest='echo_supervisord_cnf',help='echo supervisord_cnf')
    parser.add_argument('-echo_centos7_service','--echo_centos7_service', action='store_true',default=False,dest='echo_centos7_service',help='echo centos7_service')
    parser.add_argument('-secret','--secret', action='store_true',default=False,dest='secret',help='secret update')
    parser.add_argument('-backup','--backup', action='store_true',default=False,dest='backup',help='backup database')
    parser.add_argument('-dbdict','--dbdict', action='store_true',default=False,dest='dbdict',help='dbdict gen')
    parser.add_argument('-c','--conf', type=str,default="/etc/radiusd.conf",dest='conf',help='config file')
    args =  parser.parse_args(sys.argv[1:])  
    
    if args.config:
        return run_config()
    
    if args.echo_my_cnf:
        return run_echo_my_cnf()
        
    if args.echo_radiusd_cnf:
        return run_echo_radiusd_cnf()
        
    if args.echo_supervisord_cnf:
        return run_echo_supervisord_cnf()
    
    if args.echo_centos7_service:
        return run_echo_centos7_service()
    
    if args.stop:
        if not args.stop in ('all','radiusd','admin','customer'):
            print 'usage %s --stop [all|radiusd|admin|customer]'%sys.argv[0]
            return
        return stop_server(args.stop)
    
    config = iconfig.find_config(args.conf)
    
    if args.start:
        if not args.start in ('all','radiusd','admin','customer'):
            print 'usage %s --start [all|radiusd|admin|customer]'%sys.argv[0]
            return
        return start_server(config,args.start)

    if args.radiusd:run_radiusd(config,args.daemon)
    elif args.admin:run_admin(config,args.daemon)
    elif args.customer:run_customer(config,args.daemon)
    elif args.secret:run_secret_update(config,args.conf)
    elif args.initdb:run_initdb(config,args.initdb)
    elif args.dbdict:run_dbdict(config)
    elif args.backup:run_backup(config)
    else: print 'do nothing'
    
        

    
    
    


