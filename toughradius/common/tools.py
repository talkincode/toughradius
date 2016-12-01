#!/usr/bin/env python
#coding:utf-8
import os
import sys
import uuid
from hashlib import md5
from twisted.internet import utils as txutils
import shutil


def get_sys_uuid():
    mac=uuid.UUID(int = uuid.getnode()).hex[-12:] 
    return md5(":".join([mac[e:e+2] for e in range(0,11,2)])).hexdigest()

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
        import traceback
        traceback.print_exc()    

if __name__ == '__main__':
    print get_sys_uuid()