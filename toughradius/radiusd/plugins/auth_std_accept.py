#!/usr/bin/env python
#coding=utf-8
from toughradius.radiusd.plugins import error_auth
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
    session_timeout = int(store.get_param("max_session_timeout"))

    if store.is_white_roster(req.get_mac_addr()):
        resp['Session-Timeout'] = session_timeout
        return resp

    expire_pool = store.get_param("expire_addrpool")
    if "Framed-Pool" in resp:
        if expire_pool in resp['Framed-Pool']:
            expire_session_timeout = int(store.get_param("expire_session_timeout"))
            if expire_session_timeout > 0:
                session_timeout = expire_session_timeout
            else:
                return error_auth(resp,'User has expired')

    acct_interim_intelval = int(store.get_param("acct_interim_intelval"))
    if acct_interim_intelval > 0:
        resp['Acct-Interim-Interval'] = acct_interim_intelval
    
    acct_policy = user['product_policy'] or BOMonth
    product = store.get_product(user['product_id'])
    
    if acct_policy in (PPMonth,BOMonth):
        expire_date = user['expire_date']
        _datetime = datetime.datetime.now()
        if _datetime.strftime("%Y-%m-%d") == expire_date:
            _expire_datetime = datetime.datetime.strptime(expire_date+' 23:59:59',"%Y-%m-%d %H:%M:%S")
            session_timeout = (_expire_datetime - _datetime).seconds 

    elif acct_policy  == BOTimes:
        _session_timeout = user["time_length"]
        if _session_timeout < session_timeout:
            session_timeout = _session_timeout
        
    elif acct_policy  == PPTimes:
        user_balance = store.get_user_balance(user['account_number'])
        fee_price = decimal.Decimal(product['fee_price']) 
        _sstime = user_balance/fee_price*decimal.Decimal(3600)
        _session_timeout = int(_sstime.to_integral_value())
        if _session_timeout < session_timeout:
            session_timeout = _session_timeout

    resp['Session-Timeout'] = session_timeout

    if user['ip_address']:
        resp['Framed-IP-Address'] = user['ip_address']

    _attrs = {}
    for attr in store.get_product_attrs(user['product_id']):
        try:
            _type = resp.dict[attr['attr_name']].type
            attr_name = str(attr['attr_name'])
            attr_value = get_type_val(_type,attr['attr_value'])
            if attr_name in _attrs:
                _attrs[attr_name].append(attr_value)
            else:
                _attrs[attr_name] = [attr_value]
        except:
            import traceback
            traceback.print_exc()
    
    print _attrs
    for _a in _attrs:        
        resp.AddAttribute(_a,_attrs[_a])
    
    for attr in req.ext_attrs:
        resp[attr] = req.ext_attrs[attr]
    # for attr in store.get_user_attrs(user['account_number']):
    #     try:resp[attr.attr_name] = attr.attr_value
    #     except:pass

    return resp


