#!/usr/bin/env python
#coding=utf-8
from twisted.python import log
from pyrad import packet
from store import store
from settings import *
import logging
import datetime
import utils

########################################################################################
#
#  parse middlewares
#
########################################################################################
class ParseMiddleWare(object):
    """ super class """
    def __init__(self,req):
        self.req = req

    def on_parse(self):
        pass

#  mac parse 
class MacParseFilter(ParseMiddleWare):
    """解析MAC地址"""
    def __init__(self,req):
        ParseMiddleWare.__init__(self,req)
        self.parses = {
            '9' : self.parse_cisco,
            '2352' : self.parse_radback,
            '3902' : self.parse_zte,
            '25506' : self.parse_h3c
        }
        
    def parse_cisco(self):
        for attr in self.req:
            if attr not in 'Cisco-AVPair':
                continue
            attr_val = self.req[attr]
            if attr_val.startswith('client-mac-address'):
                mac_addr = attr_val[len("client-mac-address="):]
                mac_addr = mac_addr.replace('.','')
                _mac = (mac_addr[0:2],mac_addr[2:4],mac_addr[4:6],mac_addr[6:8],mac_addr[8:10],mac_addr[10:])
                self.req.client_macaddr =  ':'.join(_mac)
    
    def parse_radback(self):
        mac_addr = self.req.get('Mac-Addr')
        if mac_addr:self.req.client_macaddr = mac_addr.replace('-',':')
        
    def parse_zte(self):
        mac_addr = self.req.get('Calling-Station-Id')
        if mac_addr:
            mac_addr = mac_addr[12:] 
            _mac = (mac_addr[0:2],mac_addr[2:4],mac_addr[4:6],mac_addr[6:8],mac_addr[8:10],mac_addr[10:])
            self.req.client_macaddr =  ':'.join(_mac)
            
    def parse_h3c(self):
        mac_addr = self.req.get('H3C-Ip-Host-Addr')
        if mac_addr and len(mac_addr) > 17:
            self.req.client_macaddr = mac_addr[:-17]
        else:
            self.req.client_macaddr = mac_addr
        
    def on_parse(self):
        try:
            if self.req.vendor_id in self.parses:
                self.parses[self.req.vendor_id]()
        except Exception as err:
            log.err("parse vlan error %s"%str(err))
     
#  vlan parse          
class VlanParseFilter(ParseMiddleWare):
    """解析VLAN"""
    def __init__(self,req):
        ParseMiddleWare.__init__(self,req)
        self.parses = {
            '0' : self.parse_std,
            '9' : self.parse_cisco,
            '3041' : self.parse_cisco,
            '2352' : self.parse_radback,
            '2011' : self.parse_std,
            '25506' : self.parse_std,
            '3902' : self.parse_zte,
            '14988' : self.parse_ros
        }
        
    def parse_cisco(self):
        '''phy_slot/phy_subslot/phy_port:XPI.XCI'''
        nasportid = self.req.get('NAS-Port-Id')
        if not nasportid:return
        nasportid = nasportid.lower()
        def parse_vlanid():
            ind = nasportid.find(':')
            if ind == -1:return
            ind2 = nasportid.find('.',ind)
            if ind2 == -1:
                self.req.vlanid = int(nasportid[ind+1])
            else:
                self.req.vlanid = int(nasportid[ind+1:ind2])
        def parse_vlanid2():
            ind = nasportid.find('.')
            if ind == -1:return
            ind2 = nasportid.find(' ',ind)
            if ind2 == -1:
                self.req.vlanid2 = int(nasportid[ind+1])
            else:
                self.req.vlanid2 = int(nasportid[ind+1:ind2])
                
        parse_vlanid()
        parse_vlanid2()
        
    def parse_radback(self):
        self.parse_ros()
        
    def parse_zte(self):
        self.parse_ros()
        
    def parse_std(self):
        ''''''
        nasportid = self.req.get('NAS-Port-Id')
        if not nasportid:return
        nasportid = nasportid.lower()
        def parse_vlanid():
            ind = nasportid.find('vlanid=')
            if ind == -1:return
            ind2 = nasportid.find(';',ind)
            if ind2 == -1:
                self.req.vlanid = int(nasportid[ind+7])
            else:
                self.req.vlanid = int(nasportid[ind+7:ind2])
                
        def parse_vlanid2():
            ind = nasportid.find('vlanid2=')
            if ind == -1:return
            ind2 = nasportid.find(';',ind)
            if ind2 == -1:
                self.req.vlanid2 = int(nasportid[ind+8])
            else:
                self.req.vlanid2 = int(nasportid[ind+8:ind2])
                
        parse_vlanid()
        parse_vlanid2() 
        
    def parse_ros(self):
        ''''''
        nasportid = self.req.get('NAS-Port-Id')
        if not nasportid:return
        nasportid = nasportid.lower()
        def parse_vlanid():
            ind = nasportid.find(':')
            if ind == -1:return        
            ind2 = nasportid.find(' ',ind)
            if ind2 == -1:return
            self.req.vlanid = int(nasportid[ind+1:ind2])
            
        def parse_vlanid2():
            ind = nasportid.find(':')
            if ind == -1:return
            ind2 = nasportid.find('.',ind)
            if ind2 == -1:return
            self.req.vlanid2 = int(nasportid[ind+1:ind2])
            
        parse_vlanid()
        parse_vlanid2()               
        
    def on_parse(self):
        try:
            if self.req.vendor_id in self.parses:
                self.parses[self.req.vendor_id]()
        except Exception as err:
            log.err("parse vlan error %s"%str(err))
        

