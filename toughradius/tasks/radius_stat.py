#!/usr/bin/env python
#coding:utf-8
import sys, struct
import msgpack
import time
import datetime
from toughradius.common import utils
from toughradius.common import logger
from txradius import statistics
from toughradius.tasks.task_base import TaseBasic
from twisted.internet import reactor,defer
from txzmq import ZmqEndpoint, ZmqFactory, ZmqPushConnection, ZmqPullConnection
from toughradius.manage.settings import  RADIUS_STATCACHE_KEY
from toughradius.manage.settings import  FLOW_STATCACHE_KEY


class RadiusStatTask(TaseBasic):

    __name__ = 'radius-stat'

    def first_delay(self):
        return 5

    def __init__(self,taskd, **kwargs):
        TaseBasic.__init__(self,taskd, **kwargs)
        self.flow_stat = {}
        self.statdata =  self.cache.get(RADIUS_STATCACHE_KEY) or statistics.MessageStat()
        self.puller = ZmqPullConnection(ZmqFactory(), ZmqEndpoint('bind', 'ipc:///tmp/radiusd-stat-task'))
        self.puller.onPull = self.update_stat

    def update_stat(self, message):
        try:
            statmsg = msgpack.unpackb(message[0])
            statattrs = statmsg['statattrs']
            for statattr in statattrs:
                self.statdata.incr(statattr,incr=1)
            self.update_flow_stat(statmsg['raddata'])
        except Exception as err:
            logger.exception(err)

    def update_flow_stat(self,raddata):
        if not raddata:
            return
        self.flow_stat = self.cache.get(FLOW_STATCACHE_KEY) 
        if not self.flow_stat:
            self.flow_stat = {
                'last_input_total' : raddata['input_total'],
                'last_output_total' : raddata['output_total'],
                'input_stat' : [(int(time.time()*1000),0)],
                'output_stat' : [(int(time.time()*1000),0)]
            }
            self.cache.set(FLOW_STATCACHE_KEY,self.flow_stat)
        else:
            _insize = raddata['input_total']-self.flow_stat['last_input_total']
            if _insize > 0:
                self.flow_stat['input_stat'].append((int(time.time()*1000),_insize))

            _outsize = raddata['output_total']-self.flow_stat['last_output_total']
            if _outsize > 0:
                self.flow_stat['output_stat'].append((int(time.time()*1000),_outsize))

            self.flow_stat['last_input_total'] = raddata['input_total']
            self.flow_stat['last_output_total'] = raddata['output_total']

            ivs = [iv for iv in self.flow_stat['input_stat'] if time.time() - (iv[0]/1000.0) < 600]
            ovs = [ov for ov in self.flow_stat['output_stat'] if time.time() - (ov[0]/1000.0) < 600]
            self.flow_stat['input_stat'] = ivs
            self.flow_stat['output_stat'] = ovs


    def process(self, *args, **kwargs):
        # self.logtimes()
        try:
            self.statdata.run_stat()
            self.cache.update(RADIUS_STATCACHE_KEY,self.statdata)
            if self.flow_stat:
                self.cache.set(FLOW_STATCACHE_KEY,self.flow_stat)
        except Exception as err:
            logger.error('radius stat process error %s' % utils.safeunicode(err.message))

        return 10.0

taskcls = RadiusStatTask

