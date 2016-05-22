#!/usr/bin/env python
#coding:utf-8
import sys
import os
import time
from twisted.python import log
from twisted.internet import reactor,defer
from sqlalchemy.orm import scoped_session, sessionmaker
from toughlib import logger, utils,dispatch
from toughradius.manage import models
from toughlib.dbengine import get_engine
from toughlib.redis_cache import CacheManager
from toughradius.manage.settings import redis_conf
from toughradius.manage import tasks
from toughradius.manage.events import radius_events
from toughradius.manage import settings
from toughlib import logger
import toughradius


class TaskDaemon():

    def __init__(self, config=None, dbengine=None, **kwargs):
        self.config = config
        self.db_engine = dbengine or get_engine(config,pool_size=20)
        self.aes = kwargs.pop("aes",None)
        self.cache = kwargs.pop("cache",CacheManager(redis_conf(config),cache_name='RadiusTaskCache-%s'%os.getpid()))
        self.cache.print_hit_stat(60)
        self.db = scoped_session(sessionmaker(bind=self.db_engine, autocommit=False, autoflush=False))
        if not kwargs.get('standalone'):
            event_params= dict(dbengine=self.db_engine, mcache=self.cache,aes=self.aes)
            event_path = os.path.abspath(os.path.dirname(toughradius.manage.events.__file__))
            dispatch.load_events(event_path,"toughradius.manage.events",event_params=event_params)

    def process_task(self,task):
        r = task.process()
        if isinstance(r, defer.Deferred):
            cbk = lambda _time,_cbk,_task: reactor.callLater(_time, _cbk,_task)
            r.addCallback(cbk,self.process_task,task).addErrback(logger.exception)
        else:
            _time = task.process()
            reactor.callLater(_time, self.process_task,task)

    def start(self):
        taskclss = []
        for name in dir(tasks):
            _module = getattr(tasks,name)
            if hasattr(_module, 'initcls'):
                taskclss.append(_module.initcls)

        for taskcls in taskclss:
            task = taskcls(self)
            first_delay = task.first_delay()
            if first_delay:
                reactor.callLater(first_delay,self.process_task,task)
            else:
                self.process_task(task)
            logger.info('init task %s done'%task.__name__)

        



def run(config, dbengine=None,**kwargs):
    app = TaskDaemon(config, dbengine)
    app.start()
