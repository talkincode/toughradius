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

class LogoutHandler(BaseHandler):
    
    @defer.inlineCallbacks
    def get(self):
        if not self.current_user:
            self.clear_all_cookies()
            self.redirect("/login")
            return
        try:   
            cli = PortalClient(secret=self.settings.share_secret)
            rl_req = PortalV2.newReqLogout(
                self.request.remote_ip,self.settings.share_secret,self.settings.ac_addr[0])
            rl_resp = yield cli.sendto(rl_req,self.settings.ac_addr)
            if rl_resp and rl_resp.errCode > 0:
                print portalv2.AckLogoutErrs[rl_resp.errCode]
            log.msg('logout success')
        except Exception as err:
            print (u"disconnect error %s"%str(err))
            import traceback
            traceback.print_exc()
        finally:
            cli.close()

        self.clear_all_cookies()    
        self.redirect("/login",permanent=False)


        