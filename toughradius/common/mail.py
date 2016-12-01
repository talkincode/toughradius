#!/usr/bin/env python
#coding:utf-8

from cStringIO import StringIO
from OpenSSL.SSL import OP_NO_SSLv3
from twisted.mail.smtp import ESMTPSenderFactory
from twisted.internet.ssl import ClientContextFactory
from twisted.internet.defer import Deferred
from twisted.internet import reactor
from toughradius.common import  utils
from toughradius.common import dispatch,logger
from twisted.mail.smtp import sendmail
# from email.mime.text import MIMEText
# from email import Header
import email

class ContextFactory(ClientContextFactory):
    def getContext(self):
        """Get the parent context but disable SSLv3."""
        ctx = ClientContextFactory.getContext(self)
        ctx.set_options(OP_NO_SSLv3)
        return ctx

class SendMail:

    def __init__(self, server='127.0.0.1', port=25, user=None, password=None,from_addr=''):
        self.smtp_server = server
        self.from_addr = from_addr
        self.smtp_port = int(port)
        self.smtp_user = user
        self.smtp_pwd = password

    def send_mail(self, mailto, topic, content, tls=False,**kwargs):
        message = email.MIMEText.MIMEText(content,'html', 'utf-8')
        message["Subject"] = email.Header.Header(topic,'utf-8')
        message["From"] = self.from_addr
        message["To"] = mailto
        message["Accept-Language"]="zh-CN"
        message["Accept-Charset"]="ISO-8859-1,utf-8"
        if not tls:
            logger.info('send mail:%s:%s:%s'%(self.smtp_server,self.smtp_port,mailto))
            return sendmail(self.smtp_server, self.from_addr, mailto, message,
                        port=self.smtp_port, username=self.smtp_user, password=self.smtp_pwd)
        else:
            logger.info('send tls mail:%s:%s:%s'%(self.smtp_server,self.smtp_port,mailto))
            contextFactory = ContextFactory()
            resultDeferred = Deferred()
            senderFactory = ESMTPSenderFactory(
                self.smtp_user,
                self.smtp_pwd,
                self.from_addr,
                mailto,
                StringIO(message.as_string()),
                resultDeferred,
                contextFactory=contextFactory,
                requireAuthentication=(self.smtp_user and self.smtp_pwd),
                requireTransportSecurity=tls)

            reactor.connectTCP(self.smtp_server, self.smtp_port, senderFactory)
            return resultDeferred


def send_mail(server='127.0.0.1', port=25, user=None, password=None, 
                from_addr=None, mailto=None, topic=None, content=None, tls=False, **kwargs):
    sender = SendMail(server,port,user,password,from_addr)
    return sender.send_mail(mailto, topic, content, tls=tls,**kwargs)










