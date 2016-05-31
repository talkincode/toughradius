#!/usr/bin/env python
#coding:utf-8
import sys
import time
from toughlib import utils
from toughlib import dispatch,logger
from toughradius.manage import models
from sqlalchemy.sql import func
from toughlib.dbutils import make_db
from toughradius.manage.tasks.task_base import TaseBasic
from twisted.internet import reactor
from toughradius.manage import taskd

class FlowStatTask(TaseBasic):

    __name__ = 'flow-stat'

    def first_delay(self):
        return 5

    def get_notify_interval(self):
        return 120        

    def process(self, *args, **kwargs):
        self.logtimes()
        with make_db(self.db) as db:
            try:
                nodes = db.query(models.TrNode)
                for node in nodes:
                    r = db.query(
                        func.sum(models.TrOnline.input_total).label("input_total"),
                        func.sum(models.TrOnline.output_total).label("output_total")
                    ).filter(
                        models.TrOnline.account_number == models.TrAccount.account_number,
                        models.TrAccount.customer_id == models.TrCustomer.customer_id,
                        models.TrCustomer.node_id == node.id
                    ).first()
                    if r and all([r.input_total,r.output_total]):
                        stat = models.TrFlowStat()
                        stat.node_id = node.id
                        stat.stat_time = int(time.time())
                        stat.input_total = r.input_total
                        stat.output_total = r.output_total
                        db.add(stat)

                # clean expire data
                _time = int(time.time()) - (86400 * 2)
                db.query(models.TrFlowStat).filter(models.TrFlowStat.stat_time < _time).delete()

                db.commit()
                logger.info("flow stat task done")
            except Exception as err:
                db.rollback()
                logger.exception(err)
        
        return self.get_notify_interval()

taskd.TaskDaemon.__taskclss__.append(FlowStatTask)