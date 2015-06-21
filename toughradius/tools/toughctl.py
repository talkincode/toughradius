#!/usr/bin/env python
# -*- coding: utf-8 -*-
from autobahn.twisted import choosereactor
choosereactor.install_optimal_reactor(False)
import argparse,ConfigParser
from toughradius.tools import config as iconfig
from toughradius.tools.shell import shell
from toughradius.tools.dbengine import get_engine
from toughradius.tools import initdb as init_db
import sys,os,signal
import tempfile
import time

def check_env(config):
    """check runtime env"""
    try:
        backup_path = config.get('database','backup_path') 
        if not os.path.exists(backup_path):
            os.makedirs(backup_path)
    except Exception as err:
        shell.err("check_env error %s"%repr(err))
        

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

def run_control(config, daemon=False):
    if not daemon:
        from toughradius.console import control_app
        control_app.run(config)
    else:
        _run_daemon(config, 'control')
        
def run_standalone(config,daemon=False):
    '''
    所有应用在一个进程的运行模式
    '''
    from twisted.internet import reactor
    from toughradius.console import admin_app
    from toughradius.console import customer_app
    from toughradius.console import control_app
    from toughradius.radiusd import server
    logfile = config.get('radiusd','logfile')
    config.set('DEFAULT','standalone','true')
    config.set('admin','logfile',logfile)
    config.set('customer','logfile',logfile)
    config.set('control', 'logfile', logfile)
    if not daemon:
        db_engine = get_engine(config)
        server.run(config,db_engine,False)
        admin_app.run(config,db_engine,False)
        customer_app.run(config,db_engine,False)
        control_app.run(config, False)
        reactor.run()
    else:
        _run_daemon(config,'standalone')
    
    


def start_server(config,app):
    '''
    启动守护进程
    '''
    apps = app == 'all' and ['radiusd','admin','customer','control'] or [app]
    for _app in apps:
        shell.info('start %s'%_app)
        _run_daemon(config,_app)
    
    
def stop_server(app):
    '''
    停止守护进程
    '''
    apps = (app == 'all' and ['radiusd','admin','customer','control'] or [app])
    for _app in apps:
        shell.info('stop %s'%_app)
        _kill_daemon(_app)
        
def restart_server(config,app):
    '''
    重启守护进程
    '''
    apps = (app == 'all' and ['radiusd','admin','customer','control'] or [app])
    for _app in apps:
        shell.info('stop %s'%_app)
        _kill_daemon(_app)
        shell.info('start %s'%_app)
        _run_daemon(config,_app)
    

def run_secret_update(config,cfgfile):
    from toughradius.tools import secret 
    secret.update(config,cfgfile)
    
def run_initdb(config):
    init_db.update(get_engine(config))
        
def run_config():
    from toughradius.tools.config import setup_config
    setup_config()
    
def run_echo_radiusd_cnf():
    from toughradius.tools.config import echo_radiusd_cnf
    print echo_radiusd_cnf()
    
def run_echo_radiusd_script():
    from toughradius.tools import livecd
    print livecd.echo_radiusd_script()
    
def run_echo_mysql_cnf():
    from toughradius.tools import livecd
    print livecd.echo_mysql_cnf()

def run_execute_sqls(config,sqlstr):
    from toughradius.tools.sqlexec import execute_sqls
    execute_sqls(config,sqlstr)
    
def run_execute_sqlf(config,sqlfile):
    from toughradius.tools.sqlexec import execute_sqlf
    execute_sqlf(config,sqlfile)
    
def run_gensql(config):
    from sqlalchemy import create_engine
    def _e(sql, *multiparams, **params): print (sql)
    engine =  create_engine(
            config.get('database',"dburl"),
            strategy = 'mock',
            executor = _e
        )
    from toughradius.console import models
    metadata = models.get_metadata(engine)
    metadata.create_all(engine)
    
def run_dumpdb(config,dumpfs):
    from toughradius.tools import backup
    backup.dumpdb(config,dumpfs)
    
def run_restoredb(config,restorefs):
    from toughradius.tools import backup
    backup.restoredb(config,restorefs)
    
def run_live_system_init():
    if not sys.platform.startswith('linux'):
        return
    from toughradius.tools import livecd
    # create database
    shell.run("echo \"create database toughradius DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;\" | mysql")
    # setup mysql user and passwd
    shell.run("echo \"GRANT ALL ON toughradius.* TO radiusd@'127.0.0.1' IDENTIFIED BY 'root' WITH GRANT OPTION;FLUSH PRIVILEGES;\" | mysql")
    shell.run("mkdir -p /var/toughradius/log")
    shell.run("mkdir -p /var/toughradius/data")
    
    with open("/etc/radiusd.conf",'wb') as ef:
        ef.write(livecd.echo_radiusd_cnf())
        
    with open("/var/toughradius/privkey.pem",'wb') as ef:
        ef.write(livecd.echo_privkey_pem())
        
    with open("/var/toughradius/cacert.pem",'wb') as ef:
        ef.write(livecd.echo_cacert_pem())
        
    shell.run("toughctl --initdb")
    
    if not os.path.exists("/etc/init.d/radiusd"):
        with open("/etc/init.d/radiusd",'wb') as rf:
            rf.write(livecd.echo_radiusd_script())
        shell.run("chmod +x /etc/init.d/radiusd")
        shell.run("update-rc.d radiusd defaults")
        
    shell.run("/etc/init.d/radiusd start",raise_on_fail=True)

