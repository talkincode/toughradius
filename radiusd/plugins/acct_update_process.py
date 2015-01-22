#!/usr/bin/env python
#coding=utf-8
from twisted.python import log
from pyrad import packet
from store import store
from settings import *
import logging
import datetime
import decimal
import utils

decimal.getcontext().prec = 11
decimal.getcontext().rounding = decimal.ROUND_UP

def process(req=None,user=None,runstat=None):
    if not req.get_acct_status_type() == STATUS_TYPE_UPDATE:
        return   

    if not user:
        return log.err("[Acct] Received an accounting update request but user[%s] not exists"%req.get_user_name())      

    runstat.acct_update += 1  
    online = store.get_online(req.get_nas_addr(),req.get_acct_sessionid())  

    if not online:         
        sessiontime = req.get_acct_sessiontime()
        updatetime = datetime.datetime.now()
        _starttime = updatetime - datetime.timedelta(seconds=sessiontime)       

        online = utils.Storage(
            account_number = user['account_number'],
            nas_addr = req.get_nas_addr(),
            acct_session_id = req.get_acct_sessionid(),
            acct_start_time = _starttime.strftime( "%Y-%m-%d %H:%M:%S"),
            framed_ipaddr = req.get_framed_ipaddr(),
            mac_addr = req.get_mac_addr(),
            nas_port_id = req.get_nas_portid(),
            billing_times = 0,
            start_source = STATUS_TYPE_UPDATE
        )
        store.add_online(online)    

    product = store.get_product(user['product_id'])
    if not product or product['product_policy'] not in (FEE_BUYOUT,FEE_MONTH,FEE_TIMES):
        return

    if product['product_policy'] == FEE_TIMES:
        # PrePay fee times policy
        user_balance = store.get_user_balance(user['account_number'])
        sessiontime = decimal.Decimal(req.get_acct_sessiontime())
        billing_times = decimal.Decimal(online['billing_times'])
        acct_length = sessiontime-billing_times
        fee_price = decimal.Decimal(product['fee_price'])
        usedfee = acct_length/decimal.Decimal(3600) * fee_price
        usedfee = actual_fee = int(usedfee.to_integral_value())
        balance = user_balance - usedfee
        if balance < 0 :  
            balance = 0
            actual_fee = user_balance
        store.update_billing(utils.Storage(
            account_number = online['account_number'],
            nas_addr = online['nas_addr'],
            acct_session_id = online['acct_session_id'],
            acct_start_time = online['acct_start_time'],
            acct_session_time = req.get_acct_sessiontime(),
            acct_length = int(acct_length.to_integral_value()),
            acct_fee = usedfee,
            actual_fee = actual_fee,
            balance = balance,
            is_deduct = 1,
            create_time = datetime.datetime.now().strftime( "%Y-%m-%d %H:%M:%S")
        ),True)





