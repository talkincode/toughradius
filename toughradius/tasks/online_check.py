#!/usr/bin/env python
#coding:utf-8
import sys
import time
import datetime
from toughlib import utils
from toughlib import dispatch,logger
from toughradius import models
from toughlib.dbutils import make_db
from toughradius.tasks.task_base import TaseBasic
from toughradius.events.settings import CLEAR_ONLINE_EVENT
from twisted.internet import reactor

class OnlineCheckTask(TaseBasic):

    __name__ = 'online-checks'

    def first_delay(self):
        return 5

    def get_notify_interval(self):
        return 30        

    def process(self, *args, **kwargs):
        self.logtimes()
        with make_db(self.db) as db:
            try:
                onlines = db.query(models.TrOnline)
                for online in onlines:
                    acct_start_time = datetime.datetime.strptime(online.acct_start_time, '%Y-%m-%d %H:%M:%S')
                    acct_session_time = online.billing_times
                    nowdate = datetime.datetime.now()
                    dt = nowdate - acct_start_time
                    online_times = dt.total_seconds()
                    max_interim_intelval = int(self.get_param_value('radius_acct_interim_intelval',240))
                    if (online_times - acct_session_time) > (max_interim_intelval+30):
                        logger.info("online %s overtime, system auto clear this online"%online.account_number)
                        dispatch.pub(CLEAR_ONLINE_EVENT,
                            online.account_number,
                            online.nas_addr, 
                            online.acct_session_id,async=True)
            except Exception as err:
                db.rollback()
                logger.exception(err)
        
        return self.get_notify_interval()

taskcls = OnlineCheckTask