#!/usr/bin/env python
#coding:utf-8
from toughlib import config as iconfig
import os
import requests

class TestMixin:

    MANAGE_URL = 'http://127.0.0.1:18160'

    def sub_path(self,path):
        return "%s%s"%(TestMixin.MANAGE_URL,path)

    def init_rundir(self):
        try:
            os.mkdir("/tmp/toughradius")
        except:
            print "/tmp/toughradius exists"

    def init_config(self):
        testfile = os.path.join(os.path.abspath(os.path.dirname(__file__)),"test.json")
        self.config = iconfig.find_config(testfile)

    def admin_login(self):
        req = requests.Session()
        r = req.post(self.sub_path("/admin/login"),data=dict(username="admin",password="root"))
        if r.status_code == 200:
            rjson =  r.json()
            msg = rjson['msg']
            if rjson['code'] == 0:
                return req
            else:
                raise Exception(msg)
        else:
            r.raise_for_status()