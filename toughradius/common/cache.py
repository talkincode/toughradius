#!/usr/bin/env python
# coding=utf-8

import time
from hashlib import md5

class Mcache:

    def __init__(self):
        self.cache = {}

    def cache(self,prefix="cache",key_name=None, expire=600):
        def func_warp1(func):
            @functools.wraps(func)
            def func_wrap2(*args, **kargs):
                if key_name and kargs.get(key_name):
                    key = "%s:%s" % (prefix, kargs.get(key_name))
                else:
                    sig = md5(repr(args) + repr(kargs)).hexdigest()
                    key = "%s:%s:%s"%(prefix,func.__name__, sig)

                data = self.get(key)
                if data is not None:
                    return data
                data = func(*args, **kargs)
                if data is not None:
                    self.set(key, data, expire)
                return data
            return func_wrap2
        return func_warp1        

    def set(self, key, obj, expire=0):
        if obj in ("", None) or key in ("", None):
            return None

        objdict = dict(
            obj=obj,
            expire=expire,
            time=time.time()
        )

        self.cache[key] = objdict


    def get(self, key):
        if key in self.cache:
            objdict = self.cache[key]
            _time = time.time()
            if objdict['expire'] == 0 or (_time - objdict['time']) < objdict['expire']:
                return objdict['obj']
            else:
                del self.cache[key]
                return None
        else:
            return None


    def aget(self, key, fetchfunc, *args, **kwargs):
        if key in self.cache:
            return self.get(key)
        elif fetchfunc:
            expire = kwargs.pop('expire',600)
            result = fetchfunc(*args,**kwargs)
            self.set(key,result,expire=expire)
            return result




