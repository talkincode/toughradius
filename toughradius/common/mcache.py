#!/usr/bin/env python
# coding=utf-8

import time

class Mcache:

    def __init__(self):
        self.cache = {}

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




