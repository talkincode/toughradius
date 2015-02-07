#!/usr/bin/env python
#coding=utf-8
from plugins import error_auth
from settings import *
from store import store
import utils


def process(req=None,resp=None,user=None):
    """执行计费策略校验，用户到期检测，用户余额，时长检测"""
    acct_policy = user['product_policy'] or FEE_MONTH
    if acct_policy in ( FEE_MONTH,FEE_BUYOUT):
        if utils.is_expire(user.get('expire_date')):
            resp['Framed-Pool'] = store.get_param("9_expire_addrpool")
    elif acct_policy == FEE_TIMES:
        user_balance = store.get_user_balance(user['account_number'])
        if user_balance <= 0:
            return error_auth(resp,'user balance poor')    

    if user['user_concur_number'] > 0 :
        if store.count_online(user['account_number']) >= user['user_concur_number']:
            return error_auth(resp,'user session to limit')    

    return resp