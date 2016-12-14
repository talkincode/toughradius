#!/usr/bin/env python
#coding:utf-8
import datetime
from toughradius.common import  utils
from toughradius.common import dispatch,logger
from toughradius import models
from toughradius.common.dbutils import make_db
from toughradius.tasks.task_base import TaseBasic
from twisted.internet import reactor


class ExpireNotifyTask(TaseBasic):

    __name__ = 'expire-notify'

    def get_next_interval(self):
        try:
            notify_interval = int(self.get_param_value("mail_notify_interval",1440)) * 60.0
            notify_time = self.get_param_value("mail_notify_time", None)
            if notify_time:
                notify_interval = utils.get_cron_interval(notify_time)
            return notify_interval
        except:
            return 120

    def first_delay(self):
        return self.get_next_interval()

    def trigger_notify(self,userinfo):
        if int(self.get_param_value("webhook_notify_enable",0)) > 0:
            dispatch.pub('webhook_account_expire',userinfo, async=False)

        if int(self.get_param_value("mail_notify_enable",0)) > 0:
            dispatch.pub('smtp_account_expire',userinfo, async=False)


    def process(self, *args, **kwargs):
        self.logtimes()
        next_interval = self.get_next_interval()
        try:
            logger.info("start process expire notify task")
            _ndays = self.get_param_value("expire_notify_days")
            _now = datetime.datetime.now()
            _date = (datetime.datetime.now() + datetime.timedelta(
                    days=int(_ndays))).strftime("%Y-%m-%d")

            with make_db(self.db) as db:
                expire_query =  db.query(
                    models.TrCustomer.mobile,
                    models.TrCustomer.realname,
                    models.TrCustomer.email,
                    models.TrProduct.product_name,
                    models.TrAccount.account_number,
                    models.TrAccount.install_address,
                    models.TrAccount.expire_date,
                    models.TrAccount.password
                ).filter(
                    models.TrCustomer.customer_id == models.TrAccount.customer_id,
                    models.TrAccount.product_id == models.TrProduct.id,
                    models.TrAccount.expire_date <= _date,
                    models.TrAccount.expire_date >= _now.strftime("%Y-%m-%d"),
                    models.TrAccount.status == 1
                )

                for userinfo in expire_query:
                    self.trigger_notify(userinfo)

                logger.info(u"到期通知任务已执行(%s个已通知)。下次执行还需等待 %s"% (
                    expire_query.count(),self.format_time(next_interval)),trace="task")
                
        except Exception as err:
            logger.info(u"到期通知任务执行失败，%s。下次执行还需等待 %s"%(
                        repr(err),self.format_time(next_interval)),trace="task")
            logger.exception(err)

        return next_interval

taskcls = ExpireNotifyTask