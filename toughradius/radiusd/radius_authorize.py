#!/usr/bin/env python
# coding=utf-8
import datetime
import decimal
import traceback
from toughlib import  utils, logger, dispatch
from toughradius.manage import models
from toughradius.manage import settings
from toughradius.radiusd.radius_basic import  RadiusBasic
from toughradius.manage.events.settings import UNLOCK_ONLINE_EVENT
from toughradius.manage.events.settings import CHECK_ONLINE_EVENT

class RadiusAuth(RadiusBasic):

    def __init__(self, dbengine=None,cache=None,aes=None,request=None):
        RadiusBasic.__init__(self, dbengine,cache,aes, request)
        self.reply = {'code':0, 'msg':'success', 'attrs':{}}
        self.filters = [
            self.status_filter,
            self.bind_filter,
            self.policy_filter,
            self.limit_filter,
            self.session_filter,
            self.set_radius_attr,
            self.start_billing,
        ]

    def failure(self, msg):
        self.reply = {}
        self.reply['code'] = 1
        self.reply['msg'] = msg
        return False

    def authorize(self):
        try:
            if not self.account:
                self.failure(u'User:%s Non-existent'% self.request.account_number)
                return self.reply

            self.product = self.get_product_by_id(self.account.product_id)
            if not self.product:
                self.failure(u'User:%s authentication, the product package  <id=%s> does not exist' % (self.request.account_number, self.account.product_id))
                return self.reply

            for filter_func in self.filters:
                if not filter_func():
                    return self.reply
            return self.reply
        except Exception as err:
            self.failure(u"User:%s 认证失败, %s" % (
                self.request.account_number, utils.safeunicode(err.message) ))
            logger.exception(err,tag="radius_auth_error")
            return self.reply

    def check_free_auth(self,errmsg=''):
        """ User free Authorization check.
        If the User subscription fees to support the expiration of the license, 
        the next issue of the default session length and the maximum speed limit set
        """
        if self.product.free_auth == 0:
            return self.failure(u'User:%s has expired, %s'% (self.request.account_number,errmsg) ) 
        else:
            self.reply['input_rate'] = self.product.free_auth_uprate
            self.reply['output_rate'] = self.product.free_auth_downrate
            self.reply['attrs']['Session-Timeout'] =  int(self.get_param_value("radius_max_session_timeout",86400))
            return True

    def start_billing(self):
        if self.account.status == settings.UsrPreAuth:   
            _now = datetime.datetime.now()
            old_start = datetime.datetime.strptime(self.account.create_time,"%Y-%m-%d %H:%M:%S")
            old_end = datetime.datetime.strptime(self.account.expire_date,"%Y-%m-%d")
            day_len = old_end - old_start
            new_expire = (_now + day_len).strftime( "%Y-%m-%d")
            self.update_user_expire(new_expire)
            logger.info("user:%s Account status update as normal, start billing",trace="radius")

    def status_filter(self):
        """ 1：User Expired, status, password check"""
        self.reply['username'] = self.request.account_number

        if self.account.status == settings.UsrExpire:
            return self.check_free_auth()

        if  self.account.status in (settings.UsrPause,settings.UsrCancel):
            return self.failure(u'User:%s status:%s invalid'%(self.request.account_number,self.account.status))

        is_pwdok = self.request.radreq.is_valid_pwd(self.aes.decrypt(self.account.password))
        self.reply['attrs'].update(self.request.radreq.resp_attrs)
        if self.request.bypass == 1 and not is_pwdok:
            return self.failure(u'User:%s Password error'%self.request.account_number)

        return True

    def bind_filter(self):
        """ 2：Usermac，vlan bind check """
        macaddr = self.request['macaddr']
        if macaddr and  self.account.mac_addr:
            if self.account.bind_mac == 1 and macaddr not in self.account.mac_addr:
                return self.failure(u"user:%s macaddr binding  error"%self.request.account_number)
        elif macaddr and not self.account.mac_addr :
            self.update_user_mac(macaddr)

        vlan_id1 = int(self.request['vlanid1'])
        vlan_id2 = int(self.request['vlanid2'])
        if vlan_id1 > 0 and self.account.vlan_id1 > 0:
            if self.account.bind_vlan == 1 and vlan_id1 <> self.account.vlan_id1:
                return self.failure(u"User:%s vlanid1 Binding error"%self.request.account_number)
        elif vlan_id1 > 0 and self.account.vlan_id1 == 0:
            self.update_user_vlan_id1(vlan_id1)

        if vlan_id2 >0 and self.account.vlan_id2 > 0:
            if self.account.bind_vlan == 1 and vlan_id2 <> self.account.vlan_id2:
                return self.failure(u"User:%s vlanid2 binding error"%self.request.account_number)
        elif vlan_id2 > 0 and self.account.vlan_id2 == 0 :
            self.update_user_vlan_id2(vlan_id2)

        return True

    
    def policy_filter(self):
        """ 3：User product policy check """
        acct_policy = self.product.product_policy or PPMonth
        input_max_limit = self.product.input_max_limit
        output_max_limit = self.product.output_max_limit

        if utils.is_expire(self.account.expire_date):
            return self.check_free_auth()
                
        if acct_policy in (PPTimes,PPFlow) :
            # 预付费时长预付费流量
            if self.get_user_balance() <= 0:
                return self.check_free_auth(u'User credit is running low')    
                
        elif acct_policy == BOTimes:
            # 买断时长 
            if self.get_user_time_length() <= 0:
                return self.check_free_auth(u'User remaining time is insufficient')
                
        elif acct_policy in (BOFlows,PPMFlows): 
            # 买断流量 / 包月流量
            if self.get_user_flow_length() <= 0 and self.get_user_fixd_flow_length() <= 0:
                return self.check_free_auth(u'User remaining flows is insufficient')

        self.reply['input_rate'] = input_max_limit
        self.reply['output_rate'] = output_max_limit
        return True

    
    def limit_filter(self):
        """ 4：User Concurrency control check """
        online_count  = self.count_online()

        if self.account.user_concur_number == 0:
            return True

        if self.account.user_concur_number > 0:
            try:
                if online_count == self.account.user_concur_number:
                    auto_unlock = int(self.get_param_value("radius_auth_auto_unlock",0)) == 1
                    online = self.get_first_online(self.request.account_number)
                    if not auto_unlock:
                        dispatch.pub(CHECK_ONLINE_EVENT,online.account_number,async=True)
                        return self.failure(u'User:%s Online number over the limit'%self.request.account_number)
                    else:
                        dispatch.pub(UNLOCK_ONLINE_EVENT,
                            online.account_number,
                            online.nas_addr, 
                            online.acct_session_id,async=True)
                        return True
            except Exception as err:
                raise Exception(u'User:%s Online number over the limit, auto unlock error: %s'%(
                    self.request.account_number,
                    utils.safeunicode(traceback.format_exc()) ))
        
        if online_count > self.account.user_concur_number: 
            return self.failure(u'User:%s Online number over the limit' % self.request.account_number)
        return True

    
    def session_filter(self):
        """ 5：User Session length calculation """
        session_timeout = int(self.get_param_value("radius_max_session_timeout",86400))

        acct_interim_intelval = int(self.get_param_value("radius_acct_interim_intelval",0))
        if acct_interim_intelval > 0:
            self.reply['attrs']['Acct-Interim-Interval'] = acct_interim_intelval

        acct_policy = self.product.product_policy or PPMonth

        def _calc_session_timeout(acct_policy):
            if acct_policy in (PPMonth,BOMonth):
                # 预付费包月/买断包月
                expire_date = self.account.expire_date
                _datetime = datetime.datetime.now()
                _expire_datetime = datetime.datetime.strptime(expire_date+' 23:59:59',"%Y-%m-%d %H:%M:%S")
                _sec = (_expire_datetime - _datetime).total_seconds() 
                _sec = session_timeout if _sec < 0 else _sec
                return _sec if _sec < session_timeout else session_timeout
            elif acct_policy  == BOTimes:
                # 买断时长 
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

        session_timeout = _calc_session_timeout(acct_policy) or session_timeout

        self.reply['attrs']['Session-Timeout'] = int(session_timeout)

        return True

    def set_radius_attr(self):
        for attr in self.get_product_attrs(self.account.product_id,radius=True) or [] :
            self.reply['attrs'].setdefault(attr.attr_name, []).append(attr.attr_value)

        if self.account.ip_address:
            self.reply['attrs']['Framed-IP-Address'] = self.account.ip_address

        rate_code_attr = self.get_product_attr(self.account.product_id,"limit_rate_code")

        if rate_code_attr:
            self.reply['rate_code'] = rate_code_attr.attr_value
        
        self.reply['domain'] = self.account.domain

        return True

    












