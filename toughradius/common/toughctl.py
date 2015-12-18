#!/usr/bin/env python
# -*- coding: utf-8 -*-
from toughradius.common import choosereactor
choosereactor.install_optimal_reactor(True)
import argparse
from toughradius.common import config as iconfig
from toughradius.common.dbengine import get_engine
from toughradius.common import initdb as init_db
import sys


def run_admin(config):
    from toughradius.console import admin_app
    admin_app.run(config)

def run_customer(config):
    from toughradius.console import customer_app
    customer_app.run(config)

def run_initdb(config):
    init_db.update(get_engine(config))


def run_dumpdb(config, dumpfs):
    from toughradius.tools import backup
    backup.dumpdb(config, dumpfs)


def run_restoredb(config, restorefs):
    from toughradius.tools import backup
    backup.restoredb(config, restorefs)


def run():
    parser = argparse.ArgumentParser()
    parser.add_argument('-admin', '--admin', action='store_true', default=False, dest='admin', help='run admin')
    parser.add_argument('-customer', '--customer', action='store_true', default=False, dest='customer', help='run customer')
    parser.add_argument('-port', '--port', type=int, default=0, dest='port', help='server port')
    parser.add_argument('-initdb', '--initdb', action='store_true', default=False, dest='initdb', help='run initdb')
    parser.add_argument('-dumpdb', '--dumpdb', type=str, default=None, dest='dumpdb', help='run dumpdb')
    parser.add_argument('-restoredb', '--restoredb', type=str, default=None, dest='restoredb', help='run restoredb')
    parser.add_argument('-debug', '--debug', action='store_true', default=False, dest='debug', help='debug option')
    parser.add_argument('-c', '--conf', type=str, default="/etc/toughradius.conf", dest='conf', help='config file')
    args = parser.parse_args(sys.argv[1:])

    config = iconfig.find_config(args.conf)

    if args.debug:
        config.defaults.debug = True

    if args.dumpdb:
        return run_dumpdb(config, args.dumpdb)

    if args.restoredb:
        return run_restoredb(config, args.restoredb)

    if args.admin:
        if args.port > 0:
            config.admin.port = args.port
        run_admin(config)    

    if args.customer:
        if args.port > 0:
            config.customer.port = args.port        
        run_customer(config)
        
    elif args.initdb:
        run_initdb(config)
    else:
        print 'do nothing'
    
        

    
    
    


