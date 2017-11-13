#!/usr/bin/env python
# coding: utf-8
import sys
import gevent
import argparse
import logging
import logging.config
from toughradius import settings
logging.config.dictConfig(settings.logger)

def run():
    parser = argparse.ArgumentParser()
    parser.add_argument('--auth-port', default='1812', dest="auth_port",type=int)
    parser.add_argument('--acct-port', default='1813', dest="acct_port",type=int)
    parser.add_argument('--pool', default='1024', dest="pool",type=int)
    parser.add_argument('--rest-auth-url', default=None, dest="rest_auth_url",type=str)
    parser.add_argument('--rest-acct-url', default=None, dest="rest_acct_url",type=str)
    parser.add_argument('-x','--debug', action='store_true',default=False,dest='debug',help='debug option')
    args =  parser.parse_args(sys.argv[1:])

    from toughradius.radiusd.master import RudiusAuthServer
    from toughradius.radiusd.master import RudiusAcctServer

    if args.debug:
        settings.radiusd['debug'] = 1

    if args.auth_port > 0:
        settings.radiusd['auth_port'] = args.auth_port

    if args.acct_port > 0:
        settings.radiusd['acct_port'] = args.acct_port

    if args.rest_auth_url:
        settings['adapters']['rest']['authurl'] = args.rest_auth_url

    if args.rest_acct_url:
        settings['adapters']['rest']['accturl'] = args.rest_acct_url

    auth_server = RudiusAuthServer()
    auth_server.start()
    logging.info(auth_server)
    acct_server = RudiusAcctServer()
    acct_server.start()
    logging.info(acct_server)
    gevent.wait()


if __name__ == "__main__":
    run()