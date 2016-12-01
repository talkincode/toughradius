#!/usr/bin/env python
# coding:utf-8 
import time
import json
from hashlib import md5
from toughradius.common import utils
from toughradius.common import logger
from toughradius.common.storage import Storage
from collections import namedtuple

ApiStatus = namedtuple('ApiStatus', 'code desc msg')  
apistatus = Storage(
    success = ApiStatus(code=0,desc='success',msg=u"处理成功"),
    sign_err = ApiStatus(code=90001,desc='message sign error',msg=u"消息签名错误"),
    parse_err = ApiStatus(code=90002,desc='param parse error',msg=u"参数解析失败"),
    verify_err = ApiStatus(code=90003,desc='message verify error',msg=u"消息校验错误"),
    timeout = ApiStatus(code=90004,desc='request timeout',msg=u"请求超时"),
    limit_err = ApiStatus(code=90005,desc='api limit',msg=u"频率限制"),
    server_err = ApiStatus(code=90006,desc='server process failure',msg=u"服务器处理失败"),
    unknow = ApiStatus(code=99999,desc='unknow error',msg=u"未知错误")
)

class SignError(Exception):
    pass

class ParseError(Exception):
    pass

def make_sign(api_secret, params=[]):
    """
        >>> make_sign("123456",[1,'2',u'中文'])
        '33C9065427EECA3490C5642C99165145'
    """
    _params = [utils.safeunicode(p) for p in params if p is not None]
    _params.sort()
    # print 'sorted params:',_params
    _params.insert(0, api_secret)
    strs = ''.join(_params)
    # print 'sign params:',strs
    mds = md5(strs.encode('utf-8')).hexdigest()
    return mds.upper()


def check_sign(api_secret, msg):
    """
        >>> check_sign("123456",dict(code=1,s='2',msg=u'中文',sign='33C9065427EECA3490C5642C99165145'))
        True

    """
    if "sign" not in msg:
        return False
    sign = msg['sign']
    params = [utils.safestr(msg[k]) for k in msg if k != 'sign' and msg[k] is not None]
    local_sign = make_sign(api_secret, params)
    result = (sign == local_sign)
    if not result:
        logger.error("check_sign failure, sign:%s != local_sign:%s" %(sign,local_sign))
    return result

def make_message(api_secret, enc_func=False, **params):
    """
        >>> json.loads(make_message("123456",**dict(code=1,msg=u"中文",nonce=1451122677)))['sign']
        u'58BAF40309BC1DC51D2E2DC43ECCC1A1'
    """
    if 'nonce' not in params:
        params['nonce' ] = str(int(time.time()))
    params['sign'] = make_sign(api_secret, params.values())
    msg = json.dumps(params, ensure_ascii=False)
    if callable(enc_func):
        return enc_func(msg)
    else:
        return msg

def make_error(api_secret, msg=None, enc_func=False):
    return make_message(api_secret,code=1,msg=msg, enc_func=enc_func)

def parse_request(api_secret, reqbody, dec_func=False):
    """
        >>> parse_request("123456",'{"nonce": 1451122677, "msg": "helllo", "code": 0, "sign": "DB30F4D1112C20DFA736F65458F89C64"}')
        <Storage {u'nonce': 1451122677, u'msg': u'helllo', u'code': 0, u'sign': u'DB30F4D1112C20DFA736F65458F89C64'}>
    """
    try:
        if type(reqbody) == type(dict):
            return self.parse_form_request(reqbody)
            
        if callable(dec_func):
            req_msg = json.loads(dec_func(reqbody))
        else:
            req_msg = json.loads(reqbody)
    except Exception as err:
        raise ParseError(u"parse params error")

    if not check_sign(api_secret, req_msg):
        raise SignError(u"message sign error")

    return Storage(req_msg)

def parse_form_request(api_secret, request):
    """
        >>> parse_form_request("123456",{"nonce": 1451122677, "msg": "helllo", "code": 0, "sign": "DB30F4D1112C20DFA736F65458F89C64"})
        <Storage {'nonce': 1451122677, 'msg': 'helllo', 'code': 0, 'sign': 'DB30F4D1112C20DFA736F65458F89C64'}>
    """
    if not check_sign(api_secret, request):
        raise SignError(u"message sign error")

    return Storage(request)


if __name__ == "__main__":
    # print apistatus
    # import doctest
    # doctest.testmod()
    #isp_name=123&isp_email=222@222.com&isp_idcard=&isp_desc=&isp_phone=22222222222&sign=9E5139D66E8E5C10634A3E96631BCF
    params = ['123','222@222.com','22222222222']
    print make_sign('LpWE9AtfDPQ3ufXBS6gJ37WW8TnSF920',params)

    