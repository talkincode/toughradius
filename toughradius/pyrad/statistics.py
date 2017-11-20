#!/usr/bin/env python
# coding=utf-8
import time
import datetime
import json
from collections import deque


class ComplexEncoder(json.JSONEncoder):
    def default(self, obj):
        if type(obj) == deque:
            return [i for i in obj]
        return json.JSONEncoder.default(self, obj)


class MessageStat(dict):

    def __init__(self,quemax=90):
        self.online = 0
        self.auth_req_old = 0
        self.auth_resp_old = 0
        self.auth_req = 0
        self.auth_accept = 0
        self.auth_reject = 0
        self.auth_drop = 0
        self.acct_start = 0
        self.acct_stop = 0
        self.acct_update = 0
        self.acct_on = 0
        self.acct_off = 0
        self.acct_req_old = 0
        self.acct_resp_old = 0
        self.acct_req = 0
        self.acct_resp = 0
        self.acct_retry = 0
        self.acct_drop = 0
        self.auth_req_stat = deque([],quemax)
        self.auth_resp_stat = deque([],quemax)
        self.acct_req_stat = deque([],quemax)
        self.acct_resp_stat = deque([],quemax)
        self.last_max_req = 0
        self.last_max_req_date = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        self.last_max_resp = 0
        self.last_max_resp_date = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")

    def to_json(self, cls=ComplexEncoder, ensure_ascii=False, **kwargs):
        return json.dumps(self, cls=cls, ensure_ascii=ensure_ascii, **kwargs)

    def incr(self, attr_name, incr=1):
        if hasattr(self, attr_name):
            setattr(self, attr_name, getattr(self,attr_name) + incr)
        
    def run_stat(self,delay=10.0):
        _time = time.time()*1000
        _auth_req_stat = self.auth_req - self.auth_req_old
        self.auth_req_old = self.auth_req

        _auth_resp_stat = (self.auth_accept+self.auth_reject) - self.auth_resp_old
        self.auth_resp_old =  (self.auth_accept+self.auth_reject) 

        _acct_req_stat = self.acct_req - self.acct_req_old
        self.acct_req_old = self.acct_req

        _acct_resp_stat = self.acct_resp - self.acct_resp_old
        self.acct_resp_old = self.acct_resp

        self.auth_req_stat.append((_time,_auth_req_stat))
        self.auth_resp_stat.append((_time,_auth_resp_stat))
        self.acct_req_stat.append((_time,_acct_req_stat))
        self.acct_resp_stat.append((_time,_acct_resp_stat))

        req_percount = int((_auth_req_stat+_acct_req_stat)/delay)
        if self.last_max_req < req_percount:
            self.last_max_req = req_percount
            self.last_max_req_date = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")

        resp_percount = int((_auth_resp_stat+_acct_resp_stat)/delay)
        if self.last_max_resp < resp_percount:
            self.last_max_resp = resp_percount
            self.last_max_resp_date = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")

    def __getattr__(self, key): 
        try:
            return self[key]
        except KeyError, k:
            raise AttributeError, k
    
    def __setattr__(self, key, value): 
        self[key] = value
    
    def __delattr__(self, key):
        try:
            del self[key]
        except KeyError, k:
            raise AttributeError, k        
