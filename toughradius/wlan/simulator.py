#!/usr/bin/env python
#coding=utf-8
import sys,os
import ConfigParser
from twisted.python.logfile import DailyLogFile
from twisted.python import log
from twisted.internet import task
from twisted.internet.defer import Deferred
from twisted.internet import protocol
from twisted.internet import reactor
from toughradius.wlan.portal import portalv2
from toughradius.wlan.portal import pktutils
import datetime
import binascii
import logging
import socket
import time
import six
import os

        
###############################################################################
#  Portal sim                                                      ####
###############################################################################

class PortalSim(protocol.DatagramProtocol):
    
    def __init__(self,config):
        self.config = config
        self.secret = config.get('portal','secret')
        
    def doAckChellenge(self,req):
        resp = portalv2.PortalV2.newMessage(
            portalv2.ACK_CHALLENGE,
            req.userIp,
            req.serialNo,
            portalv2.CurrentSN(),
            auth = req.auth,
            secret = self.secret
        )
        resp.attrNum = 2
        resp.attrs = [
            (0x03,'\x01\x02\x03\x04\x05\x06\x07\x08\x01\x02\x03\x04\x05\x06\x07\x08'),
            (0x0a,'\x7f\x00\x00\x01')
        ]
        resp.auth_packet()
        return resp
        
    def doAckAuth(self,req):
        resp = portalv2.PortalV2.newMessage(
            portalv2.ACK_AUTH,
            req.userIp,
            req.serialNo,
            req.reqId,
            auth = req.auth,
            secret = self.secret
        )
        resp.attrNum = 2
        resp.attrs = [
            (0x0a,'\x7f\x00\x00\x01'),
            (0x0b,'\x01\x02\x03\x04\x05\x06')
        ]
        resp.auth_packet()
        return resp
        
    def doAckLogout(self,req):
        resp = portalv2.PortalV2.newMessage(
            portalv2.ACK_LOGOUT,req.userIp,
            req.serialNo,0,
            auth = req.auth,
            secret = self.secret
        )
        resp.attrNum = 1
        resp.attrs = [
            (0x0a,'\x7f\x00\x00\x01')
        ]
        resp.auth_packet()
        
        resp2 = portalv2.PortalV2.newMessage(
            portalv2.NTF_LOGOUT,req.userIp,
            0,req.reqId,
            auth = req.auth,
            secret = self.secret
        )
        resp2.auth_packet()
        reactor.callLater(1,self.sendtoPortald,resp2)
        return resp
        
    def doAckInfo(self,req):
        resp = portalv2.PortalV2.newMessage(
            portalv2.ACK_INFO,
            req.userIp,
            req.serialNo,
            0,
            auth = req.auth,
            secret = self.secret
        )
        resp.attrNum = 1
        resp.attrs = [
            (0x0a,'\x7f\x00\x00\x01')
        ]
        resp.auth_packet()
        return resp
        
    def sendtoPortald(self,msg):
        portal_addr = (
            self.config.get("portal",'host'),
            self.config.getint('portal','listen')
        )
        log.msg(":: Send Message to Portal Listen %s: %s"%(portal_addr,repr(msg)))
        self.transport.write(str(msg), portal_addr)
    
    def datagramReceived(self, datagram, (host, port)):
        try:
            print 
            log.msg("*"*120)
            print
            
            print portalv2.hexdump(datagram,len(datagram))
            _packet = portalv2.PortalV2(secret=self.secret,packet=datagram,source=(host, port))
            log.msg(":: Received portal packet from %s:%s: %s"%(host,port,repr(_packet)),level=logging.INFO)
            resp = None
            if _packet.type == portalv2.REQ_CHALLENGE:
                resp = self.doAckChellenge(_packet)
            elif _packet.type == portalv2.REQ_AUTH:
                resp = self.doAckAuth(_packet)
            elif _packet.type == portalv2.REQ_LOGOUT:
                resp = self.doAckLogout(_packet)
            elif _packet.type == portalv2.REQ_INFO:
                resp = self.doAckInfo(_packet)
            elif _packet.type == portalv2.AFF_ACK_AUTH:
                pass
            elif _packet.type == portalv2.ACK_NTF_LOGOUT:
                pass
            elif _packet.type == portalv2.NTF_HEARTBEAT:
                pass
            else:
                log.msg("not support packet type: %s"%_packet.type)
            print
            if resp:
                print portalv2.hexdump(str(resp),len(resp))
                log.msg(":: Send response to %s:%s: %s"%(host,port,repr(resp)))
                self.transport.write(str(resp), (host, port))
        except Exception as err:
            log.err(err,':: Dropping invalid packet from %s: %s'%((host, port),str(err)))
            import traceback
            traceback.print_exc()

 
    def on_exception(self,err):
        log.msg(':: Packet process error：%s' % str(err))   
        
    def run_normal(self):
        log.startLogging(sys.stdout)
        reactor.listenUDP(2000, self)
        reactor.run()
            

def run(config):
    print 'running portal sim server...'
    portal = PortalSim(config)
    portal.run_normal()