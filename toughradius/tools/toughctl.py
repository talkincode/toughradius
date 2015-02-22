#!/usr/bin/env python
# -*- coding: utf-8 -*-
import sys,os
import argparse,ConfigParser

def run_radiusd(config):
    from toughradius.radiusd import server 
    server.run(config)
    
def run_admin(config):
    from toughradius.console import admin_app
    admin_app.run(config)
    
def run_customer(config):
    from toughradius.console import customer_app
    customer_app.run(config)

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
    
    if not args.conf:
        print 'Please specify a configuration file'
        return
        
    cfgs = [
        args.conf,
        '/etc/radiusd.conf',
        '/var/toughradius/radiusd.conf',
        './radiusd.conf',
        os.path.join(os.getenv("WINDIR"),'radiusd.conf')
    ]
    config = ConfigParser.ConfigParser()
    for c in cfgs:
        if os.path.exists(c):
            config.read(c)
            break

    if args.radiusd:run_radiusd(config)
    elif args.admin:run_admin(config)
    elif args.customer:run_customer(config)
    elif args.secret:run_secret_update(config,args.conf)
    elif args.initdb:run_initdb(config,args.initdb)
    elif args.dbdict:run_dbdict(config)
    elif args.backup:run_backup(config)
    else: print 'do nothing'
    
        

    
    
    


