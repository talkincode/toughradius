#!/usr/bin/env python
#coding=utf-8
from toughradius.console.mps import base
from toughradius.console.mps import mpsmsg
from hashlib import sha1
from cyclone.util import ObjectDict
from twisted.python import log
from toughradius.console.mps.mpsmsg import (
    MSG_TYPE_TEXT, 
    MSG_TYPE_LOCATION, 
    MSG_TYPE_IMAGE, 
    MSG_TYPE_LINK, 
    MSG_TYPE_EVENT, 
    MSG_TYPE_MUSIC, 
    MSG_TYPE_NEWS,
    MSG_TYPE_CUSTOMER
)

ACCESS_TOKEN = ''
TOKEN_TIMEOUT = 0


class IndexHandler(base.BaseHandler):
    """ 微信消息主要处理控制器 """

    def check_xsrf_cookie(self):
        """ 对于微信消息不做加密cookie处理 """

    def get_error_html(self, status_code=500, **kwargs):
        if self.settings['debug']:
            import traceback
            return self.render_json(code=1, msg=traceback.format_exc())
        return self.render_json(code=1, msg=u"%s:服务器处理失败，请联系管理员" % status_code)


    def check_signature(self):
        """ 微信消息验签处理 """
        if self.settings.test:
            return True
        signature = self.get_argument('signature', '')
        timestamp = self.get_argument('timestamp', '')
        nonce = self.get_argument('nonce', '')
        tmparr = [self.settings.mps_token, timestamp, nonce]
        tmparr.sort()
        tmpstr = ''.join(tmparr)
        tmpstr = sha1(tmpstr).hexdigest()
        return tmpstr == signature

    def get(self):
        echostr = self.get_argument('echostr', '')
        if self.check_signature():
            self.write(echostr)
            log.msg("Signature check success.")
        else:
            self.logging.warning("Signature check failed.")


    def post(self):
        """ 微信消息处理 """
        if not self.check_signature():
            log.msg("Signature check failed.")
            return

        self.set_header("Content-Type", "application/xml;charset=utf-8")
        body = self.request.body
        msg = mpsmsg.parse_msg(body)
        if not msg:
            log.msg('Empty message, ignored')
            return

        log.msg(u'message type %s from %s with %s'%(
            msg.type,msg.fromuser, body.decode("utf-8")))


        reply_msg = self.process(msg)
        log.msg(u'Replied to %s with "%s"'% (msg.fromuser, reply_msg))

        self.write(reply_msg)
        self.finish()

    def process(self,msg):
        if msg.type == MSG_TYPE_TEXT:
            return mpsmsg.gen_reply(msg.touser, msg.fromuser, self.process_text(msg))
        elif msg.type == MSG_TYPE_LOCATION:
            return mpsmsg.gen_reply(msg.touser, msg.fromuser, self.process_location(msg))
        elif msg.type == MSG_TYPE_IMAGE:
            return mpsmsg.gen_reply(msg.touser, msg.fromuser, self.process_image(msg))
        elif msg.type == MSG_TYPE_EVENT:
            return mpsmsg.gen_reply(msg.touser, msg.fromuser, self.process_event(msg))
        elif msg.type == MSG_TYPE_LINK:
            return mpsmsg.gen_reply(msg.touser, msg.fromuser, self.process_link(msg))
        else:
            log.msg('message type unknown')

    def process_response(self,result):
        if isinstance(result, list):
            return ObjectDict(msg_type=MSG_TYPE_NEWS, response=result)
        if isinstance(result, dict):
            if result.get('msg_type') == 'transfer_customer_service':
                return ObjectDict(msg_type=MSG_TYPE_CUSTOMER,kfaccount=result['kfaccount'])
        else:
            return ObjectDict(msg_type=MSG_TYPE_TEXT, response=mpsmsg.decode(result))

    def process_text(self,msg):
        result = self.middware.respond(
            msg.content, msg=msg, db=self.db,config=self.config,mpsapi=self.mpsapi)
        log.msg(u'bot response %s'%result)
        return self.process_response(result)

    def process_event(self,msg):
        _event_key = 'event:%s:%s' % (msg.event, msg.event_key)
        log.msg('event -> %s'%_event_key)
        result = self.middware.respond(
            _event_key, msg=msg, db=self.db,config=self.config,mpsapi=self.mpsapi)

        log.msg(u'bot response %s'%result)

        return self.process_response(result)

    def process_location(self,msg):
        return self.process_nothing()

    def process_image(self,msg):
        result = self.middware.respond(
            'image:message', msg=msg, db=self.db,config=self.config,mpsapi=self.mpsapi)

        log.msg(u'bot response %s'%result)
            
        return self.process_response(result)

    def process_link(self,msg):
        return self.process_nothing()

    def process_nothing(self):
        return ObjectDict(msg_type=MSG_TYPE_TEXT, response=u'')





