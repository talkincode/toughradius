#!/usr/bin/env python
#coding:utf-8
import sys
import time
from toughlib import utils
from toughlib import dispatch,logger
from toughradius.manage import models
from toughlib.dbutils import make_db
from toughradius.manage.tasks.task_base import TaseBasic
from twisted.internet import reactor
from toughradius.manage import taskd

class OnlineStatTask(TaseBasic):

    __name__ = 'online-stat'

    def first_delay(self):
        return 0

    def process(self, *args, **kwargs):
        self.logtimes()
        with make_db(self.db) as db:
            try:
                nodes = db.query(models.TrNode)
                for node in nodes:
                    online_count = db.query(models.TrOnline.id).filter(
                        models.TrOnline.account_number == models.TrAccount.account_number,
                        models.TrAccount.customer_id == models.TrCustomer.customer_id,
                        models.TrCustomer.node_id == node.id
                    ).count()
                    stat = models.TrOnlineStat()
                    stat.node_id = node.id
                    stat.stat_time = int(time.time())
                    stat.total = online_count
                    db.add(stat)
                db.commit()
                logger.info("online stat task done")
            except Exception as err:
                db.rollback()
                logger.error('online_stat_job err,%s'%(str(err)))
        
        return 120.0

taskd.TaskDaemon.__taskclss__.append(OnlineStatTask)
