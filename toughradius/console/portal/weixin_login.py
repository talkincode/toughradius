#!/usr/bin/env python
#coding:utf-8
import sys
import socket
import os.path
import cyclone.auth
import cyclone.escape
import cyclone.web
import binascii
from hashlib import md5
from twisted.python import log
from toughradius.console.libs import utils
from toughradius.console.portal.base import BaseHandler
from toughradius.wlan.portal.portalv2 import PortalV2,hexdump
from toughradius.wlan.portal import portalv2
from toughradius.wlan.client import PortalClient
from twisted.internet import defer
from toughradius.console import models

def can_gift(offdate,interval=3600):
    if not offdate:
        return True
    cdate = datetime.datetime.strptime(offdate, '%Y-%m-%d %H:%M:%S')
    nowdate = datetime.datetime.now()
    dt = nowdate - cdate
    return  dt.total_seconds() > base

class WeixinError(Exception):pass

class MpLoginHandler(BaseHandler):

    def check_xsrf_cookie(self):
        pass

    def set_user_cookie(self,username):
        self.set_secure_cookie("portal_user", username, expires_days=1)
        self.set_secure_cookie("portal_logintime", utils.get_currtime(), expires_days=1)
    
    @defer.inlineCallbacks
    def get(self):
        account = None
        try:
            account = self.authreg()
        except:
            import traceback
            self.render("error.html",msg=u"自动登录失败:%s"%traceback.format_exc())
            return

        #判断是否欠费，间隔指定时间再给其充值
        if account.flow_length <= 0:
            _interval = self.db.query(models.SlcWlanParam).filter_by(
                param_name = "wlan_free_interval"
            ).first() 
            interval = _interval and _interval.param_value or 10
            if can_gift(account.last_offline,int(interval)*60):
                product = self.db.query(models.SlcRadProduct).get(account.product_id)
                account.flow_length = int(product.fee_times)
                self.db.commit()

        username = account.account_number
        password = utils.decrypt(account.password)

        secret = self.settings.share_secret
        ac_addr = self.settings.ac_addr
        userIp = self.request.remote_ip
        
        try:
            cli = PortalClient(secret=secret)
            # req info 
            ri_req = PortalV2.newReqInfo(userIp,secret)
            ri_resp = yield cli.sendto(ri_req,ac_addr)
            
            if ri_resp.errCode > 0:
                print portalv2.AckInfoErrs[ri_resp.errCode]
            
            # req chellenge    
            rc_req = PortalV2.newReqChallenge(userIp,secret,serialNo=ri_req.serialNo)
            rc_resp = yield cli.sendto(rc_req,ac_addr)
            
            if not rc_resp.check_resp_auth(rc_req.auth):
                self.render("login.html",msg=u"认证请求失败")
                print (u"Challenge响应验证错误，消息被丢弃")
                return

            if rc_resp.errCode > 0:
                if rc_resp.errCode == 2:
                    self.set_user_cookie(username)
                    self.redirect("/")
                    return
                self.render("login.html",msg=portalv2.AckChallengeErrs[rc_resp.errCode])
                return
                
            challenge = rc_resp.get_challenge()
            
            # req auth
            ra_req = PortalV2.newReqAuth(
                userIp,
                username,
                password,
                rc_resp.reqId,
                challenge,
                secret,
                ac_addr[0],
                serialNo=ri_req.serialNo
            )
            ra_resp = yield cli.sendto(ra_req,ac_addr)
            if not ra_resp.check_resp_auth(ra_req.auth):
                self.render("login.html",msg=u"认证请求失败")
                print (u"认证响应验证错误，消息被丢弃")
                return

            if ra_resp.errCode > 0:
                if rc_resp.errCode == 2:
                    self.set_user_cookie(username)
                    self.redirect("/")
                    return                
                self.render("login.html",msg=portalv2.AckAuthErrs[ra_resp.errCode])
                return

            # aff_ack     
            aa_req = PortalV2.newAffAckAuth(userIp,secret,ac_addr[0],ra_req.serialNo,rc_resp.reqId)
            yield cli.sendto(aa_req,ac_addr,recv=False)
            
            log.msg('auth success')
        
            self.set_user_cookie(username)
            self.redirect("/")
            
        except Exception as err:
            self.render("login.html",msg=u"auth fail,%s"%str(err))
            print (u"auth fail %s"%str(err))
            import traceback
            traceback.print_exc()
        finally:
            cli.close()


    def rand_number(self,is_pwd=False):
        r = ['0','1','2','3','4','5','6','7','8','9']
        rg = utils.random_generator
        def random_number():
            _num = ''.join([rg.choice(r) for _ in range(7)])
            if is_pwd:
                return _num
            if self.db.query(models.SlcRadAccount).filter_by(account_number=_num).count() > 0:
                return random_account()
            else:
                return _num
        return random_number()

    def authreg(self):
        mp_openid = self.get_argument("mp_openid",None)
        mp_username = self.get_argument("mp_username",mp_openid)
        mp_product_id = self.get_argument("product_id",None)
        mp_node_id = self.get_argument("node_id",None)

        if not mp_openid:
            raise WeixinError(u"微信id为空")

        member = self.db.query(models.SlcMember).filter(
            models.SlcMember.weixin_id==mp_openid
        ).first()
        product = self.db.query(models.SlcRadProduct).get(mp_product_id)
        node = self.db.query(models.SlcNode).get(mp_node_id)

        if not product or not node:
            raise WeixinError(u"资费套餐或区域节点不存在")

        account = None

        if member:
            account = self.db.query(models.SlcRadAccount).filter_by(
                account_number=member.member_name).first()
            print member.member_name
            print 'account:',account
        else:
            rand_account = self.rand_number()
            rand_pwd = self.rand_number(True)
            member = models.SlcMember()
            member.weixin_id = mp_openid
            member.node_id = mp_node_id
            member.realname = rand_account
            member.member_name = rand_account
            mpwd = rand_pwd
            member.password = md5(mpwd.encode()).hexdigest()
            member.idcard = mp_openid
            member.sex = '1'
            member.age = '0'
            member.email = ''
            member.mobile = ''
            member.address = ''
            member.create_time = utils.get_currtime()
            member.update_time = utils.get_currtime()
            member.email_active = 0
            member.mobile_active = 0
            member.active_code = utils.get_uuid()
            self.db.add(member)
            self.db.flush()
            self.db.refresh(member)

            accept_log = models.SlcRadAcceptLog()
            accept_log.accept_type = 'open'
            accept_log.accept_source = 'weixin'
            accept_log.account_number = rand_account
            accept_log.accept_time = member.create_time
            accept_log.operator_name = rand_account
            accept_log.accept_desc = u"用户微信新开户：(%s)%s"%(member.member_name,member.realname)
            self.db.add(accept_log)
            self.db.flush()
            self.db.refresh(accept_log)

            order = models.SlcMemberOrder()
            order.order_id = utils.gen_order_id()
            order.member_id = member.member_id
            order.product_id = product.id
            order.account_number = rand_account
            order.order_fee = 0
            order.actual_fee = 0
            order.pay_status = 1
            order.accept_id = accept_log.id
            order.order_source = 'console'
            order.create_time = member.create_time
            order.order_desc = u"用户微信新开账号"
            self.db.add(order)

            account = models.SlcRadAccount()
            account.account_number = rand_account
            account.ip_address = ''
            account.member_id = member.member_id
            account.product_id = order.product_id
            account.install_address = member.address
            account.mac_addr = ''
            account.password = utils.encrypt(rand_pwd)
            account.status = 1
            account.balance = 0
            account.time_length = int(product.fee_times)
            account.flow_length = int(product.fee_flows)
            account.expire_date = '3000-12-30'
            account.user_concur_number = 1
            account.bind_mac = product.bind_mac
            account.bind_vlan = product.bind_vlan
            account.vlan_id = 0
            account.vlan_id2 = 0
            account.create_time = member.create_time
            account.update_time = member.create_time
            self.db.add(account)
            self.db.commit()

        return account

        