def run():
    parser = argparse.ArgumentParser()
    parser.add_argument('-radiusd','--radiusd', action='store_true',default=False,dest='radiusd',help='run radiusd')
    parser.add_argument('-admin','--admin', action='store_true',default=False,dest='admin',help='run admin')
    parser.add_argument('-customer','--customer', action='store_true',default=False,dest='customer',help='run customer')
    parser.add_argument('-control', '--control', action='store_true', default=False, dest='control',help='run control')
    parser.add_argument('-standalone','--standalone', action='store_true',default=False,dest='standalone',help='run standalone')
    parser.add_argument('-d','--daemon', action='store_true',default=False,dest='daemon',help='daemon mode')
    parser.add_argument('-start','--start', type=str,default=None,dest='start',help='start server all|radiusd|admin|customer|control')
    parser.add_argument('-restart','--restart', type=str,default=None,dest='restart',help='restart server all|radiusd|admin|customer|control')
    parser.add_argument('-stop','--stop', type=str,default=None,dest='stop',help='stop server all|radiusd|admin|customer')
    parser.add_argument('-initdb','--initdb', action='store_true',default=False,dest='initdb',help='run initdb')
    parser.add_argument('-dumpdb','--dumpdb', type=str,default=None,dest='dumpdb',help='run dumpdb')
    parser.add_argument('-restoredb','--restoredb', type=str,default=None,dest='restoredb',help='run restoredb')
    parser.add_argument('-echo_radiusd_cnf','--echo_radiusd_cnf', action='store_true',default=False,dest='echo_radiusd_cnf',help='echo radiusd_cnf')
    parser.add_argument('-echo_mysql_cnf','--echo_mysql_cnf', action='store_true',default=False,dest='echo_mysql_cnf',help='echo mysql cnf')    
    parser.add_argument('-echo_radiusd_script','--echo_radiusd_script', action='store_true',default=False,dest='echo_radiusd_script',help='echo radiusd script')
    parser.add_argument('-secret','--secret', action='store_true',default=False,dest='secret',help='secret update')
    parser.add_argument('-sqls','--sqls', type=str,default=None,dest='sqls',help='execute sql string')
    parser.add_argument('-sqlf','--sqlf', type=str,default=None,dest='sqlf',help='execute sql script file')
    parser.add_argument('-gensql','--gensql', action='store_true',default=False,dest='gensql',help='export sql script file')
    parser.add_argument('-debug','--debug', action='store_true',default=False,dest='debug',help='debug option')
    parser.add_argument('-x','--xdebug', action='store_true',default=False,dest='xdebug',help='xdebug option')
    parser.add_argument('-c','--conf', type=str,default="/etc/radiusd.conf",dest='conf',help='config file')
    args =  parser.parse_args(sys.argv[1:])  

        
    if args.echo_radiusd_cnf:
        return run_echo_radiusd_cnf()
        
    if args.echo_radiusd_script:
        return run_echo_radiusd_script()
    
    if args.echo_mysql_cnf:
        return run_echo_mysql_cnf()
        
    if args.stop:
        if not args.stop in ('all','radiusd','admin','customer','control','standalone'):
            print 'usage %s --stop [all|radiusd|admin|customer|control|standalone]'%sys.argv[0]
            return
        return stop_server(args.stop)
    
    config = iconfig.find_config(args.conf)
    
    if not config:
        return run_live_system_init()
        
    check_env(config)
    
    if args.debug or args.xdebug:
        config.set('DEFAULT','debug','true')
        
    if args.gensql:
        return run_gensql(config)
        
    if args.dumpdb:
        return run_dumpdb(config,args.dumpdb)
        
    if args.restoredb:
        return run_restoredb(config,args.restoredb)

    if args.sqls:
        return run_execute_sqls(config,args.sqls)
    
    if args.sqlf:
        return run_execute_sqlf(config,args.sqlf)
        
    if args.start:
        if not args.start in ('all','radiusd','admin','customer','control','standalone'):
            print 'usage %s --start [all|radiusd|admin|customer|control|standalone]'%sys.argv[0]
            return
        return start_server(config,args.start)
    
    if args.restart:
        if not args.restart in ('all','radiusd','admin','customer','control','standalone'):
            print 'usage %s --restart [all|radiusd|admin|customer|control|standalone]'%sys.argv[0]
            return
        return restart_server(config,args.restart)

    if args.radiusd:run_radiusd(config,args.daemon)
    elif args.admin:run_admin(config,args.daemon)
    elif args.customer:run_customer(config,args.daemon)
    elif args.control:run_control(config, args.daemon)
    elif args.standalone:run_standalone(config,args.daemon)
    elif args.secret:run_secret_update(config,args.conf)
    elif args.initdb:run_initdb(config)
    else: print 'do nothing'
    
        

    
    
    


