#!/usr/bin/env python
# coding: utf-8

import sys
import six
import argparse
import gevent
import logging
import logging.config
from toughradius import settings

logging.config.dictConfig(settings.LOGGER)

logger = logging.getLogger(__name__)

def test_auth():
    import radclient
    parser = argparse.ArgumentParser()
    parser.add_argument('-H','--server', default='127.0.0.1', dest="server",type=str)
    parser.add_argument('-P','--port', default='1812', dest="port",type=int)
    parser.add_argument('-s','--secret', default='testing123', dest="secret",type=str)
    parser.add_argument('-u','--user', default='test01', dest="user",type=str)
    parser.add_argument('-p','--passwd', default='888888', dest="passwd",type=str)
    parser.add_argument('-t','--timeout', default='5', dest="timeout",type=int)
    parser.add_argument('-e','--encrypt', default='pap', dest="encrypt",type=str)
    parser.add_argument('-x','--debug', action='store_true',default=False,dest='debug',help='debug option')
    args =  parser.parse_args(sys.argv[1:])
    auth_req = {'User-Name':args.user}
    if args.encrypt == "pap":
        auth_req['User-Password'] = args.passwd
    elif args.encrypt == "chap":
        auth_req['CHAP-Password-Plaintext'] = args.passwd

    auth_req["NAS-IP-Address"] = "192.168.0.1"
    auth_req["NAS-Port"] = 0
    auth_req["Service-Type"] =  "Login-User"
    auth_req["NAS-Identifier"] = "toughac"
    auth_req["Calling-Station-Id"] = "00:00:00:00:00:00"
    auth_req["Framed-IP-Address"] = "10.10.10.10"
    radclient.send_auth(args.server,port=args.port,secret=six.b(args.secret),debug=args.debug,timeout=int(args.timeout),**auth_req)

def test_acct():
    import radclient
    parser = argparse.ArgumentParser()
    parser.add_argument('-H','--server', default='127.0.0.1', dest="server",type=str)
    parser.add_argument('-P','--port', default='1813', dest="port",type=int)
    parser.add_argument('-s','--secret', default='testing123', dest="secret",type=str)
    parser.add_argument('-u','--user', default='test01', dest="user",type=str)
    parser.add_argument('-t','--timeout', default='5', dest="timeout",type=int)
    parser.add_argument('-x','--debug', action='store_true',default=False,dest='debug',help='debug option')
    args =  parser.parse_args(sys.argv[1:])
    acct_req = {
        'User-Name': args.user,
        'Acct-Session-Time': 0,
        'Acct-Status-Type': 1,
        'Session-Timeout': 6,
        'Acct-Session-Id': "123456789",
        "NAS-IP-Address": "192.168.0.1",
        "NAS-Port": 0,
        "NAS-Identifier": "toughac",
        "Calling-Station-Id": "00:00:00:00:00:00",
        "Framed-IP-Address": "10.10.10.10",
        "Acct-Output-Octets": 0,
        "Acct-Input-Octets": 0,
        "NAS-Port-Id": "3/0/1:0.0"
    }
    logger.info('accounting start...')
    radclient.send_acct(args.server,port=args.port,secret=six.b(args.secret),debug=args.debug,timeout=int(args.timeout),**acct_req)
    gevent.sleep(3)
    acct_req['Acct-Status-Type'] = 3
    acct_req['Acct-Session-Time'] = 3
    acct_req['Acct-Input-Octets'] = 1024 * 8
    acct_req['Acct-Output-Octets'] = 1024 * 8
    logger.info('accounting update...')
    radclient.send_acct(args.server, port=args.port, secret=six.b(args.secret), debug=args.debug, timeout=int(args.timeout),**acct_req)
    gevent.sleep(3)
    acct_req['Acct-Status-Type'] = 2
    acct_req['Acct-Session-Time'] = 6
    acct_req['Acct-Input-Octets'] = 1024 * 16
    acct_req['Acct-Output-Octets'] = 1024 * 16
    logger.info('accounting stop...')
    radclient.send_acct(args.server, port=args.port, secret=six.b(args.secret), debug=args.debug, timeout=int(args.timeout),**acct_req)