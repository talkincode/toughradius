#!/usr/bin/env python
#coding:utf-8
import os
import sys
import time
import datetime
from urllib import urlencode
from toughlib import utils
from twisted.internet import defer
from cyclone import httpclient
from toughlib import dispatch,logger
from toughradius.manage.tasks.task_base import TaseBasic
from toughradius.manage.settings import toughcloud_ping_key
from toughradius.manage import taskd
from toughradius.common import tools

class ToughCloudPingTask(TaseBasic):

    __name__ = 'toughcloud-ping'    

    def __init__(self,taskd, **kwargs):
        TaseBasic.__init__(self,taskd, **kwargs)      

    def get_notify_interval(self):
        return 300

    def first_delay(self):
        return self.get_notify_interval()

    @defer.inlineCallbacks
    def process(self, *args, **kwargs):
        next_interval = self.get_notify_interval()
        try:
            api_url = "https://www.toughcloud.net/api/v1/ping"
            api_token = yield tools.get_sys_token()
            param_str = urlencode({'token':api_token})
            resp = yield httpclient.fetch(api_url+"?"+param_str,followRedirect=True)
            logger.info("toughcloud ping resp code: %s"%resp.code)
            if resp.code == 200:
                self.cache.set(toughcloud_ping_key,resp.body,expire=3600)
        except Exception as err:
            logger.error(err)
        defer.returnValue(next_interval)

taskd.TaskDaemon.__taskclss__.append(ToughCloudPingTask)

