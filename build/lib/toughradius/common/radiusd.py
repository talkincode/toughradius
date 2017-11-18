#!/usr/bin/env python
# coding: utf-8
import os
import sys
import gevent
import argparse
import logging
import logging.config
import importlib

def run():
    from toughradius import settings
    logging.config.dictConfig(settings.LOGGER)
    logger = logging.getLogger(__name__)
    parser = argparse.ArgumentParser()
    parser.add_argument('--auth-port', default='1812', dest="auth_port",type=int)
    parser.add_argument('--acct-port', default='1813', dest="acct_port",type=int)
    parser.add_argument('--pool', default='1024', dest="pool",type=int)
    parser.add_argument('--rest-auth-url', default=None, dest="rest_auth_url",type=str)
    parser.add_argument('--rest-acct-url', default=None, dest="rest_acct_url",type=str)
    parser.add_argument('-x','--debug', action='store_true',default=False,dest='debug',help='debug option')
    args =  parser.parse_args(sys.argv[1:])

    if args.debug:
        settings.RADIUSD.update(debug=1)
        os.environ['TOUGHRADIUS_DEBUG_ENABLED'] = "1"

    if args.auth_port > 0:
        settings.RADIUSD.update(auth_port=args.auth_port)

    if args.acct_port > 0:
        settings.RADIUSD.update(acct_port=args.acct_port)

    if args.rest_auth_url:
        settings.ADAPTERS['rest'].update(authurl=args.rest_auth_url)

    if args.rest_acct_url:
        settings.ADAPTERS['rest'].update(accturl=args.rest_acct_url)

    from toughradius.radiusd.master import RudiusAuthServer
    from toughradius.radiusd.master import RudiusAcctServer
    host = settings.RADIUSD['host']
    auth_port = settings.RADIUSD['auth_port']
    acct_port = settings.RADIUSD['acct_port']
    pool_size = settings.RADIUSD['pool_size']
    adapter_class = settings.RADIUSD['adapter']
    adapter = importlib.import_module(adapter_class).adapter(settings)
    auth_server = RudiusAuthServer(adapter, host=host, port=auth_port, pool_size=pool_size)
    acct_server = RudiusAcctServer(adapter, host=host, port=acct_port, pool_size=pool_size)
    auth_server.start()
    acct_server.start()
    logger.info(auth_server)
    logger.info(acct_server)
    gevent.wait()

if __name__ == "__main__":
    run()