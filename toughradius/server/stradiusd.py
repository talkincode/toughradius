#!/usr/bin/env python
# coding: utf-8
from gevent import monkey; monkey.patch_all(thread=False)
from gevent.server import DatagramServer
from gevent.pool import Pool
import sys
import argparse
import logging.config
import importlib
import time
import socket
import gevent
import logging
import os

def setsockopt(_socket):
    try:
        _socket.setsockopt(socket.SOL_SOCKET, socket.SO_SNDBUF, 16 * 1024 * 1024)
        _socket.setsockopt(socket.SOL_SOCKET, socket.SO_RCVBUF, 16 * 1024 * 1024)
    except:
        pass

class RadiusServer(DatagramServer):
    """Radius auth server"""
    def __init__(self, host="0.0.0.0", port=1812, adapter_handle=None, pool_size=1024):
        DatagramServer.__init__(self,(host,port))
        self.adapter_handle = adapter_handle
        self.pool = Pool(pool_size)
        self.init_socket()
        setsockopt(self.socket)
        self.start()
        logging.info('%s started' % self)

    def handle_radius(self, data, address):
        reply = self.adapter_handle(data, address)
        if reply:
            self.socket.sendto(reply, address)

    def handle(self, data, address):
        self.pool.spawn(self.handle_radius, data, address)





def run():
    """ startup default radius server
    """
    parser = argparse.ArgumentParser()
    parser.add_argument('--rundir', default='.', dest="rundir",type=str, help="run path")
    parser.add_argument('--settings', default=None, dest="settings",type=str, help="settings module")
    parser.add_argument('--listen', default='0.0.0.0', dest="listen",type=str, help="listen address")
    parser.add_argument('--auth-port', default='1812', dest="auth_port",type=int, help="radiusd auth port")
    parser.add_argument('--acct-port', default='1813', dest="acct_port",type=int, help="radiusd acct port")
    parser.add_argument('--pool', default='20', dest="pool",type=int, help="pool size")
    parser.add_argument('--adapter', default=None, dest="adapter",type=str, help="radius handle adapter module")
    parser.add_argument('--dictionary', default=None, dest="dictionary",type=str, help="radius dictionary dir")
    parser.add_argument('--auth', action='store_true',default=True, dest='auth',help='run auth listen')
    parser.add_argument('--acct', action='store_true',default=True, dest='acct',help='run acct listen')
    parser.add_argument('-x', '--trace', action='store_true',default=False,dest='trace', help='radius trace option')
    args = parser.parse_args(sys.argv[1:])

    env_settings = os.environ.get("TOUGHRADIUS_SETTINGS_MODULE", "toughradius.settings")
    if args.rundir and os.path.exists(args.rundir):
        sys.path.insert(0,args.rundir)

    if args.settings:
        settings = importlib.import_module(args.settings)
    else:
        settings = importlib.import_module(env_settings)

    logging.config.dictConfig(settings.LOGGER)
    logger = logging.getLogger(__name__)

    if args.trace:
        os.environ['TOUGHRADIUS_TRACE_ENABLED'] = "1"

    if args.listen != '0.0.0.0':
        settings.RADIUSD.update(host=args.listen)

    if args.auth_port > 0:
        settings.RADIUSD.update(auth_port=args.auth_port)

    if args.acct_port > 0:
        settings.RADIUSD.update(acct_port=args.acct_port)

    if args.adapter:
        settings.RADIUSD.update(adapter=args.adapter)

    if args.dictionary:
        settings.RADIUSD.update(dictionary=args.dictionary)


    host = settings.RADIUSD['host']
    auth_port = settings.RADIUSD['auth_port']
    acct_port = settings.RADIUSD['acct_port']
    adapter_class = settings.RADIUSD['adapter']
    adapter = importlib.import_module(adapter_class).adapter(settings)

    auth_server = RadiusServer(host,port=auth_port,adapter_handle= adapter.handleAuth, pool_size=args.pool)
    acct_server = RadiusServer(host,port=acct_port,adapter_handle= adapter.handleAcct, pool_size=args.pool)
    auth_server.start()
    acct_server.start()
    gevent.wait()



if __name__ == "__main__":
    run()