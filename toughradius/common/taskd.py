#!/usr/bin/env python
#coding:utf-8
import sys
import os
import time
import importlib
from twisted.internet import reactor,defer
from toughradius.common import logger, utils

class TaskDaemon(object):
    """ 定时任务管理守护进程
    """

    def __init__(self, gdata=None, **kwargs):
        self.gdata = gdata

    def process_task(self,task):
        r = task.process()
        if isinstance(r, defer.Deferred):
            cbk = lambda _time,_cbk,_task: reactor.callLater(_time, _cbk,_task)
            r.addCallback(cbk,self.process_task,task).addErrback(logger.exception)
        elif isinstance(r,(int,float)):
            reactor.callLater(r, self.process_task,task)

    def start_task(self,taskcls):
        try:
            task = taskcls(self)
            first_delay = task.first_delay()
            if first_delay:
                reactor.callLater(first_delay,self.process_task,task)
            else:
                self.process_task(task)
            logger.info('init task %s done'%repr(task))
        except Exception as err:
            logger.exception(err)

    def load_tasks(self,prifix,task_path):
        evs = set(os.path.splitext(it)[0] for it in os.listdir(task_path))
        for ev in evs:
            try:
                robj = importlib.import_module("%s.%s"% (prifix,ev))
                if hasattr(robj, 'taskcls'):
                    self.start_task(robj.taskcls)
            except Exception as err:
                logger.exception(err)
                continue

