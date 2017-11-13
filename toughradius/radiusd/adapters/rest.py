#!/usr/bin/env python
#coding:utf-8

from .base import BasicAdapter
from toughradius.common import tools
from toughradius import settings
from hashlib import md5
import urllib2
import json

class RestError(BaseException):pass

class RestAdapter(BasicAdapter):

    def makeSign(self,message):
        secret = tools.safestr(settings.adapters['rest']['secret'])
        emsg = tools.safestr(message)
        return md5( emsg + secret ).hexdigest()

    def processAuth(self,req):
        url = settings.adapters['rest']['authurl']
        msg = json.dumps(req.dict_message)
        sign = self.makeSign(msg)
        try:
            req = urllib2.Request('%s?sign=%s'%(url,sign),msg)
            resp = urllib2.urlopen(req)
            return json.loads(resp.read())
        except:
            raise RestError("rest request error")


    def processAcct(self,req):
        url = settings.adapters['rest']['accturl']
        msg = json.dumps(req.dict_message)
        sign = self.makeSign(msg)
        try:
            req = urllib2.Request('%s?sign=%s'%(url,sign),msg)
            resp = urllib2.urlopen(req)
            return json.loads(resp.read())
        except:
            raise RestError("rest request error")