auth_parse_objs = [MacParseFilter,VlanParseFilter]
acct_parse_objs = [MacParseFilter]

########################################################################################
#
#  auth middlewares
#
########################################################################################

class AuthMiddleWare(object):
    """ super class """
    def __init__(self,req,resp,user={}):
        self.req = req
        self.resp = resp
        self.user = user

    def on_auth(self):
        return self.resp

    def error(self,errmsg):
        self.resp.code = packet.AccessReject
        self.resp['Reply-Message'] = errmsg
        return self.resp
        

class BasicFilter(AuthMiddleWare):
    """执行基本校验：密码校验，状态检测，并发数限制"""
    def on_auth(self):
        if not self.user:
            return self.error('user %s not exists'%self.req.get_user_name())

        if not self.req.is_valid_pwd(self.user['password']):
            return self.error('user password not match')

        if not self.user['status'] == 1:
            return self.error('user status not ok')

        return self.resp

class DomainFilter(AuthMiddleWare):
    """执行基本校验：密码校验，状态检测，并发数限制"""
    def on_auth(self):
        domain = self.req.get_domain()
        user_domain  = self.user.get("domain_name")
        if domain and user_domain:
            if domain not in user_domain:
                return self.error('user domain %s not match'%domain  )
        return self.resp       
              

class BindFilter(AuthMiddleWare):
    """执行绑定校验，检查MAC地址与VLANID"""
    def on_auth(self):

        macaddr = self.req.get_mac_addr()
        if macaddr and  self.user['mac_addr']:
            if self.user['bind_mac'] == 1 and macaddr not in self.user['mac_addr']:
                return self.error("macaddr bind not match")
        elif macaddr and not self.user['mac_addr'] :
            store.update_user_mac(self.user['account_number'], macaddr)

        vlan_id,vlan_id2 = self.req.get_vlanids()
        if vlan_id and self.user['vlan_id']:
            if self.user['bind_vlan'] == 1 and vlan_id != self.user['vlan_id']:
                return self.error("vlan_id bind not match")
        elif vlan_id and not self.user['vlan_id']:
            self.user['vlan_id'] = vlan_id
            store.update_user_vlan_id(self.user['account_number'],vlan_id)

        if vlan_id2 and self.user['vlan_id2']:
            if self.user['bind_vlan'] == 1 and vlan_id2 != self.user['vlan_id2']:
                return self.error("vlan_id2 bind not match")
        elif vlan_id2 and not self.user['vlan_id2']:
            self.user['vlan_id2'] = vlan_id2
            store.update_user_vlan_id2(self.user['account_number'],vlan_id2)

        return self.resp

class GroupFilter(AuthMiddleWare):
    """执行用户组策略校验，检查MAC与VLANID绑定，并发数限制 """
    def on_auth(self):
        group = store.get_group(self.user['group_id'])

        if not group:
            return self.resp

        if group['bind_mac']:
            if self.user['mac_addr'] and self.get_mac_addr() not in self.user['mac_addr']:
                return self.error("macaddr not match")
            if not self.user['mac_addr']:
                self.user['mac_addr'] = self.get_mac_addr()
                store.update_user_mac(self.user['account_number'],self.get_mac_addr())

        if group['bind_vlan']:
            vlan_id,vlan_id2 = self.req.get_vlanids()
            #update user vlan_bind
            if vlan_id and self.user['vlan_id']:
                if vlan_id != self.user['vlan_id']:
                    return self.error("vlan_id bind not match")
            elif vlan_id and not self.user['vlan_id']:
                self.user['vlan_id'] = vlan_id
                store.update_user_vlan_id(self.user['account_number'],vlan_id)

            if vlan_id2 and self.user['vlan_id2']:
                if vlan_id2 != self.user['vlan_id2']:
                    return self.error("vlan_id2 bind not match")
            elif vlan_id2 and not self.user['vlan_id2']:
                self.user['vlan_id2'] = vlan_id2
                store.update_user_vlan_id2(self.user['account_number'],vlan_id2)

        return self.resp


