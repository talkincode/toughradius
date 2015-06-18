#!/usr/bin/env python
#coding:utf-8
import json
from cyclone.util import ObjectDict

class ApiMessage(ObjectDict):

    def __init__(self,**kwargs):
        if kwargs.has_key("code"):
            self.code = kwargs.pop("code")
        if kwargs.has_key("msg"):
            self.msg = kwargs.pop("msg","")
        self.rdata = kwargs.pop("rdata",{})

    def set_rval(self, k, v):
        self.rdata[k] = v

    def get_rval(self, k):
        return self.rdata[k]

    @staticmethod
    def parse(jstr):
        return self.__init__(**json.loads(jstr))

    def dumps(self):
        return json.dumps(self, ensure_ascii=False)
