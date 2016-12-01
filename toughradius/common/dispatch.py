#!/usr/bin/env python
# coding=utf-8
import os
import types
import importlib
from twisted.internet.threads import deferToThread
from twisted.python import reflect
from twisted.internet import defer
from twisted.python import log

class EventDispatcher:

    def __init__(self, prefix="event_"):
        self.prefix = prefix
        self.callbacks = {}

    def sub(self, name, func, check_exists=False):
        if check_exists and name in self.callbacks:
            return
        self.callbacks.setdefault(name, []).append(func)
        log.msg('register event %s %s' % (name,(func.__doc__ or '')))

    def register(self, obj, check_exists=False):
        d = {}
        reflect.accumulateMethods(obj, d, self.prefix)
        for k,v in d.items():
            self.sub(k, v, check_exists=check_exists)

    def pub(self, name, *args, **kwargs):
        if name not in self.callbacks:
            return
        async = kwargs.pop("async",False)
        results = []
        for func in self.callbacks[name]:
            if async:
                deferd = deferToThread(func, *args, **kwargs)
                deferd.addErrback(log.err)
                results.append(deferd)
            else:
                results.append(func(*args, **kwargs))
        if async:
            return defer.DeferredList(results)
        else:
            return results


dispatch = EventDispatcher()
sub = dispatch.sub
pub = dispatch.pub
register = dispatch.register

def load_events(event_path=None,pkg_prefix=None,excludes=[],event_params={}):
    _excludes = ['__init__','settings','.DS_Store'] + excludes
    evs = set(os.path.splitext(it)[0] for it in os.listdir(event_path))
    evs = [it for it in evs if it not in _excludes]
    for ev in evs:
        try:
            sub_module = os.path.join(event_path, ev)
            if os.path.isdir(sub_module):
                load_events(
                    event_path=sub_module,
                    pkg_prefix="{0}.{1}".format(pkg_prefix, ev),
                    excludes=excludes,
                    event_params=event_params,
                )
            _ev = "{0}.{1}".format(pkg_prefix, ev)
            robj = importlib.import_module(_ev)
            if hasattr(robj, 'evobj'):
                dispatch.register(robj.evobj)
            if hasattr(robj, '__call__'):
                dispatch.register(robj.__call__(**event_params))
        except Exception as err:
            import traceback
            traceback.print_exc()
            continue




