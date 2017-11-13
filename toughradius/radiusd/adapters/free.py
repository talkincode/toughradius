#!/usr/bin/env python
#coding:utf-8

from .base import BasicAdapter
from toughradius.common import tools
from hashlib import md5
import urllib2
import json

class RestError(BaseException):pass

class FreeAdapter(BasicAdapter):

    def processAuth(self,req):
        return dict(code=0, msg='ok')


    def processAcct(self,req):
        return dict(code=0, msg='ok')

