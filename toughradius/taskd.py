#!/usr/bin/env python
#coding:utf-8
import sys
import os
import time
import importlib
from twisted.python import log
from twisted.internet import reactor,defer
from sqlalchemy.orm import scoped_session, sessionmaker
from toughlib import logger, utils,dispatch
from toughradius import models
from toughlib.dbengine import get_engine
from toughlib.redis_cache import CacheManager
from toughradius.manage.settings import redis_conf
from toughradius.events import radius_events
from toughradius.manage import settings
from toughradius import tasks
from toughradius.common import log_trace
from toughlib import logger
import toughradius
import functools

class TaskDaemon():

    __taskclss__ = []

    def __init__(self, config=None, dbengine=None, **kwargs):
        self.config = config
        self.db_engine = dbengine or get_engine(config,pool_size=20)
        self.aes = kwargs.pop("aes",None)
        self.cache = kwargs.pop("cache",CacheManager(redis_conf(config),cache_name='RadiusTaskCache-%s'%os.getpid()))
        self.cache.print_hit_stat(300)
        self.db = scoped_session(sessionmaker(bind=self.db_engine, autocommit=False, autoflush=False))
        self.taskclss = []
        self.load_tasks()
        if not kwargs.get('standalone'):
            logger.info("start register taskd events")
            dispatch.register(log_trace.LogTrace(redis_conf(config)),check_exists=True)
            event_params= dict(dbengine=self.db_engine, mcache=self.cache,aes=self.aes)
            event_path = os.path.abspath(os.path.dirname(toughradius.events.__file__))
            dispatch.load_events(event_path,"toughradius.events",event_params=event_params)

    def process_task(self,task):
        r = task.process()
        if isinstance(r, defer.Deferred):
            cbk = lambda _time,_cbk,_task: reactor.callLater(_time, _cbk,_task)
            r.addCallback(cbk,self.process_task,task).addErrback(logger.exception)
        elif isinstance(r,(int,float)):
            reactor.callLater(r, self.process_task,task)

    def start(self):
        for taskcls in TaskDaemon.__taskclss__:
            task = taskcls(self)
            first_delay = task.first_delay()
            if first_delay:
                reactor.callLater(first_delay,self.process_task,task)
            else:
                self.process_task(task)
            logger.info('init task %s done'%task.__name__)

        logger.info("init task num : %s"%len(TaskDaemon.__taskclss__))


    def load_tasks(self):
        evs = set(os.path.splitext(it)[0] for it in os.listdir(os.path.dirname(tasks.__file__)))
        for ev in evs:
            try:
                taskmdl = importlib.import_module("toughradius.tasks.%s"% ev)
                if hasattr(taskmdl, 'taskcls'):
                    TaskDaemon.__taskclss__.append(getattr(taskmdl,'taskcls'))
            except Exception as err:
                logger.exception(err)
                continue


def run(config, dbengine=None,**kwargs):
    app = TaskDaemon(config, dbengine,**kwargs)
    app.start()
