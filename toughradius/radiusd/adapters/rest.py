#!/usr/bin/env python
#coding:utf-8

from base import BasicAdapter
from toughradius.common import tools
from hashlib import md5
import urllib2
import json

class RestError(BaseException):pass

class RestAdapter(BasicAdapter):
    """Http rest mode adapter"""

    def getClients(self):
        nas = dict(status=1, nasid='toughac', name='toughac', vendor=0, ipaddr='127.0.0.1', secret='testing123', coaport=3799)
        return { 'toughac' : nas, '127.0.0.1' : nas}

    def makeSign(self,message):
        secret = tools.safestr(self.settings.ADAPTERS['rest']['secret'])
        emsg = tools.safestr(message)
        return md5( emsg + secret ).hexdigest()

    def processAuth(self,req):
        url = self.settings.ADAPTERS['rest']['authurl']
        msg = json.dumps(req.dict_message)
        sign = self.makeSign(msg)
        try:
            req = urllib2.Request('%s?sign=%s'%(url,sign),msg)
            resp = urllib2.urlopen(req)
            return json.loads(resp.read())
        except:
            raise RestError("rest request error")


    def processAcct(self,req):
        url = self.settings.ADAPTERS['rest']['accturl']
        msg = json.dumps(req.dict_message)
        sign = self.makeSign(msg)
        try:
            req = urllib2.Request('%s?sign=%s'%(url,sign),msg)
            resp = urllib2.urlopen(req)
            return json.loads(resp.read())
        except:
            raise RestError("rest request error")


adapter = RestAdapter