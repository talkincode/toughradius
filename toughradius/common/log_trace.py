#!/usr/bin/env python
# coding=utf-8

from toughlib import utils
from toughlib import logger
try:
    import redis
except:
    pass


class LogTrace(object):

    radius_key = "toughradius.syslog.trace.radius.{0}".format,
    trace_key = "toughradius.syslog.trace.{0}".format,

    def __init__(self, cache_config,db=2):
        self.cache_config = cache_config
        self.redis = redis.StrictRedis(host=cache_config.get('host'), 
            port=cache_config.get("port"), password=cache_config.get('passwd'),db=db)
        logger.info('LogTrace connected')

    def count(self):
        return self.redis.dbsize()    

    def clean(self):
        logger.info('clear system trace')
        return self.redis.flushdb()

    def trace_radius(self,username,message):
        key = self.radius_key(username)
        if self.redis.llen(key) >= 64:
            self.redis.ltrim(key,0,63)
        self.redis.lpush(key,message)

    def trace_log(self,name,message):
        key = self.trace_key(name)
        if self.redis.llen(key) >= 256:
            self.redis.ltrim(key,0,255)
        self.redis.lpush(key,message)

    def list_radius(self,username):
        key = self.radius_key(username)
        return [utils.safeunicode(v) for v in self.redis.lrange(key,0,31) ]

    def list_trace(self,name):
        key = self.trace_key(name)
        return [utils.safeunicode(v) for v in self.redis.lrange(key,0,255) ]

    def delete_radius(self,username):
        key = self.radius_key(username)
        return self.redis.delete(key)

    def delete_trace(self,name):
        key = self.trace_key(name)
        return self.redis.delete(key)

    def event_syslog_trace(self, name, message,**kwargs):
        """ syslog trace event """
        if name == 'radius' and 'username' in kwargs:
            self.trace_radius(kwargs['username'], message)
        else:
            self.trace_log(name, message)

if __name__ == '__main__':
    pass


