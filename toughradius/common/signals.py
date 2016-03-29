#!/usr/bin/env python
# coding=utf-8
from twisted.internet import reactor
from toughlib import logger
from txzmq import ZmqEndpoint, ZmqFactory,ZmqSubConnection

class TermSignalSubscriber:

    def __init__(self, signal_str):
        self.subscriber = ZmqSubConnection(ZmqFactory(), 
            ZmqEndpoint('connect', 'ipc:///tmp/radiusd-exit-sub'))
        self.subscriber.subscribe(signal_str)
        self.subscriber.gotMessage = self.on_quit

    def on_quit(self,*args):
        logger.info("Termination signal received: %r" % (args, ))
        reactor.callFromThread(reactor.stop)
