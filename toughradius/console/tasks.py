#!/usr/bin/env python
from twisted.python import log
from twisted.internet import reactor
from twisted.internet import threads
from twisted.web.client import getPage
from twisted.internet import task
from sqlalchemy.sql import func
from toughradius.console import models
from toughradius.console.libs.smail import mail
from urllib import quote
import datetime
import time

def __online_stat_job(mk_db):
    def execute(mk_db):
        log.msg("start online stat task..")
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
            log.msg("online stat task done")
        except Exception as err:
            db.rollback()
            log.err(err,'online_stat_job err')
        finally:
            db.close()

    reactor.callInThread(execute,mk_db)

def __flow_stat_job(mk_db):
    def execute(mk_db):
        log.msg("start flow stat task..")
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
            log.msg("flow stat task done")
        except  Exception as err:
            db.rollback()
            log.err(err,'flow_stat_job err')
        finally:
            db.close()

    reactor.callInThread(execute,mk_db)


def __expire_notify(mk_db):
    def _getpage(url):
        errback = lambda x:log.err(x,'getPage error')
        d = getPage(url)
        d.addErrback(errback)
        
    log.msg("start expire notify task..")
    db = mk_db()
    try:
        _ndays = db.query(models.SlcParam.param_value).filter_by(
            param_name = 'expire_notify_days').scalar()

        notify_tpl = db.query(models.SlcParam.param_value).filter_by(
            param_name = 'expire_notify_tpl').scalar()

        notify_url = db.query(models.SlcParam.param_value).filter_by(
            param_name = 'expire_notify_url').scalar()

        _now = datetime.datetime.now()
        _date = (datetime.datetime.now() + datetime.timedelta(days=int(_ndays))).strftime("%Y-%m-%d")
        expire_query = db.query(
            models.SlcRadAccount.account_number,
            models.SlcRadAccount.expire_date,
            models.SlcMember.email,
            models.SlcMember.mobile
        ).filter(
            models.SlcRadAccount.member_id == models.SlcMember.member_id,
            models.SlcRadAccount.expire_date <= _date,
            models.SlcRadAccount.expire_date >= _now.strftime("%Y-%m-%d"),
            models.SlcRadAccount.status == 1
        )

        log.msg('expire_notify total: %s'%expire_query.count())
        commands = []
        for account,expire,email,mobile in expire_query:
            ctx = notify_tpl.replace('#account#',account)
            ctx = ctx.replace('#expire#',expire)
            topic = ctx[:ctx.find('\n')]
            commands.append( (mail.sendmail,[email,topic,ctx],{}) )
            
            url = notify_url.replace('{account}',account)
            url = url.replace('{expire}',expire)
            url = url.replace('{email}',email)
            url = url.replace('{mobile}',mobile)
            url = url.encode('utf-8')
            url = quote(url,":?=/&")
            commands.append( (_getpage,[url],{}) )

        threads.callMultipleInThread(commands)
        log.msg("expire_notify to ansync task")

    except Exception as err:
        db.rollback()
        log.err(err,'expire notify erro')
    finally:
        db.close()
        
def __clear_ticket_job(mk_db):
    def execute(mk_db):
        log.msg("start clear ticket task..")
        db = mk_db()
        try:
            _days = db.query(models.SlcParam.param_value).filter_by(
                param_name='ticket_expire_days').scalar()

            td = datetime.timedelta(days=int(_days or 90))
            _now = datetime.datetime.now() 
            edate = (_now - td).strftime("%Y-%m-%d 23:59:59")
            db.query(models.SlcRadTicket).filter(
                models.SlcRadTicket.acct_stop_time < edate
            ).delete()
            db.commit()
            log.msg("clear ticket task done")
        except  Exception as err:
            db.rollback()
            log.err(err,'clear_ticket_job err')
        finally:
            db.close()

    reactor.callInThread(execute,mk_db)

def start_online_stat_job(mk_db):
    print ('start online_stat_job...')
    _task = task.LoopingCall(__online_stat_job,mk_db)
    _task.start(300)

def start_flow_stat_job(mk_db):
    print ('start flow_stat_job...')
    _task = task.LoopingCall(__flow_stat_job,mk_db)
    _task.start(300)

def start_expire_notify_job(mk_db):
    print ('start flow_stat_job...')
    _task = task.LoopingCall(__expire_notify,mk_db)
    _task.start(3600*12)


def start_clear_ticket_job(mk_db):
    print ('start clear_ticket_job...')
    _task = task.LoopingCall(__clear_ticket_job,mk_db)
    _task.start(3600*12)