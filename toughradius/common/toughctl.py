#!/usr/bin/env python
# -*- coding: utf-8 -*-
from toughlib import choosereactor
choosereactor.install_optimal_reactor(True)
from twisted.internet import reactor
from twisted.python import log
import argparse
from toughlib import config as iconfig
from toughlib import logger
from toughlib.dbengine import get_engine
from toughradius.common import initdb as init_db
from toughradius.manage import webserver
from toughradius.manage import radiusd
import sys
import os

def update_timezone(config):
    try:
        if 'TZ' not in os.environ:
            os.environ["TZ"] = config.system.tz
        time.tzset()
    except:
        pass

def check_env(config):
    try:
        backup_path = config.database.backup_path
        if not os.path.exists(backup_path):
            os.system("mkdir -p  %s" % backup_path)
        if not os.path.exists("/var/toughradius"):
            os.system("mkdir -p /var/toughradius")
    except Exception as err:
        import traceback
        traceback.print_exc()

def run_initdb(config):
    init_db.update(config)


def run():
    log.startLogging(sys.stdout)
    parser = argparse.ArgumentParser()
    parser.add_argument('-manage', '--manage', action='store_true', default=False, dest='manage', help='run manage')
    parser.add_argument('-auth', '--auth', action='store_true', default=False, dest='auth', help='run auth')
    parser.add_argument('-acct', '--acct', action='store_true', default=False, dest='acct', help='run acct')
    parser.add_argument('-standalone', '--standalone', action='store_true', default=False, dest='standalone', help='run standalone')
    parser.add_argument('-initdb', '--initdb', action='store_true', default=False, dest='initdb', help='run initdb')
    parser.add_argument('-debug', '--debug', action='store_true', default=False, dest='debug', help='debug option')
    parser.add_argument('-c', '--conf', type=str, default="/etc/toughradius.json", dest='conf', help='config file')
    args = parser.parse_args(sys.argv[1:])

    config = iconfig.find_config(args.conf)
    syslog = logger.Logger(config)
    dbengine = get_engine(config)

    update_timezone(config)
    check_env(config)

    if args.debug:
        config.defaults.debug = True

    if args.manage:
        webserver.run(config,dbengine=dbengine,log=syslog)
        reactor.run()    

    elif args.auth:
        radiusd.run_auth(config,dbengine=dbengine,log=syslog)
        reactor.run()
    
    elif args.acct:
        radiusd.run_acct(config,dbengine=dbengine,log=syslog)
        reactor.run()

    elif args.standalone:
        radiusd.run_auth(config,dbengine=dbengine,log=syslog)
        radiusd.run_acct(config,dbengine=dbengine,log=syslog)
        webserver.run(config,dbengine=dbengine,log=syslog)
        reactor.run()
        
    elif args.initdb:
        run_initdb(config)
    else:
        parser.print_help()
    
        

    
    
    


