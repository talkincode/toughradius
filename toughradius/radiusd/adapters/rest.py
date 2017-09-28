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

    def auth(self,req):
        url = self.config.adapters.rest.authurl
        msg = json.dumps(req.dict_message)
        sign = self.makeSign(msg)
        try:
            req = urllib2.Request('%s?sign=%s'%(url,sign),msg)
            resp = urllib2.urlopen(req)
            return json.loads(resp.read())
        except:
            raise RestError("rest request error")


    def acct(self,req):
        url = self.config.adapters.rest.accturl
        msg = json.dumps(req.dict_message)
        sign = self.makeSign(msg)
        try:
            req = urllib2.Request('%s?sign=%s'%(url,sign),msg)
            resp = urllib2.urlopen(req)
            return json.loads(resp.read())
        except:
            raise RestError("rest request error")

