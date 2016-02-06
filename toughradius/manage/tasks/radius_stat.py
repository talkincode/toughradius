#!/usr/bin/env python
#coding:utf-8
import sys, struct
import msgpack
from toughlib import utils
from toughlib import logger
from txradius import statistics
from toughradius.manage.tasks.task_base import TaseBasic
from twisted.internet import reactor,defer
from txzmq import ZmqEndpoint, ZmqFactory, ZmqPushConnection, ZmqPullConnection
from toughradius.manage.settings import  radius_statcache_key

class RadiusStatTask(TaseBasic):

    def __init__(self,taskd, **kwargs):
        TaseBasic.__init__(self,taskd, **kwargs)
        self.statdata = statistics.MessageStat()
        self.puller = ZmqPullConnection(ZmqFactory(), ZmqEndpoint('bind', 'ipc:///tmp/radiusd-stat-task'))
        self.puller.onPull = self.update_stat
        logger.info("init Radius stat puller : %s " % ( self.puller))

    def update_stat(self, message):
        statattrs = msgpack.unpackb(message[0])
        for statattr in statattrs:
            self.statdata.incr(statattr,incr=1)

    def process(self, *args, **kwargs):
        try:
            self.statdata.run_stat()
            if self.cache.get(radius_statcache_key):
                self.cache.update(radius_statcache_key,self.statdata)
            else:
                self.statdata = statistics.MessageStat()
                self.cache.set(radius_statcache_key,self.statdata)
        except Exception as err:
            logger.error('radius stat process error %s' % utils.safeunicode(err.message))

        return 10.0
