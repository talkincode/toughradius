#!/usr/bin/env python
# coding: utf-8
from gevent import monkey; monkey.patch_all(thread=False)
from multiprocessing import Process, Queue, cpu_count
import os
import sys
import gevent
import argparse
import logging
import logging.config
import importlib
import time

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

    from toughradius.radiusd.master import RadiusServer, RudiusWorker

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