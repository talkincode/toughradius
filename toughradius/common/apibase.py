#!/usr/bin/env python
# coding:utf-8 
import time
import json
from hashlib import md5
from toughradius.common import utils


def mksign(secret, params=[], debug=True):
    _params = [utils.safestr(p) for p in params if p is not None]
    _params.sort()
    _params.insert(0, secret)
    strs = ''.join(_params)
    mds = md5(strs.encode()).hexdigest()
    return mds.upper()


def check_sign(secret, msg, debug=True):
    if "sign" not in msg:
        return False
    sign = msg['sign']
    params = [utils.safestr(msg[k]) for k in msg if k != 'sign']
    local_sign = mksign(secret, params)
    return sign == local_sign

def make_response(secret, **result):
    if 'code' not in result:
        result["code"] = 0
    if 'nonce' not in result:
        result['nonce' ] = str(int(time.time()))
    result['sign'] = mksign(secret, result.values())
    return json.dumps(result, ensure_ascii=False)


def parse_request(secret, reqbody):
    try:
        req_msg = json.loads(reqbody)
    except Exception as err:
        raise ValueError(u"parse params error")

    if not check_sign(secret, req_msg):
        raise ValueError(u"message sign error")

    return req_msg


    