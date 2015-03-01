#!/usr/bin/env python
#coding=utf-8
from toughradius.radiusd.plugins import error_auth
from toughradius.radiusd.settings import *
from toughradius.radiusd import utils

def process(req=None,resp=None,user=None,radiusd=None,**kwargs):
    """执行计费策略校验，用户到期检测，用户余额，时长检测"""
    store = radiusd.store
    acct_policy = user['product_policy'] or PPMonth
    if acct_policy in ( PPMonth,BOMonth):
        if utils.is_expire(user['expire_date']):
            resp['Framed-Pool'] = store.get_param("expire_addrpool")
            
    elif acct_policy in (PPTimes,PPFlow):
        user_balance = store.get_user_balance(user['account_number'])
        if user_balance <= 0:
            return error_auth(resp,'user balance poor')    
            
    elif acct_policy == BOTimes:
        time_length = store.get_user_time_length(user['account_number'])
        if time_length <= 0:
            return error_auth(resp,'user time_length poor')
            
    elif acct_policy == BOFlows:
        flow_length = store.get_user_flow_length(user['account_number'])
        if flow_length <= 0:
            return error_auth(resp,'user flow_length poor')

    if user['user_concur_number'] > 0 :
        if store.count_online(user['account_number']) >= user['user_concur_number']:
            return error_auth(resp,'user session to limit')    

    return resp