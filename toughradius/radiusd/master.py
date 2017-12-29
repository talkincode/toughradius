#!/usr/bin/env python
#coding:utf-8
from gevent import monkey; monkey.patch_all(thread=False)
from gevent.server import DatagramServer
import socket
import gevent
import logging
import os

logger = logging.getLogger(__name__)

def setsockopt(_socket):
    try:
        _socket.setsockopt(socket.SOL_SOCKET, socket.SO_SNDBUF, 32 * 1024 * 1024)
        _socket.setsockopt(socket.SOL_SOCKET, socket.SO_RCVBUF, 32 * 1024 * 1024)
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
        logger.info(self)
        jobs = [gevent.spawn(self.handle_result) for x in range(pool_size)]
        # jobs.append(self.print_que())
        gevent.joinall(jobs)

    def print_que(self):
        while 1:
            logger.info("rquest queue: %s; response queue: %s;" % (self.req_q.qsize(), self.rep_q.qsize()))
            gevent.sleep(1)


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
        jobs = [gevent.spawn(self.handle) for x in range(pool_size)]
        gevent.joinall(jobs)

    def handle(self):
        while 1:
            if not self.req_q.empty():
                data, address = self.req_q.get()
                gevent.spawn(self.adapter_handle, data, address, self.rep_q)
                gevent.sleep(0)
            else:
                gevent.sleep(0.01)

























