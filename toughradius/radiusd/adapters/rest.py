#!/usr/bin/env python
#coding:utf-8

from .base import BasicAdapter
from toughradius.common import tools
from hashlib import md5
import grequests
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
            resp = grequests.post('%s?sign=%s'%(url,sign),data=msg)
        except:
            resp = grequests.post('%s?sign=%s'%(url,sign),data=msg,verify=False)

        if resp.status_code == 200:
            return resp.json()
        else:
            raise  RestError("rest request error")