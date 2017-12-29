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


def run():
    """ startup default radius server
    """
    parser = argparse.ArgumentParser()
    parser.add_argument('--settings', default='toughradius.settings', dest="settings",type=str, help="settings module")
    parser.add_argument('--auth-port', default='1812', dest="auth_port",type=int, help="radiusd auth port")
    parser.add_argument('--acct-port', default='1813', dest="acct_port",type=int, help="radiusd acct port")
    parser.add_argument('--pool', default='10', dest="pool",type=int, help="pool size")
    parser.add_argument('--worker', default=cpu_count(), dest="worker",type=int, help="worker num")
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
    adapter_class = settings.RADIUSD['adapter']
    adapter = importlib.import_module(adapter_class).adapter(settings)

    from toughradius.radiusd.master import RadiusServer, RudiusWorker

    auth_req_queue = Queue()
    auth_rep_queue = Queue()
    acct_req_queue = Queue()
    acct_rep_queue = Queue()

    jobs = []
    if args.auth:
        jobs.append(Process(name="AuthServer", target=RadiusServer, args=(auth_req_queue, auth_rep_queue, host, auth_port ,args.pool)))
        for x in range(args.worker):
            jobs.append(Process(name="AuthWorker", target=RudiusWorker, args=(auth_req_queue, auth_rep_queue, adapter.handleAuth, args.pool)))

    if args.acct:
        jobs.append(Process(name="AcctServer", target=RadiusServer, args=(acct_req_queue, acct_rep_queue, host, acct_port ,args.pool)))
        for x in range(args.worker):
            jobs.append(Process(name="AcctWorker", target=RudiusWorker, args=(acct_req_queue, acct_rep_queue, adapter.handleAcct, args.pool)))

    for job in jobs:
        job.start()
        logger.info('start process %s' % job)

    for job in jobs:
        job.join()

if __name__ == "__main__":
    run()