#!/usr/bin/env python
#coding:utf-8
import sys
import socket
import os.path
import cyclone.auth
import cyclone.escape
import cyclone.web
import binascii
from twisted.python import log
from toughradius.console.libs import utils
from toughradius.console.portal.base import BaseHandler
from toughradius.wlan.portal.portalv2 import PortalV2,hexdump
from toughradius.wlan.portal import portalv2
from toughradius.wlan.client import PortalClient
from twisted.internet import defer



class LoginHandler(BaseHandler):
    
    def get(self):
        self.render(self.get_login_template())
  
    @defer.inlineCallbacks
    def post(self):
        def set_user_cookie():
            self.set_secure_cookie("portal_user", username, expires_days=1)
            self.set_secure_cookie("portal_logintime", utils.get_currtime(), expires_days=1)
        username = self.get_argument("username",None)
        password = self.get_argument("password",None)
        wlanuserip = self.get_argument("wlanuserip", None)
        if not username or not password:
            self.render(self.get_login_template(),msg=u"请输入用户名和密码")
            return
            
        secret = self.settings.share_secret
        ac_addr = self.settings.ac_addr
        userIp = wlanuserip or self.request.remote_ip
        
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
                print (u"Challenge resp error,msg droped")
                return

            if rc_resp.errCode > 0:
                if rc_resp.errCode == 2:
                    set_user_cookie()
                    self.redirect("/")
                    return
                self.render(self.get_login_template(),msg=portalv2.AckChallengeErrs[rc_resp.errCode])
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
                print (u"Challenge resp error,msg droped")
                return

            if ra_resp.errCode > 0:
                if rc_resp.errCode == 2:
                    set_user_cookie()
                    self.redirect("/")
                    return                
                self.render(self.get_login_template(),msg=portalv2.AckAuthErrs[ra_resp.errCode])
                return

            # aff_ack     
            aa_req = PortalV2.newAffAckAuth(userIp,secret,ac_addr[0],ra_req.serialNo,rc_resp.reqId)
            yield cli.sendto(aa_req,ac_addr,recv=False)
            
            log.msg('auth success')
        
            set_user_cookie()
            self.redirect("/")
            
        except Exception as err:
            self.render(self.get_login_template(),msg=u"认证请求失败,%s"%str(err))
            import traceback
            traceback.print_exc()
        finally:
            cli.close()

        