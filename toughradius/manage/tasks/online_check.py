#!/usr/bin/env python
#coding:utf-8
import sys
import time
import datetime
from toughlib import utils
from toughlib import dispatch,logger
from toughradius.manage import models
from toughlib.dbutils import make_db
from toughradius.manage.tasks.task_base import TaseBasic
from toughradius.manage.events.settings import UNLOCK_ONLINE_EVENT
from twisted.internet import reactor
from toughradius.manage import taskd

class OnlineCheckTask(TaseBasic):

    __name__ = 'online-checks'

    def first_delay(self):
        return 5

    def get_notify_interval(self):
        return 3600        

    def process(self, *args, **kwargs):
        self.logtimes()
        with make_db(self.db) as db:
            try:
                onlines = db.query(models.TrOnline)
                for online in onlines:
                    acct_start_time = datetime.datetime.strptime(online.acct_start_time, '%Y-%m-%d %H:%M:%S')
                    nowdate = datetime.datetime.now()
                    dt = nowdate - acct_start_time
                    online_times = dt.total_seconds()
                    max_session_time = int(self.get_param_value('radius_max_session_timeout',86400))
                    if online_times > (max_session_time):
                        logger.info("online %s overtime, system auto disconnect this online"%online.account_number)
                        dispatch.pub(UNLOCK_ONLINE_EVENT,
                            online.account_number,
                            online.nas_addr, 
                            online.acct_session_id,async=True)
                logger.info("online overtime check task done")
            except Exception as err:
                db.rollback()
                logger.error('online overtime check job err,%s'%(str(err)))
        
        return self.get_notify_interval()

taskd.TaskDaemon.__taskclss__.append(OnlineCheckTask)