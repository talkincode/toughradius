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
    """ startup default radius server
    """
    parser = argparse.ArgumentParser()
    parser.add_argument('--settings', default='toughradius.settings', dest="settings",type=str, help="settings module")
    parser.add_argument('--auth-port', default='1812', dest="auth_port",type=int, help="radiusd auth port")
    parser.add_argument('--acct-port', default='1813', dest="acct_port",type=int, help="radiusd acct port")
    parser.add_argument('--pool', default='1024', dest="pool",type=int, help="work pool size")
    parser.add_argument('--adapter', default=None, dest="adapter",type=str, help="radius handle adapter module")
    parser.add_argument('--auth', action='store_true',default=True, dest='auth',help='run auth listen')
    parser.add_argument('--acct', action='store_true',default=True, dest='acct',help='run acct listen')
    parser.add_argument('-x','--debug', action='store_true',default=False,dest='debug',help='debug option')
    args =  parser.parse_args(sys.argv[1:])

    settings = importlib.import_module(args.settings)
    logging.config.dictConfig(settings.LOGGER)
    logger = logging.getLogger(__name__)

    if args.debug:
        settings.RADIUSD.update(debug=1)
        os.environ['TOUGHRADIUS_DEBUG_ENABLED'] = "1"

    if args.auth_port > 0:
        settings.RADIUSD.update(auth_port=args.auth_port)

    if args.acct_port > 0:
        settings.RADIUSD.update(acct_port=args.acct_port)

    if args.adapter:
        settings.RADIUSD.update(adapter=args.adapter)


    host = settings.RADIUSD['host']
    auth_port = settings.RADIUSD['auth_port']
    acct_port = settings.RADIUSD['acct_port']
    pool_size = settings.RADIUSD['pool_size']
    adapter_class = settings.RADIUSD['adapter']
    adapter = importlib.import_module(adapter_class).adapter(settings)

    if args.auth:
        from toughradius.radiusd.master import RudiusAuthServer
        auth_server = RudiusAuthServer(adapter, host=host, port=auth_port, pool_size=pool_size)
        auth_server.start()
        logger.info(auth_server)

    if args.acct:
        from toughradius.radiusd.master import RudiusAcctServer
        acct_server = RudiusAcctServer(adapter, host=host, port=acct_port, pool_size=pool_size)
        acct_server.start()
        logger.info(acct_server)

    if args.auth or args.acct:
        gevent.wait()

if __name__ == "__main__":
    run()