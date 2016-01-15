#!/usr/bin/env python
#coding:utf-8

from toughlib import  utils,httpclient
from toughlib import dispatch,logger
from toughradius.manage import models
from toughlib.dbutils import make_db
from toughradius.manage.tasks.task_base import TaseBasic
from twisted.internet import reactor
from twisted.mail.smtp import sendmail
from email.mime.text import MIMEText
from email import Header

class ExpireNotifyTask(TaseBasic):

    def send_mail(self, mailto, topic, content):
        smtp_server = self.get_param_value("smtp_server",'127.0.0.1')
        from_addr = self.get_param_value("smtp_from")
        smtp_port = int(self.get_param_value("smtp_port",25))
        smtp_user = self.get_param_value("smtp_user",None)
        smtp_pwd = self.get_param_value("smtp_pwd",None)
        to = ["bob@gmail.com", "charlie@gmail.com"]

        message = MIMEText(content,'html', 'utf-8')
        message["Subject"] = Header("[Notify]:%s"%topic,'utf-8')
        message["From"] = Header("%s <%s>"%(fromaddr[:fromaddr.find('@')],fromaddr),'utf-8')
        message["To"] = mailto
        message["Accept-Language"]="zh-CN"
        message["Accept-Charset"]="ISO-8859-1,utf-8"
        if '@toughradius.org' in fromaddr:
            message['X-Mailgun-SFlag'] = 'yes'
            message['X-Mailgun-SScore'] = 'yes'

        return sendmail(smtp_server, from_addr, mailto, message,
                        port=smtp_port, username=smtp_user, password=smtp_pwd)

    def process(self, *args, **kwargs):
        dispatch.pub(logger.EVENT_INFO,"process expire notify task..")
        with make_db(self.db) as db:
            _enable = int(self.get_param_value("expire_notify_enable",0))
            if not _enable:
                return 120.0
            _ndays = self.get_param_value("expire_notify_days")
            notify_tpl = self.get_param_value("expire_notify_tpl")
            notify_url = self.get_param_value("expire_notify_url")
            notify_interval = int(self.get_param_value("expire_notify_interval",1440)) * 60.0

            _now = datetime.datetime.now()
            _date = (datetime.datetime.now() + datetime.timedelta(days=int(_ndays))).strftime("%Y-%m-%d")
            expire_query = db.query(
                models.TrAccount.account_number,
                models.TrAccount.expire_date,
                models.TrCustomer.email,
                models.TrCustomer.mobile
            ).filter(
                models.TrAccount.customer_id == models.TrCustomer.customer_id,
                models.TrAccount.expire_date <= _date,
                models.TrAccount.expire_date >= _now.strftime("%Y-%m-%d"),
                models.TrAccount.status == 1
            )

            dispatch.pub(logger.EVENT_INFO,'expire_notify total: %s'%expire_query.count())
            for account,expire,email,mobile in expire_query:
                ctx = notify_tpl.replace('#account#',account)
                ctx = ctx.replace('#expire#',expire)
                topic = ctx[:ctx.find('\n')]
                if email:
                    self.send_mail(email, topic, ctx).addCallbacks(self.syslog.info,self.syslog.error)
                
                url = notify_url.replace('{account}',account)
                url = url.replace('{expire}',expire)
                url = url.replace('{email}',email)
                url = url.replace('{mobile}',mobile)
                url = url.encode('utf-8')
                url = quote(url,":?=/&")
                httpclient.get(url).addCallbacks(self.syslog.info,self.syslog.error)

        return notify_interval



