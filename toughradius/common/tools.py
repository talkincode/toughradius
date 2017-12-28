#!/usr/bin/env python
#coding:utf-8
import json
import functools
import os
import time
import logging

logger = logging.getLogger(__name__)

TOUGHRADIUS_TIMECAST_LOG = int(os.environ.get('TOUGHRADIUS_TIMECAST_LOG','0'))

def safestr(val):
    '''
    Convert to string

    :param val: source str

    :return:
    '''
    if val is None:
        return ''

    if isinstance(val, unicode):
        try:
            return val.encode('utf-8')
        except:
            return val.encode('gb2312')
    elif isinstance(val, str):
        return val
    elif isinstance(val, int):
        return str(val)
    elif isinstance(val, float):
        return str(val)
    elif isinstance(val, (dict,list)):
        return json.dumps(val, ensure_ascii=False)
    else:
        try:
            return str(val)
        except:
            return val

def safeunicode(val):
    '''
    Convert to unicode

    :param val:

    :return:
    '''
    if val is None:
        return u''

    if isinstance(val, str):
        try:
            return val.decode('utf-8')
        except:
            try:
                return val.decode('gbk')
            except:
                return val
    elif isinstance(val, unicode):
        return val
    elif isinstance(val, int):
        return str(val).decode('utf-8')
    elif isinstance(val, float):
        return str(val).decode('utf-8')
    elif isinstance(val, (dict,list)):
        return json.dumps(val)
    else:
        try:
            return str(val).decode('utf-8')
        except:
            return val


def timecast(func):
    @functools.wraps(func)
    def warp(*args, **kargs):
        if TOUGHRADIUS_TIMECAST_LOG == 1:
            _start = time.clock()
            result = func(*args, **kargs)
            logger.info("%s.%s cast %.6f second"%(func.__module__, func.__name__, time.clock()-_start))
            return result
        else:
            return func(*args,**kargs)
    return warp