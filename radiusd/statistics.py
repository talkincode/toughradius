#coding:utf-8

class RunStat(dict):

    def __init__(self):
        self.online = 0
        self.auth_all = 0
        self.auth_accept = 0
        self.auth_reject = 0
        self.auth_drop = 0
        self.acct_all = 0
        self.acct_start = 0
        self.acct_stop = 0
        self.acct_update = 0
        self.acct_on = 0
        self.acct_off = 0
        self.acct_retry = 0
        self.acct_drop = 0

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

