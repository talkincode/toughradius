#!/usr/bin/env python
#coding:utf-8

from .base import BasicAdapter
from toughradius.common import tools
from hashlib import md5
import urllib2
import json

class RestError(BaseException):pass

class RestAdapter(BasicAdapter):

    def makeSign(self,message):
        secret = tools.safestr(self.config.adapters.rest.secret)
        emsg = tools.safestr(message)
        return md5( emsg + secret ).hexdigest()

    def send(self,req):
        url = self.config.adapters.rest.url
        msg = json.dumps(req.dict_message)
        sign = self.makeSign(msg)
        try:
            req = urllib2.Request('%s?sign=%s'%(url,sign),msg)
            resp = urllib2.urlopen(req)
            return json.loads(resp.read())
        except:
            self.logger.exception("send rest request error")
            raise RestError("rest request error")

