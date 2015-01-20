#!/usr/bin/env python
#coding=utf-8
from twisted.python import log
from store import store
from settings import *
import datetime

def get_type_val(typ,src):
    if typ == 'integer' or typ == 'date':
        return int(src)
    else:
        return src

def process(req=None,resp=None,user=None):
    product = store.get_product(user['product_id'])
    session_timeout = int(store.get_param("max_session_timeout"))
    acct_policy = user['product_policy'] or FEE_BUYOUT
    if acct_policy in (FEE_BUYOUT,FEE_MONTH):
        expire_date = user.get('expire_date')
        _expire_datetime = datetime.datetime.strptime(expire_date+' 23:59:59',"%Y-%m-%d %H:%M:%S")
        _datetime = datetime.datetime.now()
        if _datetime > _expire_datetime:
            session_timeout += (_expire_datetime - _datetime).seconds 
    elif acct_policy == FEE_TIMES:
        balance = int(user.get("balance",0))
        if balance == 0:
            session_timeout = 0
        else:
            time_len = balance * 3600 / product['fee_price']
            session_timeout = time_len
    resp['Session-Timeout'] = session_timeout

    input_limit = str(product['input_max_limit'])
    output_limit = str(product['output_max_limit'])
    _class = input_limit.zfill(8) + input_limit.zfill(8) + output_limit.zfill(8) + output_limit.zfill(8)
    resp['Class'] = _class

    if user['ip_address']:
        resp['Framed-IP-Address'] = user['ip_address']

    for attr in store.get_product_attrs(user['product_id']):
        try:
            _type = resp.dict[attr['attr_name']].type
            print _type
            resp[str(attr['attr_name'])] = get_type_val(_type,attr['attr_value'])
        except:
            import traceback
            traceback.print_exc()

    
    # for attr in store.get_user_attrs(user['account_number']):
    #     try:resp[attr.attr_name] = attr.attr_value
    #     except:pass

    return resp


