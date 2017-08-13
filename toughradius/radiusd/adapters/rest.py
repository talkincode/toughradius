#!/usr/bin/env python
#coding:utf-8

from .base import BasicAdapter
import grequests

class GrequestRest(BasicAdapter):

    def send_rest(self,req):
        raise NotImplementedError()    