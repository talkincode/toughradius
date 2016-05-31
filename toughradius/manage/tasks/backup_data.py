#!/usr/bin/env python
#coding:utf-8
import os
import sys
import time
from toughlib import utils
from toughlib import dispatch,logger
from toughradius.manage import models
from toughlib.db_backup import DBBackup
from toughradius.manage.tasks.task_base import TaseBasic
from twisted.internet import reactor
from toughradius.manage import taskd

class BackupDataTask(TaseBasic):

    __name__ = 'db-backup'    

    def __init__(self,taskd, **kwargs):
        TaseBasic.__init__(self,taskd, **kwargs)
        self.db_backup = DBBackup(models.get_metadata(taskd.db_engine), excludes=[
            'tr_online','system_session','system_cache','tr_ticket','tr_billing','tr_online_stat','tr_flow_stat'
        ])

    def get_notify_interval(self):
        return utils.get_cron_interval('02:00')

    def first_delay(self):
        return self.get_notify_interval()

    def process(self, *args, **kwargs):
        self.logtimes()
        next_interval = self.get_notify_interval()
        backup_path = self.config.database.backup_path
        backup_file = "trdb_cron_backup_%s.json.gz" % utils.gen_backep_id()
        try:
            self.db_backup.dumpdb(os.path.join(backup_path, backup_file))
            logger.info(u"数据备份完成,下次执行还需等待 %s"%(self.format_time(next_interval)),trace="task")
        except Exception as err:
            logger.info(u"数据备份失败,%s, 下次执行还需等待 %s"%( repr(err), self.format_time(next_interval)),trace="task")
            logger.exception(err)

        try:
            bak_list = [ bd for bd in os.listdir(backup_path) if 'trdb_cron_backup' in bd]
            if len(bak_list) > 7:
                logger.info("find expire backup file and remove")
                _count = 0
                for fname in bak_list:
                    fpath = os.path.join(backup_path, fname)
                    if (time.time() - os.path.getctime(fpath))/(3600*24)  > 14:
                        os.remove(fpath)
                        _count += 1
                        logger.debug("remove expire backup file %s"%fpath)
                logger.info("remove expire backup file total %s"%_count,trace="task")
        except Exception as err:
            logger.exception(err)
            
        return next_interval

taskd.TaskDaemon.__taskclss__.append(BackupDataTask)

