#!/usr/bin/env python
#coding=utf-8
from autobahn.twisted import choosereactor
choosereactor.install_optimal_reactor(True)
import sys,os
import ConfigParser
from twisted.python.logfile import DailyLogFile
from twisted.python import log
from twisted.internet import task
from twisted.internet.defer import Deferred
from twisted.internet import protocol
from twisted.internet import reactor
from autobahn.twisted.websocket import WebSocketServerProtocol
from autobahn.twisted.websocket import WebSocketServerFactory
from toughradius.radiusd.settings import *
from toughradius.radiusd.pyrad import dictionary
from toughradius.radiusd.pyrad import host
from toughradius.radiusd.pyrad import packet
from toughradius.radiusd.store import store
from toughradius.radiusd import middleware
from toughradius.radiusd import settings
import datetime
import logging
import pprint
import socket
import utils
import time
import json
import six
import os

__verson__ = '0.7'
        
class PacketError(Exception):pass

###############################################################################
# Coa Client                                                             ####
###############################################################################

class CoAClient(protocol.DatagramProtocol):
    
    def __init__(self, bas,dict=None,debug=False):
        assert bas 
        self.bas = bas
        self.dict = dict
        self.secret = six.b(str(self.bas['bas_secret']))
        self.addr = self.bas['ip_addr']
        self.port = self.bas['coa_port']
        self.debug=debug
        reactor.listenUDP(0, self)
        
    def processPacket(self, pkt):
        pass

    def createPacket(self, **kwargs):
        return utils.CoAPacket2(dict=self.dict,secret=self.secret,**kwargs)

    def createDisconnectPacket(self, **kwargs):
        return utils.CoAPacket2(
            code=packet.DisconnectRequest,
            dict=self.dict,
            secret=self.secret,
            **kwargs)    
    
    def sendCoA(self,pkt):
        log.msg("send radius Coa Request: %s"%(pkt),level=logging.INFO)
        try:
            self.transport.write(pkt.RequestPacket(),(self.addr, self.port))
        except packet.PacketError as err:
            log.err(err,'::send radius Coa Request error %s: %s'%((host, port),str(err)))

    def datagramReceived(self, datagram, (host, port)):
        if host != self.addr:
            return log.msg('Dropping Radius Coa Packet from unknown host ' + host,level=logging.INFO)
        try:
            coaResponse = self.createPacket(packet=datagram)
            coaResponse.source = (host, port)
            log.msg("::Received Radius Coa Response: %s"%(str(coaResponse)),level=logging.INFO)
            if self.debug:
                log.msg(coaResponse.format_str(),level=logging.DEBUG)    
            self.processPacket(coaResponse)
        except packet.PacketError as err:
            log.err(err,'::Dropping invalid CoA Response packet from %s: %s'%((host, port),str(err)))

    def on_exception(self,err):
        log.msg('CoA Packet process error：%s' % str(err))   

###############################################################################
# Basic RADIUS                                                            ####
###############################################################################

class RADIUS(host.Host, protocol.DatagramProtocol):
    def __init__(self, dict=None,trace=None,midware=None,runstat=None,coa_clients=None,debug=False):
        _dict = dictionary.Dictionary(dict)
        host.Host.__init__(self,dict=_dict)
        self.debug = debug
        self.user_trace = trace
        self.midware = midware
        self.runstat = runstat
        self.coa_clients = coa_clients
        self.auth_delay = utils.AuthDelay(int(store.get_param("reject_delay") or 0))

    def processPacket(self, pkt):
        pass

    def createPacket(self, **kwargs):
        raise NotImplementedError('Attempted to use a pure base class')

    def datagramReceived(self, datagram, (host, port)):
        bas = store.get_bas(host)
        if not bas:
            return log.msg('Dropping packet from unknown host ' + host,level=logging.DEBUG)
        secret,vendor_id = bas['bas_secret'],bas['vendor_id']
        try:
            _packet = self.createPacket(packet=datagram,dict=self.dict,secret=six.b(str(secret)),vendor_id=vendor_id)
            _packet.deferred.addCallbacks(self.reply,self.on_exception)
            _packet.source = (host, port)
            log.msg("::Received radius request: %s"%(str(_packet)),level=logging.INFO)
            if self.debug:
                log.msg(_packet.format_str(),level=logging.DEBUG)    
            self.processPacket(_packet)
        except packet.PacketError as err:
            log.err(err,'::Dropping invalid packet from %s: %s'%((host, port),str(err)))

    def reply(self,reply):
        log.msg("send radius response: %s"%(reply),level=logging.INFO)
        if self.debug:
            log.msg(reply.format_str(),level=logging.DEBUG)
        self.transport.write(reply.ReplyPacket(), reply.source)  
        if reply.code == packet.AccessReject:
            self.runstat.auth_reject += 1
        elif reply.code == packet.AccessAccept:
            self.runstat.auth_accept += 1
 
    def on_exception(self,err):
        log.msg('Packet process error：%s' % str(err))   

    def process_delay(self):
        while self.auth_delay.delay_len() > 0:
            try:
                reject = self.auth_delay.get_delay_reject(0)
                if (datetime.datetime.now() - reject.created).seconds < self.auth_delay.reject_delay:
                    return
                else:
                    self.reply(self.auth_delay.pop_delay_reject())
            except:
                log.err("process_delay error")

