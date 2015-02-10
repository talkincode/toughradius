#!/usr/bin/env python
from twisted.internet import task
from sqlalchemy.sql import func
import models
import time

def __online_stat_job(mk_db):
    db = mk_db()
    try:
        nodes = db.query(models.SlcNode)
        for node in nodes:
            online_count = db.query(models.SlcRadOnline.id).filter(
                models.SlcRadOnline.account_number == models.SlcRadAccount.account_number,
                models.SlcRadAccount.member_id == models.SlcMember.member_id,
                models.SlcMember.node_id == node.id
            ).count()
            stat = models.SlcRadOnlineStat()
            stat.node_id = node.id
            stat.stat_time = int(time.time())
            stat.total = online_count
            db.add(stat)
        db.commit()
    except:
        db.rollback()
        import traceback
        traceback.print_exc()
    finally:
        db.close()
        
def __flow_stat_job(mk_db):
    db = mk_db()
    try:
        nodes = db.query(models.SlcNode)
        for node in nodes:
            r = db.query(
                func.sum(models.SlcRadOnline.input_total).label("input_total"),
                func.sum(models.SlcRadOnline.output_total).label("output_total")
            ).filter(
                models.SlcRadOnline.account_number == models.SlcRadAccount.account_number,
                models.SlcRadAccount.member_id == models.SlcMember.member_id,
                models.SlcMember.node_id == node.id
            ).first()
            if r:
                stat = models.SlcRadFlowStat()
                stat.node_id = node.id
                stat.stat_time = int(time.time())
                stat.input_total = r.input_total
                stat.output_total = r.output_total
                db.add(stat)
        db.commit()
    except:
        db.rollback()
        import traceback
        traceback.print_exc()
    finally:
        db.close()
            
def start_online_stat_job(mk_db):
    print ('start online_stat_job...')
    _task = task.LoopingCall(__online_stat_job,mk_db)
    _task.start(30)
            
def start_flow_stat_job(mk_db):
    print ('start flow_stat_job...')
    _task = task.LoopingCall(__flow_stat_job,mk_db)
    _task.start(30)
    
        



        
    
        
        
    