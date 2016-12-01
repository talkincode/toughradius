#!/usr/bin/env python
#coding=utf-8
import time

class ValidateCache(object):

    def __init__(self, max_times=5):
        self.max_times = max_times

    validates = {}
    def incr(self,mid,vid):
        key = "%s_%s"%(mid,vid)
        if key not in self.validates:
            self.validates[key] = [1,time.time()]
        else:
            self.validates[key][0] += 1
            
    def errs(self,mid,vid):
        key = "%s_%s"%(mid,vid)    
        if key in  self.validates:
            return self.validates[key][0] 
        return 0
    
    def clear(self,mid,vid):
        key = "%s_%s"%(mid,vid)    
        if key in  self.validates:
            del self.validates[key]
        
    def is_over(self,mid,vid):
        key = "%s_%s"%(mid,vid)
        if key not in self.validates:
            return False
        elif (time.time() - self.validates[key][1]) > 3600:
            del self.validates[key]
            return False
        else:
            return self.validates[key][0] >= self.max_times

vcache = ValidateCache() 