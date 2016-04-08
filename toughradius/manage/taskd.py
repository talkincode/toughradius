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
from toughlib.redis_cache import CacheManager
from toughradius.manage.settings import redis_conf
from toughradius.manage.tasks import (
    expire_notify, ddns_update, radius_stat, online_stat,flow_stat
)
from toughradius.manage.events import radius_events
from toughradius.manage import settings
import toughradius


class TaskDaemon():

    def __init__(self, config=None, dbengine=None, **kwargs):
        self.config = config
        self.db_engine = dbengine or get_engine(config,pool_size=20)
        self.aes = kwargs.pop("aes",None)
        self.cache = kwargs.pop("cache",CacheManager(redis_conf(config),cache_name='RadiusTaskCache-%s'%os.getpid()))
        self.cache.print_hit_stat(60)
        self.db = scoped_session(sessionmaker(bind=self.db_engine, autocommit=False, autoflush=False))
        # init task
        self.expire_notify_task = expire_notify.ExpireNotifyTask(self)
        self.ddns_update_task = ddns_update.DdnsUpdateTask(self)
        self.radius_stat_task = radius_stat.RadiusStatTask(self)
        self.online_stat_task = online_stat.OnlineStatTask(self)
        self.flow_stat_task = flow_stat.FlowStatTask(self)
        if not kwargs.get('standalone'):
            event_params= dict(dbengine=self.db_engine, mcache=self.cache,aes=self.aes)
            event_path = os.path.abspath(os.path.dirname(toughradius.manage.events.__file__))
            dispatch.load_events(event_path,"toughradius.manage.events",event_params=event_params)

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



def run(config, dbengine=None,**kwargs):
    app = TaskDaemon(config, dbengine)
    app.start()
