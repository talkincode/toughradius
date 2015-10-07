#!/usr/bin/env python
#coding:utf-8
import sys,os
import time
import smtplib
from email.mime.text import MIMEText
from email import Header

class Mail(object):
    
    def setup(self,server=None,user=None,pwd=None,fromaddr=None,sender=None):
        self.server = server
        self.user = user
        self.pwd = pwd
        self.fromaddr = fromaddr
        self.sender = sender

    def sendmail(self,mailto,topic,content):
        if not mailto or not topic:return
        # print 'mailto',mailto,topic,content
        topic = topic.replace("\n","<br>")
        content = content.replace("\n","<br>")
        mail = MIMEText(content, 'html', 'utf-8')
        mail['Subject'] = Header("[Notify]:%s"%topic,'utf-8')
        mail['From'] = Header("%s <%s>"%(self.fromaddr[:self.fromaddr.find('@')],self.fromaddr),'utf-8')
        mail['To'] = mailto
        mail["Accept-Language"]="zh-CN"
        mail["Accept-Charset"]="ISO-8859-1,utf-8"
        if '@toughradius.org' in self.fromaddr:
            mail['X-Mailgun-SFlag'] = 'yes'
            mail['X-Mailgun-SScore'] = 'yes'
        try:
            serv = smtplib.SMTP()
            # serv.set_debuglevel(True)
            serv.connect(self.server)
            if self.pwd and self.pwd not in ('no','none','anonymous'):
                serv.login(self.user,self.pwd)
            serv.sendmail(self.fromaddr, [mailto], mail.as_string())
            serv.quit()
            print "Successfully sent email to %s"%mailto
        except Exception,e:
            print "Error: unable to send email %s"%str(e)
            
mail = Mail()