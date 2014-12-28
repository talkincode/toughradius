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


"""记账结束包处理"""
def process(req=None,user=None,runstat=None):
    if not req.get_acct_status_type() == STATUS_TYPE_STOP:
        return  
    runstat.acct_stop += 1   
    ticket = req.get_ticket()
    if not ticket.nas_addr:
        ticket.nas_addr = req.source[0]

    _datetime = datetime.datetime.now() 
    online = store.get_online(ticket.nas_addr,ticket.acct_session_id)    
    if online:
        store.del_online(ticket.nas_addr,ticket.acct_session_id)
        ticket.acct_start_time = online['acct_start_time']
        ticket.acct_stop_time= _datetime.strftime( "%Y-%m-%d %H:%M:%S")
        ticket.start_source = online['start_source']
        ticket.stop_source = STATUS_TYPE_STOP
    else:
        session_time = ticket.acct_session_time 
        stop_time = _datetime.strftime( "%Y-%m-%d %H:%M:%S")
        start_time = (_datetime - datetime.timedelta(seconds=int(session_time))).strftime( "%Y-%m-%d %H:%M:%S")
        ticket.acct_start_time = start_time
        ticket.acct_stop_time = stop_time
        ticket.start_source= STATUS_TYPE_STOP
        ticket.stop_source = STATUS_TYPE_STOP

    log.msg('%s Accounting stop request, remove online'%req.get_user_name(),level=logging.INFO)

    user = store.get_user(ticket.account_number)

    def _err_ticket(_ticket):
        _ticket.fee_receivables= 0
        _ticket.acct_fee = 0
        _ticket.is_deduct = 0
        store.add_ticket(_ticket)

    if not user:
        return _err_ticket(ticket)

    product = store.get_product(user['product_id'])
    if not product or product['product_policy'] not in (FEE_BUYOUT,FEE_TIMES):
        _err_ticket(ticket)

    if product['product_policy'] == FEE_BUYOUT:
        # buyout fee policy
        ticket.fee_receivables = 0
        ticket.acct_fee = 0
        ticket.is_deduct = 0
        store.add_ticket(ticket)

    elif product['product_policy'] == FEE_TIMES:
        # PrePay fee times policy
        sessiontime = decimal.Decimal(req.get_acctsessiontime())
        fee_price = decimal.Decimal(product['fee_price'])
        usedfee = sessiontime/decimal.Decimal(3600) * fee_price
        usedfee = int(usedfee.to_integral_value())
        balance = user['balance'] - usedfee
        if balance < 0:
            user['balance'] = 0
        else:
            user['balance'] = balance
        store.update_user_balance(user['account_number'],balance)
        ticket.fee_receivables = usedfee
        ticket.acct_fee = usedfee
        ticket.is_deduct = 1
        store.add_ticket(ticket)



        