class AcctPoicyFilter(AuthMiddleWare):
    """执行计费策略校验，用户到期检测，用户余额，时长检测"""
    def on_auth(self):
        acct_policy = self.user['product_policy'] or FEE_BUYOUT
        if acct_policy == FEE_BUYOUT:
            if utils.is_expire(self.user.get('expire_date')):
                return self.error('user is  expired')
        elif acct_policy == FEE_TIMES:
            if int(self.user.get("balance",0)) <= 0:
                return self.error('user balance poor')    
        return self.resp


auth_objs = [ BasicFilter,DomainFilter,GroupFilter,BindFilter,AcctPoicyFilter ]   


########################################################################################
#
#  acct middlewares
#
########################################################################################

class AccountingMiddleWare(object):
    """ super class """
    def __init__(self, req,user={}):
        self.req = req
        self.user = user

    def on_acct(self):
        pass
        

class AccountingStart(AccountingMiddleWare):
    """记账开始包处理"""
    def on_acct(self):
        if not self.req.get_acct_status_type() == STATUS_TYPE_START:
            return
            
        if not self.user:
            return log.err('user %s not exists'%self.req.get_user_name())
            
        online = utils.Storage(
            account_number = self.user['account_number'],
            nas_addr = self.req.get_nas_addr(),
            sessionid = self.req.get_acct_sessionid(),
            acct_start_time = datetime.datetime.now().strftime( "%Y-%m-%d %H:%M:%S"),
            framed_ipaddr = self.req.get_framed_ipaddr(),
            macaddr = self.req.get_mac_addr(),
            nasportid = self.req.get_nas_portid(),
            start_source = STATUS_TYPE_START
        )

        store.add_online(online)

        log.msg('%s Accounting start request, add new online'%self.req.get_user_name(),level=logging.INFO)
        
        
class AccountingStop(AccountingMiddleWare):
    """记账结束包处理"""
    def on_acct(self):
        if not self.req.get_acct_status_type() == STATUS_TYPE_STOP:
            return  
        
        ticket = self.req.get_ticket()
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

        log.msg('%s Accounting stop request, remove online'%self.req.get_user_name(),level=logging.INFO)

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
            sessiontime = round(req.get_acctsessiontime()/60,0)
            usedfee = round(sessiontime/60*product['fee_price'],0)
            remaind = round(sessiontime%60,0)
            if remaind > 0 :
                usedfee = usedfee + round(remaind*product.fee_price/60,0);
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
            
class AccountingUpdate(AccountingMiddleWare):
    """记账更新包处理"""
    def on_acct(self):
        if not self.req.get_acct_status_type() == STATUS_TYPE_UPDATE:
            return   

        online = store.get_online(self.req.get_nas_addr(),self.req.get_acct_sessionid())  

        if not online:
            user = store.get_user(self.req.get_user_name())
            if not user:
                return log.err("[Acct] Received an accounting update request but user[%s] not exists"%self.req.get_user_name())
                            
            sessiontime = self.req.get_acct_sessiontime()
            updatetime = datetime.datetime.now()
            _starttime = updatetime - datetime.timedelta(seconds=sessiontime)       

            online = utils.Storage(
                account_number = self.user['account_number'],
                nas_addr = self.req.get_nas_addr(),
                sessionid = self.req.get_acct_sessionid(),
                acct_start_time = _starttime.strftime( "%Y-%m-%d %H:%M:%S"),
                framed_ipaddr = self.req.get_framed_ipaddr(),
                macaddr = self.req.get_mac_addr(),
                nasportid = self.req.get_nas_portid(),
                start_source = STATUS_TYPE_UPDATE
            )
            store.add_online(online)       


class AccountingClose(AccountingMiddleWare):
    """记账启动关闭处理"""
    def on_acct(self):
        if  self.req.get_acct_status_type() in (STATUS_TYPE_ACCT_ON,STATUS_TYPE_ACCT_OFF):
            onlines = store.del_nas_onlines(self.req.get_nas_addr()) 

        if self.req.get_acct_status_type() == STATUS_TYPE_ACCT_ON:
            log.msg('bas accounting on success',level=logging.INFO)
        else:
            log.msg('bas accounting off success',level=logging.INFO)
              

acct_objs = [AccountingStart,AccountingStop,AccountingUpdate,AccountingClose]



