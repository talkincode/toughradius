#!/usr/bin/env python
#coding=utf-8
from plugins import error_auth
from settings import *
from store import store
import utils

"""执行计费策略校验，用户到期检测，用户余额，时长检测"""
def process(req=None,resp=None,user=None):
    acct_policy = user['product_policy'] or FEE_MONTH
    if acct_policy in ( FEE_MONTH,FEE_BUYOUT):
        if utils.is_expire(user.get('expire_date')):
            return error_auth(resp,'user is  expired')
    elif acct_policy == FEE_TIMES:
        user_balance = store.get_user_balance(user['account_number'])
        if user_balance <= 0:
            return error_auth(resp,'user balance poor')    
    return resp