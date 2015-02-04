#!/usr/bin/env python
from twisted.python import log
from twisted.internet import task
from datetime import datetime
import models
import platform

__last_online_stat_hour = -1
__last_mysql_backup_day = ''

def __online_stat_job(mk_db):
    global __last_online_stat_hour
    now = datetime.now()
    if now.hour == __last_online_stat_hour:
        return 
    if now.minute >= 58:
        log.msg('start exec online_stat_job @ %s...'%now.hour)
        db = mk_db()
        try:
            day_code = now.strftime( "%Y-%m-%d")
            nodes = db.query(models.SlcNode)
            for node in nodes:
                online_count = db.query(models.SlcRadOnline.id).filter(
                    models.SlcRadOnline.account_number == models.SlcRadAccount.account_number,
                    models.SlcRadAccount.member_id == models.SlcMember.member_id,
                    models.SlcMember.node_id == node.id
                ).count()
                stat = db.query(models.SlcRadOnlineStat).filter_by(
                    node_id = node.id,
                    day_code = day_code,
                    time_num = now.hour
                ).first()
                if not stat:
                    stat = models.SlcRadOnlineStat()
                    stat.node_id = node.id
                    stat.day_code = day_code
                    stat.time_num = now.hour
                    stat.total = online_count
                    db.add(stat)
                else:
                    stat.total = online_count
                log.msg('online_stat %s,%s,%s,%s'%(node.id,day_code,now.hour,stat.total))
            db.commit()
            __last_online_stat_hour = now.hour
            log.msg('exec online_stat_job done @ %s'%now.hour)
        except:
            log.err('exec online_stat_job error @ %s'%now.hour)
            db.rollback()
            import traceback
            traceback.print_exc()
        finally:
            db.close()
            
def start_online_stat_job(mk_db):
    log.msg('start online_stat_job...')
    _task = task.LoopingCall(__online_stat_job,mk_db)
    _task.start(30)
            

    
        



        
    
        
        
    