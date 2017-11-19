#!/usr/bin/env python
#coding:utf-8

from .base import BasicAdapter
from toughradius.common import tools
from hashlib import md5
import urllib2
import json

class RestError(BaseException):pass

class FreeAdapter(BasicAdapter):
    """Free auth mode"""

    def getClients(self):
        nas = dict(status=1, nasid='toughac', name='toughac', vendor=0, ipaddr='127.0.0.1', secret='testing123', coaport=3799)
        return { 'toughac' : nas, '127.0.0.1' : nas}

    def processAuth(self,req):
        return dict(code=0, msg='ok')


    def processAcct(self,req):
        return dict(code=0, msg='ok')

adapter = FreeAdapter