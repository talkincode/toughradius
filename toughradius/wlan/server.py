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
from toughradius.tools.dbengine import get_engine
from sqlalchemy.sql import text as _sql
import datetime
import logging
import socket
import time
import six
import os

        
###############################################################################
# Basic Portal listen                                                      ####
###############################################################################

class PortalListen(protocol.DatagramProtocol):
    
    actions = {}
    
    def __init__(self, config,daemon=False):
        self.config = config
        self.daemon = daemon
        self.init_config()
        self.db_engine = get_engine(config)
        self.actions = {
            portalv2.NTF_LOGOUT : self.doAckNtfLogout
        }
        reactor.callLater(5,self.init_task)
        
    def init_config(self):
        self.logfile = self.config.get('portal','logfile')
        self.standalone = self.config.has_option('DEFAULT','standalone') and \
            self.config.getboolean('DEFAULT','standalone') or False
        self.secret = self.config.get('portal','secret')
        self.timezone = self.config.has_option('DEFAULT','tz') and \
            self.config.get('DEFAULT','tz') or "CST-8"
        self.debug = self.config.getboolean('DEFAULT','debug')
        self.ac1 = self.config.get('portal','ac1').split(':')
        self.ac2 = self.config.has_option('portal','ac2') and \
            self.config.get('portal','ac2').split(':') or None
        self.listen_port = self.config.getint('portal','listen')
        self.portal_port = self.config.getint('portal','port')
        self.portal_host = self.config.has_option('portal','host') \
            and self.config.get('portal','host') or  '0.0.0.0'
        self.ntf_heart = self.config.getint("portal","ntf_heart")
        try:
            os.environ["TZ"] = self.timezone
            time.tzset()
        except:pass
        
    def init_task(self):
        _task = task.LoopingCall(self.send_ntf_heart)
        _task.start(self.ntf_heart)
    
    def send_ntf_heart(self):
        host,port = self.ac1[0], int(self.ac1[1])
        req = portalv2.PortalV2.newNtfHeart(self.secret,host)
        log.msg(":: Send NTF_HEARTBEAT to %s:%s: %s"%(host,port,repr(req)),level=logging.INFO)
        try:
            self.transport.write(str(req), (host,port))
        except:
            pass
        
    def validAc(self,host):
        if host in self.ac1:
            return self.ac1
        if self.ac2 and host in self.ac2:
            return self.ac2
            
    def doAckNtfLogout(self,req,(host, port)):
        resp = portalv2.PortalV2.newMessage(
            portalv2.ACK_NTF_LOGOUT,
            req.userIp,
            req.serialNo,
            req.reqId,
            auth = req.auth,
            secret = self.secret
        )

        try:
            log.msg(":: Send portal packet to %s:%s: %s"%(host,port,repr(req)),level=logging.INFO)
            self.transport.write(str(resp), (host, port))
            # delete session on db
            sql = _sql('delete from slc_rad_online where framed_ipaddr = :user_ip')
            with self.db_engine.begin() as conn:
                conn.execute(sql,user_ip=req.userIp)
        except:
            pass
            
    
    def datagramReceived(self, datagram, (host, port)):
        ac = self.validAc(host)
        if not ac:
            return log.msg(':: Dropping packet from unknown ac host ' + host,level=logging.INFO)
        try:
            req = portalv2.PortalV2(
                secret=self.secret,
                packet=datagram,
                source=(host, port)
            )
            log.msg(":: Received portal packet from %s:%s: %s"%(host,port,repr(req)),level=logging.INFO)
            if req.type in self.actions:
                self.actions[req.type](req,(host, port))
            else:
                log.msg(':: Not support packet from ac host ' + host,level=logging.INFO)
                
        except Exception as err:
            log.err(err,':: Dropping invalid packet from %s: %s'%((host, port),str(err)))
 
    def on_exception(self,err):
        log.msg(':: Packet process error：%s' % str(err))   
        
    def run_normal(self):
        if self.debug:
            log.startLogging(sys.stdout)
        else:
            log.startLogging(DailyLogFile.fromFullPath(self.logfile))
        log.msg('portal server listen %s'%self.portal_host)  
        reactor.listenUDP(self.listen_port, self,interface=self.portal_host)
        if not self.standalone:
            reactor.run()
            
    def get_service(self):    
        from twisted.application import service, internet
        return internet.UDPServer(self.listen_port,self,interface=self.portal_host)
        
        
def run(config,is_serrvice=False):
    print 'running portal server...'
    portal = PortalListen(config,daemon=is_serrvice)
    if is_serrvice:
        return portal.get_service()
    else:
        portal.run_normal()