###############################################################################
# Auth Server                                                              ####
###############################################################################
class RADIUSAccess(RADIUS):

    def createPacket(self, **kwargs):
        vendor_id = 0
        if 'vendor_id' in kwargs:
            vendor_id = kwargs.pop('vendor_id')
        pkt = utils.AuthPacket2(**kwargs)
        pkt.vendor_id = vendor_id
        return pkt

    def processPacket(self, req):
        self.runstat.auth_all += 1
        if req.code != packet.AccessRequest:
            self.runstat.auth_drop += 1
            raise PacketError('non-AccessRequest packet on authentication socket')
        
        reply = req.CreateReply()
        reply.source = req.source
        user = store.get_user(req.get_user_name())
        if user:self.user_trace.push(user['account_number'],req)
        # middleware execute
        for plugin in auth_plugins:
            self.midware.process(plugin,req=req,resp=reply,user=user)
            if reply.code == packet.AccessReject:
                self.auth_delay.add_roster(req.get_mac_addr())
                if user:self.user_trace.push(user['account_number'],reply)
                if self.auth_delay.over_reject(req.get_mac_addr()):
                    return self.auth_delay.add_delay_reject(reply)
                else:
                    return req.deferred.callback(reply)
                    
        # send accept
        reply['Reply-Message'] = 'success!'
        reply.code=packet.AccessAccept
        if user:self.user_trace.push(user['account_number'],reply)
        self.auth_delay.del_roster(req.get_mac_addr())
        req.deferred.callback(reply)
        
        
###############################################################################
# Acct Server                                                              ####
############################################################################### 

class RADIUSAccounting(RADIUS):

    def createPacket(self, **kwargs):
        vendor_id = 0
        if 'vendor_id' in kwargs:
            vendor_id = kwargs.pop('vendor_id')
        pkt = utils.AcctPacket2(**kwargs)
        pkt.vendor_id = vendor_id
        return pkt

    def processPacket(self, req):
        self.runstat.acct_all += 1
        if req.code != packet.AccountingRequest:
            self.runstat.acct_drop += 1
            raise PacketError('non-AccountingRequest packet on authentication socket')

        for plugin in acct_before_plugins:
            self.midware.process(plugin,req=req)
                 
        user = store.get_user(req.get_user_name())
        if user:self.user_trace.push(user['account_number'],req)        
          
        reply = req.CreateReply()
        reply.source = req.source
        if user:self.user_trace.push(user['account_number'],reply)   
        req.deferred.callback(reply)
        # middleware execute
        for plugin in acct_plugins:
            self.midware.process(plugin,req=req,user=user,
            runstat=self.runstat,coa_clients=self.coa_clients
        )
        
 
 ###############################################################################
 # admin  Server                                                            ####
 ###############################################################################
 
class AdminServerProtocol(WebSocketServerProtocol):

    user_trace = None
    midware = None
    runstat = None
    coa_clients = {}
    auth_server = None
    acct_server = None

    def onConnect(self, request):
        log.msg("Client connecting: {0}".format(request.peer))

    def onOpen(self):
        log.msg("WebSocket connection open.")

    def onMessage(self, payload, isBinary):
        req_msg = json.loads(payload)
        log.msg("websocket admin request: %s"%str(req_msg))
        plugin = req_msg.get("process")
        self.midware.process(plugin,req=req_msg,admin=self)

    def onClose(self, wasClean, code, reason):
        log.msg("WebSocket connection closed: {0}".format(reason))

###############################################################################
# Run  Server                                                              ####
###############################################################################     
                 
def run(config):
    log.startLogging(sys.stdout)
    secret = config.get('DEFAULT','secret')
    dictfile = config.get('radiusd','dictfile')
    tz = config.get('DEFAULT','tz')
    is_debug = config.getboolean('DEFAULT','debug')
    authport = config.getint('radiusd','authport')
    acctport = config.getint('radiusd','acctport')
    adminport = config.getint('radiusd','adminport')
    
    #init dbstore
    store.setup(config)
    # update aescipher,timezone
    utils.aescipher.setup(secret)
    utils.update_tz(tz)
    # rundata
    _trace = utils.UserTrace()
    _runstat = utils.RunStat()
    _middleware = middleware.Middleware()
    # init coa clients
    coa_pool = {}
    for bas in store.list_bas():
        coa_pool[bas['ip_addr']] = CoAClient(bas,
            dictionary.Dictionary(dictfile),
            debug=is_debug
        )

    auth_protocol = RADIUSAccess(
        dict=dictfile,trace=_trace,midware=_middleware,
        runstat=_runstat,coa_clients=coa_pool,debug=is_debug
    )
    
    acct_protocol = RADIUSAccounting(
        dict=dictfile,trace=_trace,midware=_middleware,
        runstat=_runstat,coa_clients=coa_pool,debug=is_debug
    )
    
    reactor.listenUDP(authport, auth_protocol)
    reactor.listenUDP(acctport, acct_protocol)
    _task = task.LoopingCall(auth_protocol.process_delay)
    _task.start(2.7)

    factory = WebSocketServerFactory("ws://0.0.0.0:%s"%adminport, debug = False)
    factory.protocol = AdminServerProtocol
    factory.protocol.user_trace = _trace
    factory.protocol.midware = _middleware
    factory.protocol.runstat = _runstat
    factory.protocol.coa_clients = coa_pool
    factory.protocol.auth_server = auth_protocol
    factory.protocol.acct_server = acct_protocol
    reactor.listenTCP(adminport, factory)

    reactor.run()

