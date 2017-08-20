#!/usr/bin/env python
#coding:utf-8
from __future__ import unicode_literals
import json
import shutil
import os
import traceback

def copydir(src, dst, excludes=[]):
    try:
        names = os.walk(src)
        for root, dirs, files in names:
            for i in files:
                srcname = os.path.join(root, i)
                dir = root.replace(src, '')
                dirname = dst + dir
                if os.path.exists(dirname):
                    pass
                else:
                    os.makedirs(dirname)
                dirfname = os.path.join(dirname, i)
                if dirfname not in excludes:
                    shutil.copy2(srcname, dirfname)
    except Exception as e:
        traceback.print_exc()

def safestr(val):
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
    return val

def safeunicode(val):
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
    return val
