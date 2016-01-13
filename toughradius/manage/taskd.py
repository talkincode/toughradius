#!/usr/bin/env python
#coding:utf-8
import sys
import os
import time
from twisted.python import log
from twisted.internet import reactor
from sqlalchemy.orm import scoped_session, sessionmaker
from toughlib import logger, utils
from toughlib.dbutils import make_db
from toughradius.manage import models
from toughlib.dbengine import get_engine
from toughradius.manage.tasks import expire_notify
import toughradius


class TaskDaemon():

    def __init__(self, config=None, dbengine=None, log=None, **kwargs):

        self.config = config
        self.syslog = log or logger.Logger(config)
        self.db_engine = dbengine or get_engine(config)
        self.db = scoped_session(sessionmaker(bind=self.db_engine, autocommit=False, autoflush=False))
        self.expire_notify_task = expire_notify.ExpireNotifyTask(config,self.db,log)

    def start_expire_notify(self):
        _time = self.expire_notify_task.process()
        reactor.callLater(_time, self.start_expire_notify)

    def start(self):
        self.start_expire_notify()



def run(config, dbengine=None,log=None):
    app = TaskDaemon(config, dbengine, log)
    app.start()