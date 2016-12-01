try:
   import cPickle as pickle
except:
   import pickle
from hashlib import md5
import time
import functools
import base64
from sqlalchemy.sql import text as _sql
from twisted.internet import reactor
from twisted.logger import Logger

CACHE_SET_EVENT = 'cache_set'
CACHE_DELETE_EVENT = 'cache_delete'
CACHE_UPDATE_EVENT = 'cache_update'

class CacheManager(object):
    log = Logger()
    def __init__(self, dbengine,cache_table='system_cache',cache_name="cache"):
        self.cache_name = cache_name
        self.dbengine = dbengine
        self.cache_table = cache_table
        self.get_total = 0
        self.set_total = 0
        self.hit_total = 0
        self.update_total = 0
        self.delete_total = 0
        self.check_expire(first_delay=10)
        self.print_hit_stat(first_delay=10)

    def print_hit_stat(self, first_delay=0):
        if first_delay > 0:
            reactor.callLater(first_delay, self.check_expire)
        logstr = """

----------------------- cache stat ----------------------
#  cache name              : {0}
#  visit cache total       : {1}
#  add cache total         : {2}
#  hit cache total         : {3}
#  update cache total      : {4}
#  delete cache total      : {5}
#  current db cache total  : {6}
---------------------------------------------------------

""".format(self.cache_name, self.get_total,self.set_total,self.hit_total,
        self.update_total,self.delete_total,self.count())
        self.log.info(logstr)
        reactor.callLater(60.0, self.print_hit_stat)

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
        result = self.get(key)
        if result:
            return result
        if fetchfunc:
            expire = kwargs.pop('expire',3600)
            result = fetchfunc(*args,**kwargs)
            if result:
                self.set(key,result,expire=expire)
            return result

    def check_expire(self, first_delay=0):
        if first_delay > 0:
            reactor.callLater(first_delay, self.check_expire)
        with self.dbengine.begin() as conn:
            try:
                conn.execute(_sql("delete from %s where _time > 0 and _time < :time" % self.cache_table),time=int(time.time()))
            except:
                pass
        reactor.callLater(120.0, self.check_expire)

    def get(self, key):
        self.get_total += 1
        raw_data = None
        _del_func = self.delete
        with self.dbengine.begin() as conn:
            try:
                cur = conn.execute(_sql("select _value, _time from %s where _key = :key " % self.cache_table),key=key)
                _cache =  cur.fetchone()
                if _cache:
                    self.hit_total += 1
                    _time = int(_cache['_time'])
                    if _time > 0 and time.time() > _time:
                        reactor.callLater(0.01, _del_func, key,)
                    else:
                        raw_data = _cache['_value']
            except:
                import traceback
                traceback.print_exc()

        try:
            if raw_data:
                return self.decode_data(raw_data)
        except:
            self.delete(key)
        return None



    def event_cache_delete(self, key):
        self.log.info("event: delete cache %s " % key)
        self.delete(key)

    def count(self):
        with self.dbengine.begin() as conn:
            try:
                cur = conn.execute(_sql("select count(_key) as count from %s " % self.cache_table))
                return int(cur.fetchone()['count'])
            except:
                self.log.error("cache count error")
                return 0

    def delete(self,key):
        self.delete_total += 1
        with self.dbengine.begin() as conn:
            try:
                conn.execute(_sql("delete from %s where _key = :key " % self.cache_table),key=key)
            except:
                import traceback
                traceback.print_exc()

    def event_cache_set(self, key, value, expire=0):
        self.log.info("event: set cache %s " % key)
        self.set(key, value, expire)

    def set(self, key, value, expire=0):
        self.set_total += 1
        raw_data = self.encode_data(value)
        with self.dbengine.begin() as conn:
            _time = expire>0 and (int(time.time()) + int(expire)) or 0
            try:
                conn.execute(_sql("insert into %s values (:key, :value, :time) " % self.cache_table),
                    key=key,value=raw_data,time=_time)
            except:
                conn.execute(_sql("delete from %s where _key = :key " % self.cache_table),key=key)
                conn.execute(_sql("insert into %s values (:key, :value, :time) " % self.cache_table),
                    key=key,value=raw_data,time=_time)
                
    def event_cache_update(self, key, value, expire=0):
        self.log.info("event: update cache %s " % key)
        self.update(key, value, expire)

    def update(self, key, value, expire=0):
        self.update_total += 1
        raw_data = self.encode_data(value)
        with self.dbengine.begin() as conn:
            _time = expire>0 and (int(time.time()) + int(expire)) or 0
            try:
                conn.execute(_sql("""update %s 
                                    set _value=:value, _time=:time
                                    where _key=:key""" % self.cache_table),
                                    key=key,value=raw_data,time=_time)
            except:
                import traceback
                traceback.print_exc()




