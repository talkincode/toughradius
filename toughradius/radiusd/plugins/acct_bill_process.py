#!/usr/bin/env python
#coding=utf-8
from twisted.python import log
from toughradius.radiusd.pyrad import packet
from toughradius.radiusd.settings import *
from toughradius.radiusd import utils
import logging
import datetime
import decimal

decimal.getcontext().prec = 32
decimal.getcontext().rounding = decimal.ROUND_UP

def send_dm(coa_clients,online):
    try:
        coa_client = coa_clients.get(online['nas_addr'])
        attrs = {
            'User-Name' : online['account_number'],
            'Acct-Session-Id' : online['acct_session_id'],
            'NAS-IP-Address' : online['nas_addr'],
            'Framed-IP-Address' : online['framed_ipaddr']
        }
        dmeq = coa_client.createDisconnectPacket(**attrs)
        coa_client.sendCoA(dmeq)
    except Exception as err:
        log.err(err,'send dm error')

def process(req=None,user=None,radiusd=None,**kwargs):
    if req.get_acct_status_type() not in (STATUS_TYPE_UPDATE,STATUS_TYPE_STOP):
        return   
    
    runstat=radiusd.runstat
    coa_clients=radiusd.coa_clients
    store = radiusd.store
        
    online = store.get_online(req.get_nas_addr(),req.get_acct_sessionid())  
    if not online:
        return
    
    online = dict(online)

    product = store.get_product(user['product_id'])
    if not product or product['product_policy'] not in (PPTimes,BOTimes,PPFlow,BOFlows):
        online['billing_times'] = req.get_acct_sessiontime()
        online['input_total'] = req.get_input_total()
        online['output_total'] = req.get_output_total()
        store.update_online(online)
        return

    def process_pptimes():
        # 预付费时长
        log.msg('%s > Prepaid long time billing '%req.get_user_name(),level=logging.INFO)
        user_balance = store.get_user_balance(user['account_number'])
        sessiontime = decimal.Decimal(req.get_acct_sessiontime())
        billing_times = decimal.Decimal(online['billing_times'])
        acct_times = sessiontime - billing_times
        fee_price = decimal.Decimal(product['fee_price'])
        usedfee = acct_times/decimal.Decimal(3600) * fee_price
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
            input_total = req.get_input_total(),
            output_total = req.get_output_total(),
            acct_times = int(acct_times.to_integral_value()),
            acct_flows = 0,
            acct_fee = usedfee,
            actual_fee = actual_fee,
            balance = balance,
            time_length = 0,
            flow_length = 0,
            is_deduct = 1,
            create_time = datetime.datetime.now().strftime( "%Y-%m-%d %H:%M:%S")
        ))
        
        if balance == 0 :
            send_dm(coa_clients,online)
        
    def process_botimes():
        #买断时长
        log.msg('%s > Buyout long time billing '%req.get_user_name(),level=logging.INFO)
        time_length = store.get_user_time_length(user['account_number'])
        sessiontime = req.get_acct_sessiontime()
        billing_times = online['billing_times']
        acct_times = sessiontime - billing_times
        user_time_length = time_length - acct_times
        if user_time_length < 0 :
            user_time_length = 0

        store.update_billing(utils.Storage(
            account_number = online['account_number'],
            nas_addr = online['nas_addr'],
            acct_session_id = online['acct_session_id'],
            acct_start_time = online['acct_start_time'],
            acct_session_time = req.get_acct_sessiontime(),
            input_total = req.get_input_total(),
            output_total = req.get_output_total(),
            acct_times = acct_times,
            acct_flows = 0,
            acct_fee = 0,
            actual_fee = 0,
            balance = 0,
            time_length = user_time_length,
            flow_length = 0,
            is_deduct = 1,
            create_time = datetime.datetime.now().strftime( "%Y-%m-%d %H:%M:%S")
        ),time_length=user_time_length)
    
        if user_time_length == 0 :
            send_dm(coa_clients,online)
        
    def process_ppflows():
        #预付费流量
        log.msg('%s > Prepaid flow billing '%req.get_user_name(),level=logging.INFO)
        user_balance = store.get_user_balance(user['account_number'])
        output_total = decimal.Decimal(req.get_output_total())
        billing_output_total = decimal.Decimal(online['output_total'])
        acct_flows = output_total - billing_output_total
        fee_price = decimal.Decimal(product['fee_price'])
        usedfee = acct_flows/decimal.Decimal(1024) * fee_price
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
            input_total = req.get_input_total(),
            output_total = req.get_output_total(),
            acct_times = 0,
            acct_flows = int(acct_flows.to_integral_value()),
            acct_fee = usedfee,
            actual_fee = actual_fee,
            balance = balance,
            time_length = 0,
            flow_length = 0,
            is_deduct = 1,
            create_time = datetime.datetime.now().strftime( "%Y-%m-%d %H:%M:%S")
        ))
        
        if balance == 0 :  
            send_dm(coa_clients,online)
        
    def process_boflows():
        #买断流量
        log.msg('%s > Buyout flow billing '%req.get_user_name(),level=logging.INFO)
        flow_length = store.get_user_flow_length(user['account_number'])
        output_total = req.get_output_total()
        billing_output_total = online['output_total']
        acct_flows = output_total - billing_output_total
        user_flow_length = flow_length - acct_flows
        if user_flow_length < 0 :
            user_flow_length = 0
            send_dm(coa_clients,online)
            
        store.update_billing(utils.Storage(
            account_number = online['account_number'],
            nas_addr = online['nas_addr'],
            acct_session_id = online['acct_session_id'],
            acct_start_time = online['acct_start_time'],
            acct_session_time = req.get_acct_sessiontime(),
            input_total = req.get_input_total(),
            output_total = req.get_output_total(),
            acct_times = 0,
            acct_flows = acct_flows,
            acct_fee = 0,
            actual_fee = 0,
            balance = 0,
            time_length = 0,
            flow_length = user_flow_length,
            is_deduct = 1,
            create_time = datetime.datetime.now().strftime( "%Y-%m-%d %H:%M:%S")
        ),flow_length=user_flow_length)
        
        if user_flow_length == 0 :
            send_dm(coa_clients,online)
    
    process_funcs = {
        PPTimes:process_pptimes,
        BOTimes:process_botimes,
        PPFlow:process_ppflows,
        BOFlows:process_boflows
    }
    
    process_funcs[product['product_policy']]()