#!/usr/bin/env python
# coding:utf-8 
from twisted.internet import reactor
from txzmq import ZmqEndpoint, ZmqFactory, ZmqREQConnection, ZmqREPConnection, ZmqRequestTimeoutError
from toughradius.common import apibase, utils
from toughradius.console import models

class ZAcctAgent:

    def __init__(self, app):
        self.app = app
        self.config = app.config
        self.cache = app.cache
        self.db = app.db
        self.syslog = app.syslog
        self.secret = app.config.defaults.secret

        zfactory = ZmqFactory()
        self.listen = "tcp://{0}:{1}".format(self.config.admin.agent_addr, int(self.config.admin.port)+97)
        endpoint = ZmqEndpoint('bind', self.listen)
        self.agent = ZmqREPConnection(zfactory, endpoint)
        self.agent.gotMessage = self.process
        self.register()
        self.syslog.info('zmq acctounting agent running %s' % self.listen)

    def register(self):
        conn = self.db()
        try:
            node = conn.query(models.TrRadAgent).filter_by(
                endpoint=self.listen,
                protocol='zeromq',
                radius_type='acctounting'
            ).first()

            if not node:
                node = models.TrRadAgent()
                node.radius_type = 'acctounting'
                node.protocol = 'zeromq'
                node.endpoint = self.listen
                node.create_time = utils.get_currtime()
                node.last_check = utils.get_currtime()
                conn.add(node)
                conn.commit()
            else:
                node.last_check = utils.get_currtime()
                conn.commit()
        except Exception as err:
            self.syslog.error(u"register acctounting agent error %s" % utils.safeunicode(err.message))
        finally:
            conn.close()
            
        reactor.callLater(10.0, self.register, )


    def process(self, msgid, message):
        print "Replying to %s, %r" % (msgid, message)
        self.agent.reply(msgid, "%s %r " % (msgid, message))