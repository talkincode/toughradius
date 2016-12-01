try:
   import cPickle as pickle
except:
   import pickle
from hashlib import md5
import time
import functools
import base64
from twisted.internet import reactor
from twisted.logger import Logger
import redis

CACHE_SET_EVENT = 'cache_set'
CACHE_DELETE_EVENT = 'cache_delete'
CACHE_UPDATE_EVENT = 'cache_update'

class CacheManager(object):
    log = Logger()
    def __init__(self, cache_config,cache_name="cache",stattimes=300,db=0):
        self.cache_name = cache_name
        self.stattimes = stattimes
        self.cache_config = cache_config
        self.redis = redis.StrictRedis(host=cache_config.get('host'), 
            port=cache_config.get("port"), password=cache_config.get('passwd'),db=db)
        self.get_total = 0
        self.set_total = 0
        self.hit_total = 0
        self.miss_total = 0
        self.update_total = 0
        self.delete_total = 0
        # self.print_hit_stat(first_delay=10)
        self.log.info('redis client connected')

    def clean(self):
        self.log.info("flush cache !")
        self.redis.flushdb()
        self.get_total = 0
        self.set_total = 0
        self.hit_total = 0
        self.miss_total = 0
        self.update_total = 0
        self.delete_total = 0

    def print_hit_stat(self, first_delay=0):
        if first_delay > 0:
            reactor.callLater(first_delay, self.print_hit_stat)
            return
            
        logstr = """

----------------------- cache stat ----------------------
#  cache name              : {0}
#  visit cache total       : {1}
#  add cache total         : {2}
#  hit cache total         : {3}
#  miss cache total        : {4}
#  update cache total      : {5}
#  delete cache total      : {6}
#  current db cache total  : {7}
---------------------------------------------------------

""".format(self.cache_name, self.get_total,self.set_total,self.hit_total,self.miss_total,
        self.update_total,self.delete_total,self.count())
        self.log.info(logstr)
        reactor.callLater(self.stattimes, self.print_hit_stat)

    def encode_data(self,data):
        return base64.b64encode(pickle.dumps(data, pickle.HIGHEST_PROTOCOL))

    def decode_data(self,raw_data):
        return pickle.loads(base64.b64decode(raw_data))

    def cache(self,prefix="cache",key_name=None, expire=3600):
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

    def aget(self, key, fetchfunc, *args, **kwargs):
        if self.redis.exists(key):
            result = self.get(key)
            if result:
                return result
            else:
                self.miss_total += 1
                self.log.debug('miss key %s' % key)
        
        expire = kwargs.pop('expire',3600)
        result = fetchfunc(*args,**kwargs)
        if result:
            self.set(key,result,expire=expire)
        return result

    def exists(self, key):
        return self.redis.exists(key)

    def get(self, key):
        self.get_total += 1
        try:
            raw_data = self.redis.get(key)
            if raw_data:
                self.hit_total += 1
                return self.decode_data(raw_data)
            else:
                self.miss_total += 1
                self.log.debug('miss key %s' % key)
        except:
            self.delete(key)
        return None



    def event_cache_delete(self, key):
        self.log.info("event: delete cache %s " % key)
        self.delete(key)

    def count(self):
        return self.redis.dbsize()

    def delete(self,key):
        self.delete_total += 1
        self.redis.delete(key)

    def event_cache_set(self, key, value, expire=3600):
        self.log.info("event: set cache %s " % key)
        self.set(key, value, expire)

    def set(self, key, value, expire=3600):
        self.set_total += 1
        raw_data = self.encode_data(value)
        self.redis.setex(key,expire,raw_data)
     
    def event_cache_update(self, key, value, expire=3600):
        self.log.info("event: update cache %s " % key)
        self.update(key, value, expire)

    def update(self, key, value, expire=3600):
        self.update_total += 1
        raw_data = self.encode_data(value)
        self.redis.setex(key,expire,raw_data)

if __name__ == '__main__':
    pass


