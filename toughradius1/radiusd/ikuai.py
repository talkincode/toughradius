#!/usr/bin/env python
#coding=utf-8
import struct
from toughradius.radiusd.pyrad import  tools

IK_RAD_PKG_VER            = 0x0001
IK_RAD_PKG_AUTH           = 0x0001
IK_RAD_PKG_USR_PWD_TAG    = 0x0011
IK_RAD_PKG_USR_NAME_TAG   = 0x0012
IK_RAD_PKG_USR_CMD_TAG    = 0x0013
IK_RAD_PKG_CMD_ARGS_TAG   = 0x0014
IK_RAD_PKG_USR_CMD_LEN    = 4
IK_RAD_MAX_TLV            = 10
IK_RAD_STRING_MAX_LENGTH  = 32 
IK_STR_MEM_LEN            = 40

VENDOR_ID = 10055


def create_dm_pkg(secret,username):
    ''' create ikuai dm message'''
    secret = tools.EncodeString(secret)
    username = tools.EncodeString(username)
    pkg_format = '>HHHH32sHH32s'
    pkg_vals = [
        IK_RAD_PKG_VER,
        IK_RAD_PKG_AUTH,
        IK_RAD_PKG_USR_PWD_TAG,
        len(secret),
        secret.ljust(32,'\x00'),
        IK_RAD_PKG_CMD_ARGS_TAG,
        len(username),
        username.ljust(32,'\x00')
    ]
    return struct.pack(pkg_format,*pkg_vals)

if __name__ == '__main__':
    pp = create_dm_pkg('123456', 'testuser')
    print repr(pp)