#!/usr/bin/env python
#coding=utf-8
from twisted.python import log
from toughradius.radiusd.settings import *
import datetime
import decimal

decimal.getcontext().prec = 32
decimal.getcontext().rounding = decimal.ROUND_UP

def get_type_val(typ,src):
    if typ == 'integer' or typ == 'date':
        return int(src)
    else:
        return src

def process(req=None,resp=None,user=None,radiusd=None,**kwargs):
    store = radiusd.store
    product = store.get_product(user['product_id'])
    session_timeout = int(store.get_param("max_session_timeout"))
    acct_policy = user['product_policy'] or BOMonth
    if acct_policy in (PPMonth,BOMonth):
        expire_date = user['expire_date']
        _expire_datetime = datetime.datetime.strptime(expire_date+' 23:59:59',"%Y-%m-%d %H:%M:%S")
        _datetime = datetime.datetime.now()
        if _datetime > _expire_datetime:
            session_timeout += (_expire_datetime - _datetime).seconds 

    elif acct_policy  == BOTimes:
        session_timeout = user["time_length"]
        
    elif acct_policy  == PPTimes:
        user_balance = store.get_user_balance(user['account_number'])
        fee_price = decimal.Decimal(product['fee_price']) 
        _sstime = user_balance/fee_price*decimal.Decimal(3600)
        session_timeout = int(_sstime.to_integral_value())

    if "Framed-Pool" in resp:
        if store.get_param("expire_addrpool") in resp['Framed-Pool']:
            session_timeout = 120
    
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


