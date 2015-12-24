import pickle
from hashlib import md5
import time
import functools
from sqlalchemy.sql import text as _sql
from twisted.internet import reactor
from toughradius.common.utils import timecast

class CacheManager(object):
    def __init__(self, dbengine):
        self.dbengine = dbengine

    def cache(self,prefix="cache", expire=3600):
        def func_warp1(func):
            @functools.wraps(func)
            def func_wrap2(*args, **kargs):
                sig = md5(repr(args) + repr(kargs)).hexdigest()
                key = "%s:%s:%s"%(prefix,func.__name__, sig)
                data = self.get(key)
                if data is not None:
                    return data
                data = func(*args, **kargs)
                if data is not None:
                    self.set(key, data)
                return data
            return func_wrap2
        return func_warp1


    def get(self, key):
        raw_data = None
        _del_func = self.delete
        with self.dbengine.begin() as conn:
            try:
                cur = conn.execute(_sql("select _value, _time from system_cache where _key = :key "),key=key)
                _cache =  cur.fetchone()
                if _cache:
                    _time = int(_cache['_time'])
                    if _time > 0 and time.time() > _time:
                        reactor.callLater(0.01, _del_func, key,)
                    else:
                        raw_data = _cache['_value']
            except:
                import traceback
                traceback.print_exc()
        return raw_data and pickle.loads(raw_data) or None

    def delete(self,key):
        with self.dbengine.begin() as conn:
            try:
                conn.execute(_sql("delete from system_cache where _key = :key "),key=key)
            except:
                import traceback
                traceback.print_exc()


    def set(self, key, value, expire=0):
        raw_data = pickle.dumps(value, pickle.HIGHEST_PROTOCOL)
        with self.dbengine.begin() as conn:
            _time = expire>0 and (int(time.time()) + int(expire)) or 0
            try:
                conn.execute(_sql("insert into system_cache values (:key, :value, :time) "),
                    key=key,value=raw_data,time=_time)
            except:
                self.update(key,value,expire)


    def update(self, key, value, expire=0):
        raw_data = pickle.dumps(value, pickle.HIGHEST_PROTOCOL)
        with self.dbengine.begin() as conn:
            _time = expire>0 and (int(time.time()) + int(expire)) or 0
            try:
                conn.execute(_sql("""update system_cache 
                                    set _value=:value, _time=:time
                                    where _key=:key"""),
                                    key=key,value=raw_data,time=_time)
            except:
                import traceback
                traceback.print_exc()





