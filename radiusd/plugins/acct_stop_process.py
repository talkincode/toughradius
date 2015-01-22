#!/usr/bin/env python
#coding=utf-8
from twisted.python import log
from pyrad import packet
from store import store
from settings import *
import logging
import decimal
import datetime
import utils

decimal.getcontext().prec = 11
decimal.getcontext().rounding = decimal.ROUND_UP

def process(req=None,user=None,runstat=None):
    if not req.get_acct_status_type() == STATUS_TYPE_STOP:
        return  
    runstat.acct_stop += 1   
    ticket = req.get_ticket()
    if not ticket.nas_addr:
        ticket.nas_addr = req.source[0]

    _datetime = datetime.datetime.now() 
    online = store.get_online(ticket.nas_addr,ticket.acct_session_id)    
    if not online:
        session_time = ticket.acct_session_time 
        stop_time = _datetime.strftime( "%Y-%m-%d %H:%M:%S")
        start_time = (_datetime - datetime.timedelta(seconds=int(session_time))).strftime( "%Y-%m-%d %H:%M:%S")
        ticket.acct_start_time = start_time
        ticket.acct_stop_time = stop_time
        ticket.start_source= STATUS_TYPE_STOP
        ticket.stop_source = STATUS_TYPE_STOP
        store.add_ticket(ticket)
    else:
        store.del_online(ticket.nas_addr,ticket.acct_session_id)
        ticket.acct_start_time = online['acct_start_time']
        ticket.acct_stop_time= _datetime.strftime( "%Y-%m-%d %H:%M:%S")
        ticket.start_source = online['start_source']
        ticket.stop_source = STATUS_TYPE_STOP
        store.add_ticket(ticket)

        if not user:return 

        product = store.get_product(user['product_id'])
        if product and product['product_policy'] == FEE_TIMES:
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
            ),False)


    log.msg('%s Accounting stop request, remove online'%req.get_user_name(),level=logging.INFO)



        



        