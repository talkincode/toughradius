#!/usr/bin/env python
# coding: utf-8
from gevent import monkey; monkey.patch_all(thread=False)
from gevent.server import DatagramServer
from multiprocessing import Process, Queue, cpu_count
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

    def __init__(self, req_q=None, rep_q=None, host="0.0.0.0", port=1812, pool_size=10):
        DatagramServer.__init__(self,(host,port))
        self.req_q = req_q
        self.rep_q = rep_q
        self.init_socket()
        setsockopt(self.socket)
        self.start()
        logging.info('%s started' % self)
        jobs = [gevent.spawn(self.handle_result) for x in range(pool_size)]
        gevent.joinall(jobs)

    def handle_result(self):
        while 1:
            if not self.rep_q.empty():
                data, address = self.rep_q.get()
                gevent.spawn(self.socket.sendto, data, address)
                gevent.sleep(0)
            else:
                gevent.sleep(0.01)


    def handle(self, data, address):
        self.req_q.put((data, address))
        gevent.sleep(0)


class RudiusWorker(object):

    def __init__(self, req_q=None, rep_q=None,adapter_handle=None, pool_size=10, env=None):
        self.req_q = req_q
        self.rep_q = rep_q
        self.adapter_handle = adapter_handle
        if env:
            os.environ.update(**env)
        logging.info('<RudiusWorker pid=%s> started' % os.getpid())
        jobs = [gevent.spawn(self.handle) for x in range(pool_size)]
        gevent.joinall(jobs)

    def handle_radius(self, data, address):
        reply = self.adapter_handle(data, address,self.rep_q)
        self.rep_q.put((reply, address))

    def handle(self):
        while 1:
            if not self.req_q.empty():
                data, address = self.req_q.get()
                gevent.spawn(self.handle_radius, data, address)
                gevent.sleep(0)
            else:
                gevent.sleep(0.01)



def run():
    """ startup default radius server
    """
    parser = argparse.ArgumentParser()
    parser.add_argument('--settings', default=None, dest="settings",type=str, help="settings module")
    parser.add_argument('--listen', default='0.0.0.0', dest="listen",type=str, help="listen address")
    parser.add_argument('--auth-port', default='1812', dest="auth_port",type=int, help="radiusd auth port")
    parser.add_argument('--acct-port', default='1813', dest="acct_port",type=int, help="radiusd acct port")
    parser.add_argument('--pool', default='10', dest="pool",type=int, help="pool size")
    parser.add_argument('--worker', default=cpu_count(), dest="worker",type=int, help="worker num")
    parser.add_argument('--adapter', default=None, dest="adapter",type=str, help="radius handle adapter module")
    parser.add_argument('--dictionary', default=None, dest="dictionary",type=str, help="radius dictionary dir")
    parser.add_argument('--auth', action='store_true',default=True, dest='auth',help='run auth listen')
    parser.add_argument('--acct', action='store_true',default=True, dest='acct',help='run acct listen')
    parser.add_argument('-x', '--trace', action='store_true',default=False,dest='trace', help='radius trace option')
    args = parser.parse_args(sys.argv[1:])

    env_settings = os.environ.get("TOUGHRADIUS_SETTINGS_MODULE", "toughradius.settings")
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

    auth_req_queue = Queue()
    auth_rep_queue = Queue()
    acct_req_queue = Queue()
    acct_rep_queue = Queue()

    jobs = []
    if args.auth:
        auth_server = Process(name="auth-server", target=RadiusServer, args=(auth_req_queue, auth_rep_queue, host, auth_port ,args.pool))
        auth_server.start()
        jobs.append(auth_server)
        for x in range(args.worker):
            worker = Process(name="auth-worker", target=RudiusWorker, args=(auth_req_queue, auth_rep_queue, adapter.handleAuth, args.pool, os.environ))
            worker.start()
            jobs.append(worker)

    time.sleep(0.01)

    if args.acct:
        acct_server = Process(name="acct-server", target=RadiusServer, args=(acct_req_queue, acct_rep_queue, host, acct_port ,args.pool))
        acct_server.start()
        jobs.append(acct_server)
        for x in range(args.worker):
            worker = Process(name="acct-worker", target=RudiusWorker, args=(acct_req_queue, acct_rep_queue, adapter.handleAcct, args.pool, os.environ))
            worker.start()
            jobs.append(worker)

    for job in jobs:
        job.join()

if __name__ == "__main__":
    run()