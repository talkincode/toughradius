#!/usr/bin/env python
#coding:utf-8

from base import BasicAdapter


class RestError(BaseException):pass

class FreeAdapter(BasicAdapter):
    """Free auth mode"""

    def getClient(self, nasip=None, nasid=None):
        return dict(status=1, nasid='toughac', name='toughac', vendor=0, ipaddr='127.0.0.1', secret='testing123', coaport=3799)

    def processAuth(self,req):
        return dict(code=0, msg='ok')


    def processAcct(self,req):
        return dict(code=0, msg='ok')

adapter = FreeAdapter