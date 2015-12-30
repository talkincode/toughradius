#!/usr/bin/env python
# coding:utf-8 
from twisted.internet import reactor
from txzmq import ZmqEndpoint, ZmqFactory, ZmqREQConnection, ZmqREPConnection, ZmqRequestTimeoutError
from toughlib import apiutils, utils
from toughradius.manage import models
from toughlib.utils import timecast

class ZAcctAgent:

    def __init__(self, app):
        self.app = app
        self.config = app.config
        self.cache = app.mcache
        self.db = app.db
        self.syslog = app.syslog
        self.secret = app.config.system.secret

        zfactory = ZmqFactory()
        self.listen = "tcp://*:{0}".format(int(self.config.admin.zacct_port))
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

    @timecast
    def process(self, msgid, message):
        self.syslog.info("accept acct message @ %s : %r" % (self.listen, utils.safeunicode(message)))
        try:
            req_msg = apiutils.parse_request(self.secret, message)
            if req_msg.get("action") == 'ping':
                return self.agent.reply(msgid, apiutils.make_message(self.secret,code=0))
        except Exception as err:
            resp = apiutils.make_message(self.secret, code=1, msg=utils.safestr(err.message))
            self.agent.reply(msgid, resp)
            return



