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
      
class BasicCheck(AuthMiddleWare):
    """执行基本校验：密码校验，状态检测，并发数限制"""
    def on_auth(self):
        if not self.req.is_valid_pwd(self.user['password']):
            return self.error('user password not match')

        if not self.user['status'] == 1:
            return self.error('user status not ok')

        return self.resp

class BindCheck(AuthMiddleWare):
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

class GroupCheck(AuthMiddleWare):
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


class AcctPoicyCheck(AuthMiddleWare):
    """执行计费策略校验，用户到期检测，用户余额，时长检测"""
    def on_auth(self):

        # acct_poicy = self.user['product_poicy'] or FEE_BUYOUT
        # if acct_poicy == FEE_BUYOUT:
        #     if not utils.is_valid_date(self.user.get('auth_begin_date'),self.user.get('auth_end_date')):
        #         return self.error('user is not effective or expired')
        # elif acct_poicy == FEE_TIMES:
        #     if int(self.user.get("time_length") or 0) <= 0:
        #         return self.error('user times poor')
        # elif acct_poicy == FEE_FLOW:
        #     if int(self.user.get("flow_length") or 0) <= 0:
        #         return self.error('user credit poor')       

        return self.resp


auth_objs = [ BasicCheck,GroupCheck,BindCheck,AcctPoicyCheck ]   


########################################################################################
#
#  acct middlewares
#
########################################################################################

class AcctMiddleWare(object):
    """ super class """
    def __init__(self, req,user={}):
        self.req = req
        self.user = user

    def on_acct(self):
        pass
        

class AcctStart(AcctMiddleWare):
    """acct start"""
    def on_acct(self):
        if not self.req.get_acct_status_type() == STATUS_TYPE_START:
            return

        if not self.user:
            log.err('user %s not exists'%self.req.get_user_name())
            return

        online = dict(
            user_name = self.user['account_number'],
            nas_addr = self.req.get_nas_addr(),
            sessionid = self.req.get_acct_sessionid(),
            acct_start_time = datetime.datetime.now().strftime( "%Y-%m-%d %H:%M:%S"),
            framed_ipaddr = self.req.get_framed_ipaddr(),
            macaddr = self.req.get_mac_addr(),
            nasportid = self.req.get_nas_portid(),
            startsource = STATUS_TYPE_START)

        # rdb.hmset('online:%s:%s'%(online['nas_addr'],online['sessionid']),online)

        log.msg('%s Accounting start request , add new online'%self.user['account_number'],level=logging.INFO)
        
        
class AcctStop(AcctMiddleWare):
    """acct stop"""
    def on_acct(self):
        if not self.req.get_acct_status_type() == STATUS_TYPE_STOP:
            return        


class AcctUpdate(AcctMiddleWare):
    """acct update"""
    def on_acct(self):
        if not self.req.get_acct_status_type() == STATUS_TYPE_UPDATE:
            return        

class AcctSOn(AcctMiddleWare):
    """acct on"""
    def on_acct(self):
        if not self.req.get_acct_status_type() == STATUS_TYPE_ACCT_ON:
            return                     


class AcctSOff(AcctMiddleWare):
    """acct off"""
    def on_acct(self):
        if not self.req.get_acct_status_type() == STATUS_TYPE_ACCT_OFF:
            return           


acct_objs = [AcctStart,AcctStop,AcctUpdate,AcctSOn,AcctSOff]



