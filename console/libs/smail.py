#!/usr/bin/env python
#coding:utf-8
import sys,os
import time
import smtplib
from email.mime.text import MIMEText
from email.Header import Header

class Mail(object):
    
    def __init__(self,server=None,user=None,pwd=None,fromaddr=None):
        self.server = server
        self.user = user
        self.pwd = pwd
        self.fromaddr = fromaddr

    def sendmail(self,mailto,topic,content):
        #print 'mailto',mailto,topic,content
        topic = topic.replace("\\n","<br>")
        content = content.replace("\\n","<br>")
        mail = MIMEText(content, 'html', 'utf-8')
        mail['Subject'] = Header("[Alert]:%s"%topic,'utf-8')
        mail['From'] = "notify <%s>"%self.fromaddr
        mail['To'] = "%s,%s"%(toaddr,mailto)
        mail["Accept-Language"]="zh-CN"
        mail["Accept-Charset"]="ISO-8859-1,utf-8"
        try:
            serv = smtplib.SMTP()
            #serv.set_debuglevel(True)
            serv.connect(self.server)
            serv.login(self.user,self.pwd)
            serv.sendmail(self.fromaddr, [mailto], mail.as_string())
            serv.quit()
            print "Successfully sent email"
        except Exception,e:
            print "Error: unable to send email %s"%str(e)
            
mail = Mail()