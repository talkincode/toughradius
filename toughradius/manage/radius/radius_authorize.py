#!/usr/bin/env python
# coding=utf-8
import datetime
import decimal
import traceback
from toughlib import  utils, logger, dispatch
from toughradius.manage import models
from toughradius.manage.settings import *
from toughradius.manage.events.settings import UNLOCK_ONLINE_EVENT
from toughradius.manage.radius.radius_basic import  RadiusBasic

class RadiusAuth(RadiusBasic):

    def __init__(self, dbengine=None,cache=None,aes=None,request=None):
        RadiusBasic.__init__(self, dbengine,cache,aes, request)
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
                if not filter_func():
                    return self.reply
            return self.reply
        except Exception as err:
            self.failure("radius authorize error, %s" % utils.safeunicode(err.message))
            traceback.print_exc()
            return self.reply


    def status_filter(self):
        self.reply['username'] = self.request.account_number
        self.reply['bypass'] = int(self.get_param_value("radiusd_bypass", 1))
        if self.reply['bypass'] == 1:
            self.reply['passwd'] = self.aes.decrypt(self.account.password)
        if self.account.status == UsrExpire:
            self.reply['Framed-Pool'] = self.get_param_value("expire_addrpool",'')

        if  self.account.status in (UsrPause,UsrCancel):
            return self.failure('user status not ok')

        return True

    def bind_filter(self):
        macaddr = self.request['macaddr']
        if macaddr and  self.account.mac_addr:
            if self.account.bind_mac == 1 and macaddr not in self.account.mac_addr:
                return self.failure("macaddr bind not match")
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

    def policy_filter(self):
        acct_policy = self.product.product_policy or PPMonth
        bill_type = self.get_account_attr('bill_type')
        input_max_limit = self.product.input_max_limit
        output_max_limit = self.product.output_max_limit
        if acct_policy in ( PPMonth,BOMonth):
             # 预付费包月/买断包月/自由时段
            if utils.is_expire(self.account.expire_date):
                self.reply['attrs']['Framed-Pool'] = self.get_param_value("expire_addrpool")
                return self.failure('overdue')
                
        elif acct_policy in (PPTimes,PPFlow) :
            # 预付费时长预付费流量
            if self.get_user_balance() <= 0:
                return self.failure('Lack of balance')    
                
        elif acct_policy == BOTimes:
            # 买断时长 / 自由买断时长
            if self.get_user_time_length() <= 0:
                return self.failure('Lack of time_length')
                
        elif acct_policy == BOFlows: 
            # 买断流量 / 自由买断流量
            if self.get_user_flow_length() <= 0:
                return self.failure('Lack of  flow_length')

        elif acct_policy == FreeFee:
            bill_type = int(self.get_account_attr('bill_type'))
            input_max_limit = self.get_account_attr('input_max_limit')
            output_max_limit = self.get_account_attr('output_max_limit')
            if bill_type == FreeFeeDate and utils.is_expire(self.account.expire_date):
                self.reply['attrs']['Framed-Pool'] = self.get_param_value("expire_addrpool")
            elif bill_type == FreeTimeLen and self.get_user_time_length() <= 0:
                return self.failure('Lack of time_length')
            elif bill_type == FreeFeeFlow and self.get_user_flow_length() <= 0:
                return self.failure('Lack of flow_length')

        self.reply['input_rate'] = input_max_limit
        self.reply['output_rate'] = output_max_limit
        return True

    def limit_filter(self):
        online_count  = self.count_online()
        if self.account.user_concur_number > 0:
            try:
                if online_count == self.account.user_concur_number:
                    auto_unlock = int(self.get_param_value("radius_auth_auto_unlock",0)) == 1
                    if not auto_unlock:
                        return self.failure('user session to limit')
                    else:
                        online = self.get_first_online(self.request.account_number)
                        dispatch.pub(UNLOCK_ONLINE_EVENT,
                            online.account_number,
                            online.nas_addr, 
                            online.acct_session_id,async=True)
                        return True
            except Exception, e:
                import traceback
                traceback.print_exc()
                return self.failure('user session to limit & send dm error')
        
        elif online_count > self.account.user_concur_number: 
            return self.failure('user session to limit')
        return True


    def session_filter(self):
        session_timeout = int(self.get_param_value("radius_max_session_timeout",86400))
        expire_pool = self.get_param_value("expire_addrpool",'')
        if "Framed-Pool" in self.reply['attrs']:
            if expire_pool in self.reply['attrs']['Framed-Pool']:
                expire_session_timeout = int(self.get_param_value("expire_session_timeout",0))
                if expire_session_timeout > 0:
                    session_timeout = expire_session_timeout
                else:
                    return self.failure('User has expired')

        acct_interim_intelval = int(self.get_param_value("radius_acct_interim_intelval",0))
        if acct_interim_intelval > 0:
            self.reply['attrs']['Acct-Interim-Interval'] = acct_interim_intelval

        acct_policy = self.product.product_policy or PPMonth
        def _calc_session_timeout(acct_policy,bill_type):
            if acct_policy in (PPMonth,BOMonth) or \
                    (acct_policy == FreeFee and bill_type == FreeFeeDate):
                # 预付费包月/买断包月/自由时段
                expire_date = self.account.expire_date
                _datetime = datetime.datetime.now()
                if _datetime.strftime("%Y-%m-%d") == expire_date:
                    _expire_datetime = datetime.datetime.strptime(expire_date+' 23:59:59',"%Y-%m-%d %H:%M:%S")
                    return (_expire_datetime - _datetime).seconds 
            elif acct_policy  == BOTimes or \
                     (acct_policy == FreeFee and bill_type == FreeTimeLen):
                # 买断时长 / 自由买断时长
                _session_timeout = self.account.time_length
                if _session_timeout < session_timeout:
                    return _session_timeout
            elif acct_policy  == PPTimes :
                # 预付费时长
                user_balance = self.get_user_balance()
                fee_price = decimal.Decimal(self.product['fee_price']) 
                _sstime = user_balance/fee_price*decimal.Decimal(3600)
                _session_timeout = int(_sstime.to_integral_value())
                if _session_timeout < session_timeout:
                    return _session_timeout
            else:
                return 0

        if acct_policy == FreeFee:
            bill_type = self.get_account_attr('bill_type') or 9999
            session_timeout = _calc_session_timeout(acct_policy,bill_type) or session_timeout
        else:
            session_timeout = _calc_session_timeout(acct_policy,9999) or session_timeout

        self.reply['attrs']['Session-Timeout'] = session_timeout

        if self.account.ip_address:
            self.reply['attrs']['Framed-IP-Address'] = self.account.ip_address

        for attr in (self.get_product_attrs(self.account.product_id) or []):
            self.reply['attrs'][attr.attr_name] = attr.attr_value

        return True

    












