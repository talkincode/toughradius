#!/usr/bin/env python
#coding:utf-8
import os
import sys
import uuid
from hashlib import md5
from twisted.internet import utils as txutils

def get_sys_uuid():
    fs = '/sys/class/dmi/id/product_uuid'
    if os.path.exists(fs):
        _file = open(fs)
        _uuid = _file.read()
        _file.close()
        return md5(_uuid).hexdigest()
    else:
        mac=uuid.UUID(int = uuid.getnode()).hex[-12:] 
        return md5(":".join([mac[e:e+2] for e in range(0,11,2)])).hexdigest()
    return md5('free-uuid').hexdigest()

def get_sys_token():
    return txutils.getProcessOutput("/usr/local/bin/toughkey")