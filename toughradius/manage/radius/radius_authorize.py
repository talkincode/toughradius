#!/usr/bin/env python
# coding=utf-8
import datetime
import decimal
import traceback
from toughlib import  utils
from toughradius.manage import models
from toughradius.manage.settings import *
from toughlib.utils import timecast
from toughradius.manage.radius.radius_basic import  RadiusBasic

class RadiusAuth(RadiusBasic):

    def __init__(self, app, request):
        RadiusBasic.__init__(self, app, request)
        self.reply = {'code':0, 'msg':'success', 'attrs':{}}
        self.filters = [
            self.status_filter,
            self.bind_filter,
            self.policy_filter,
            self.limit_filter,
            self.session_filter
        ]

    def failure(self, msg):
        self.reply = {}
        self.reply['code'] = 1
        self.reply['msg'] = msg
        return False

    @timecast
    def authorize(self):
        try:
            if not self.account:
                self.failure('user %s not exists'% self.request.account_number)
                return self.reply

            self.product = self.get_product_by_id(self.account.product_id)
            if not self.product:
                self.failure('product %s not exists'% self.account.product_id)
                return self.reply

            for filter_func in self.filters:
                flag = filter_func()
                if not flag:
                    return self.reply
            return self.reply
        except Exception as err:
            self.failure("radius authorize error, %s" % utils.safeunicode(err.message))
            traceback.print_exc()
            return self.reply


    @timecast
    def status_filter(self):
        self.reply['username'] = self.request.account_number
        self.reply['bypass'] = self.get_param_value("radiusd_bypass", 0)
        if self.reply['bypass'] == 1:
            self.reply['passwd'] = self.app.aes.decrypt(self.account.password)
        if self.account.status == UsrExpire:
            self.reply['Framed-Pool'] = self.get_param_value("expire_addrpool",'')

        if  self.account.status in (UsrPause,UsrCancel):
            return self.failure('user status not ok')

        return True

    @timecast
    def bind_filter(self):
        macaddr = self.request['macaddr']
        if macaddr and  self.account.mac_addr:
            if self.account.bind_mac == 1 and macaddr not in self.account.mac_addr:
                return failure("macaddr bind not match")
        elif macaddr and not self.account.mac_addr :
            self.update_user_mac(macaddr)

        vlan_id1 = int(self.request['vlanid1'])
        vlan_id2 = int(self.request['vlanid2'])
        if vlan_id1 > 0 and self.account.vlan_id1 > 0:
            if self.account.bind_vlan == 1 and vlan_id1 <> self.account.vlan_id1:
                return self.failure("vlan_id1 bind not match")
        elif vlan_id1 > 0 and self.account.vlan_id1 == 0:
            self.update_user_vlan_id1(vlan_id1)

        if vlan_id2 >0 and self.account.vlan_id2 > 0:
            if self.account.bind_vlan == 1 and vlan_id2 <> self.account.vlan_id2:
                return self.failure("vlan_id2 bind not match")
        elif vlan_id2 > 0 and self.account.vlan_id2 == 0 :
            self.update_user_vlan_id2(vlan_id2)

        return True

    @timecast
    def policy_filter(self):
        acct_policy = self.product.product_policy or PPMonth
        if acct_policy in ( PPMonth,BOMonth):
            if utils.is_expire(self.account.expire_date):
                self.reply['attrs']['Framed-Pool'] = self.get_param_value("expire_addrpool")
                
        elif acct_policy in (PPTimes,PPFlow):
            user_balance = self.get_user_balance()
            if user_balance <= 0:
                return self.failure('Lack of balance')    
                
        elif acct_policy == BOTimes:
            time_length = self.get_user_time_length()
            if time_length <= 0:
                return self.failure('Lack of time_length')
                
        elif acct_policy == BOFlows:
            flow_length = self.get_user_flow_length()
            if flow_length <= 0:
                return self.failure('Lack of  flow_length')

        self.reply['input_limit'] = self.product.input_max_limit
        self.reply['output_limit'] = self.product.output_max_limit
        return True

    @timecast
    def limit_filter(self):
        if self.account.user_concur_number > 0:
            if self.count_online(self.account.account_number) > self.account.user_concur_number:
                return self.failure('user session to limit')
        return True

    @timecast
    def session_filter(self):
        session_timeout = int(self.get_param_value("max_session_timeout",86400))
        expire_pool = self.get_param_value("expire_addrpool",'')
        if "Framed-Pool" in self.reply['attrs']:
            if expire_pool in self.reply['attrs']['Framed-Pool']:
                expire_session_timeout = int(self.get_param_value("expire_session_timeout",0))
                if expire_session_timeout > 0:
                    session_timeout = expire_session_timeout
                else:
                    return self.failure('User has expired')

        acct_interim_intelval = int(self.get_param_value("acct_interim_intelval",0))
        if acct_interim_intelval > 0:
            self.reply['attrs']['Acct-Interim-Interval'] = acct_interim_intelval

        acct_policy = self.product.product_policy or PPMonth
        
        if acct_policy in (PPMonth,BOMonth):
            expire_date = self.account.expire_date
            _datetime = datetime.datetime.now()
            if _datetime.strftime("%Y-%m-%d") == expire_date:
                _expire_datetime = datetime.datetime.strptime(expire_date+' 23:59:59',"%Y-%m-%d %H:%M:%S")
                session_timeout = (_expire_datetime - _datetime).seconds 

        elif acct_policy  == BOTimes:
            _session_timeout = self.account.time_length
            if _session_timeout < session_timeout:
                session_timeout = _session_timeout
            
        elif acct_policy  == PPTimes:
            user_balance = self.get_user_balance()
            fee_price = decimal.Decimal(product['fee_price']) 
            _sstime = user_balance/fee_price*decimal.Decimal(3600)
            _session_timeout = int(_sstime.to_integral_value())
            if _session_timeout < session_timeout:
                session_timeout = _session_timeout

        self.reply['attrs']['Session-Timeout'] = session_timeout

        if self.account.ip_address:
            self.reply['attrs']['Framed-IP-Address'] = self.account.ip_address

        for attr in self.get_product_attrs(self.account.product_id):
            self.reply['attrs'][attr.attr_name] = attr.attr_value

        return True

    












