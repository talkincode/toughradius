#!/usr/bin/env python
#coding:utf-8
import os
import sys
import uuid
from hashlib import md5
from twisted.internet import utils as txutils

def get_sys_uuid():
    mac=uuid.UUID(int = uuid.getnode()).hex[-12:] 
    return md5(":".join([mac[e:e+2] for e in range(0,11,2)])).hexdigest()


def get_sys_token():
    return txutils.getProcessOutput("/usr/local/bin/toughkey")

if __name__ == '__main__':
    print get_sys_uuid()