#!/usr/bin/env python
#coding:utf-8
import sys
import os
import time
from twisted.python import log
from twisted.internet import reactor
from sqlalchemy.orm import scoped_session, sessionmaker
from toughlib import logger, utils,dispatch
from toughradius.manage import models
from toughlib.dbengine import get_engine
from toughlib import db_cache as cache
from toughlib import redis_cache
from toughradius.manage.tasks import (
    expire_notify, ddns_update, radius_stat, online_stat,flow_stat
)
from toughradius.manage.events import radius_events
import toughradius


class TaskDaemon():

    def __init__(self, config=None, dbengine=None, **kwargs):

        self.config = config
        self.db_engine = dbengine or get_engine(config,pool_size=20)
        redisconf = config.get('redis')
        if redisconf:
            self.cache = redis_cache.CacheManager(redisconf,cache_name='RadiusTaskCache-%s'%os.getpid())
            self.cache.print_hit_stat(10)
        else:
            self.cache = cache.CacheManager(self.db_engine,cache_name='RadiusTaskCache-%s'%os.getpid())
        self.db = scoped_session(sessionmaker(bind=self.db_engine, autocommit=False, autoflush=False))
        # init task
        self.expire_notify_task = expire_notify.ExpireNotifyTask(self)
        self.ddns_update_task = ddns_update.DdnsUpdateTask(self)
        self.radius_stat_task = radius_stat.RadiusStatTask(self)
        self.online_stat_task = online_stat.OnlineStatTask(self)
        self.flow_stat_task = flow_stat.FlowStatTask(self)

        dispatch.register(radius_events.__call__(self.db_engine,self.cache))

    def start_expire_notify(self):
        _time = self.expire_notify_task.process()
        logger.info("next expire_notify times: %s" % _time)
        reactor.callLater(_time, self.start_expire_notify)

    def start_ddns_update(self):
        d = self.ddns_update_task.process()
        d.addCallback(reactor.callLater,self.start_ddns_update)

    def start_radius_stat_update(self):
        _time = self.radius_stat_task.process()
        reactor.callLater(_time, self.start_radius_stat_update)

    def start_online_stat_task(self):
        _time = self.online_stat_task.process()
        reactor.callLater(_time, self.start_online_stat_task)

    def start_flow_stat_task(self):
        _time = self.flow_stat_task.process()
        reactor.callLater(_time, self.start_flow_stat_task)


    def start(self):
        self.start_expire_notify()
        self.start_ddns_update()
        self.start_radius_stat_update()
        self.start_online_stat_task()
        self.start_flow_stat_task()
        logger.info('init task done')



def run(config, dbengine=None):
    app = TaskDaemon(config, dbengine)
    app.start()