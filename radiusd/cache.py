#coding:utf-8
import hashlib
import functools

_cache = {}

def _mk_cache_sig(*args, **kwargs):
    src_data = repr(args) + repr(kwargs)
    m = hashlib.md5(src_data)
    sig = m.hexdigest()
    return sig

def cache(category='all'):
    def func_warp1(func):
        @functools.wraps(func)
        def func_wrap2(*args, **kwargs):
            sig = _mk_cache_sig(*args, **kwargs)
            key = "%s:%s:%s"%(category,func.__name__, sig)
            data = _cache.get(key)
            if data is not None:
                return data
            data = func(*args, **kwargs)
            if data is not None:
                _cache[key] = data
            return data
        return func_wrap2
    return func_warp1

def delete(category,func,*args,**kwargs):
    sig = _mk_cache_sig(*args, **kwargs)
    key = "%s:%s:%s"%(category,func.__name__, sig)
    if key in _cache:
        del _cache[key]

def clear():
    _cache.clear()

if __name__ == '__main__':
    
    @cache('all')
    def getcache(p1,p2,p3=4):
        return 'cache data %s %s %s'%(p1,p2,p3)

    print getcache(1,2,3)
    print _cache
    delete('all',getcache,1,2,3)
    print _cache
    print getcache(1,2,3)
    # clear()
    print _cache










