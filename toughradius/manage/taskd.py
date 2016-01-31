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
from toughradius.manage.tasks import expire_notify, ddns_update, radius_stat
from toughradius.manage.events import radius_events
import toughradius


class TaskDaemon():

    def __init__(self, config=None, dbengine=None, **kwargs):

        self.config = config
        self.db_engine = dbengine or get_engine(config)
        self.cache = cache.CacheManager(self.db_engine)
        self.db = scoped_session(sessionmaker(bind=self.db_engine, autocommit=False, autoflush=False))
        # init task
        self.expire_notify_task = expire_notify.ExpireNotifyTask(self)
        self.ddns_update_task = ddns_update.DdnsUpdateTask(self)
        self.radius_stat_task = radius_stat.RadiusStatTask(self)

        dispatch.register(radius_events.__call__(self.db_engine,self.cache))

    def start_expire_notify(self):
        _time = self.expire_notify_task.process()
        reactor.callLater(_time, self.start_expire_notify)

    def start_ddns_update(self):
        d = self.ddns_update_task.process()
        d.addCallback(reactor.callLater,self.start_ddns_update)

    def start_radius_stat_update(self):
        _time = self.radius_stat_task.process()
        reactor.callLater(_time, self.start_radius_stat_update)


    def start(self):
        self.start_expire_notify()
        self.start_ddns_update()
        self.start_radius_stat_update()



def run(config, dbengine=None):
    app = TaskDaemon(config, dbengine)
    app.start()