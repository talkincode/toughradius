#!/usr/bin/env python
# coding:utf-8 
from twisted.internet import reactor
from txzmq import ZmqEndpoint, ZmqFactory, ZmqREQConnection, ZmqREPConnection, ZmqRequestTimeoutError
from toughradius.common import apibase, utils
from toughradius.console import models
from toughradius.common.utils import timecast

class ZAuthAgent:

    def __init__(self, app):
        self.app = app
        self.config = app.config
        self.cache = app.mcache
        self.db = app.db
        self.syslog = app.syslog
        self.secret = app.config.defaults.secret

        zfactory = ZmqFactory()
        self.listen = "tcp://*:{0}".format(int(self.config.admin.zauth_port))
        endpoint = ZmqEndpoint('bind', self.listen)
        self.agent = ZmqREPConnection(zfactory, endpoint)
        self.agent.gotMessage = self.process
        self.register()
        self.syslog.info('zmq authorize agent running %s' % self.listen)

    def register(self):
        conn = self.db()
        try:
            node = conn.query(models.TrRadAgent).filter_by(
                endpoint=self.listen,
                protocol='zeromq',
                radius_type='authorize'
            ).first()

            if not node:
                node = models.TrRadAgent()
                node.radius_type = 'authorize'
                node.protocol = 'zeromq'
                node.endpoint = self.listen
                node.create_time = utils.get_currtime()
                node.last_check = utils.get_currtime()
                conn.add(node)
                conn.commit()
        except Exception as err:
            self.syslog.error(u"register authorize agent error %s" % utils.safeunicode(err.message))
        finally:
            conn.close()

        reactor.callLater(10.0, self.register, )

    @timecast
    def process(self, msgid, message):
        self.syslog.info("accept auth message @ %s : %r" % (self.listen, utils.safeunicode(message)))
        @self.cache.cache(expire=600)   
        def get_account_by_username(username):
            return self.db.query(models.TrAccount).filter_by(account_number=username).first()

        @self.cache.cache(expire=600)   
        def get_product_by_id(product_id):
            return self.db.query(models.TrProduct).filter_by(id=product_id).first()

        try:
            req_msg = apibase.parse_request(self.secret, message)
            if 'username' not in req_msg:
                raise ValueError('username is empty')
        except Exception as err:
            resp = apibase.make_response(self.secret, code=1, msg=utils.safestr(err.message))
            self.agent.reply(msgid, resp)
            return
            
        try:
            username = req_msg['username']
            account = get_account_by_username(username)
            if not account:
                apibase.make_response(self.secret, code=1, msg=u'user  {0} not exists'.format(utils.safeunicode(username)))
                self.agent.reply(msgid, resp)
                return
                
            passwd = self.app.aes.decrypt(account.password)
            product = get_product_by_id(account.product_id)

            result = dict(
                code=0,
                msg='success',
                username=username,
                passwd=passwd,
                input_rate=product.input_max_limit,
                output_rate=product.output_max_limit,
                attrs={
                    "Session-Timeout"      : 86400,
                    "Acct-Interim-Interval": 300
                }
            )

            resp = apibase.make_response(self.secret, **result)
            self.agent.reply(msgid, resp)
            self.syslog.info("send auth response %r" % (utils.safeunicode(resp)))
        except Exception as err:
            self.syslog.error(u"api authorize error %s" % utils.safeunicode(err.message))
            resp = apibase.make_response(self.secret, code=1, msg=utils.safestr(err.message))
            return self.agent.reply(msgid, resp)
