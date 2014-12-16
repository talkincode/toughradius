#!/usr/bin/env python
#coding=utf-8
# from twisted.internet import kqreactor
# kqreactor.install()

from twisted.internet.defer import Deferred
from twisted.internet import protocol
from twisted.internet import reactor
from twisted.python import log
from pyrad import dictionary
from pyrad import host
from pyrad import packet
from store import store
from admin import UserTrace,AdminServerProtocol
from settings import auth_plugins,acct_plugins,acct_before_plugins
import middleware
import settings
import statistics
import logging
import six
import sys
import pprint
import utils
import json

###############################################################################
# Basic Defined                                                            ####
###############################################################################
        
class PacketError(Exception):pass

class RADIUS(host.Host, protocol.DatagramProtocol):
    def __init__(self, 
                dict=dictionary.Dictionary("res/dictionary"),
                trace=None,
                midware=None,
                runstat=None,
                debug=False):
        host.Host.__init__(self,dict=dict)
        self.debug = debug
        self.user_trace = trace
        self.midware = midware
        self.runstat = runstat
        self.bas_ip_pool = {bas['ip_addr']:bas for bas in store.list_bas()}

    def processPacket(self, pkt):
        pass

    def createPacket(self, **kwargs):
        raise NotImplementedError('Attempted to use a pure base class')

    def datagramReceived(self, datagram, (host, port)):
        bas = self.bas_ip_pool.get(host)
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
                if user:self.user_trace.push(user['account_number'],reply)
                return req.deferred.callback(reply)
                    
        # send accept
        reply['Reply-Message'] = 'success!'
        reply.code=packet.AccessAccept
        if user:self.user_trace.push(user['account_number'],reply)
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
            self.midware.process(plugin,req=req,user=user,runstat=self.runstat)
                

###############################################################################
# Run  Server                                                              ####
###############################################################################     
                 
if __name__ == '__main__':
    from autobahn.twisted.websocket import WebSocketServerFactory
    log.startLogging(sys.stdout, 0)
    _trace = UserTrace()
    _runstat = statistics.RunStat()
    _middleware = middleware.Middleware()
    _debug = settings.debug
    # radius server
    reactor.listenUDP(1812, RADIUSAccess(trace=_trace,midware=_middleware,runstat=_runstat,debug=_debug))
    reactor.listenUDP(1813, RADIUSAccounting(trace=_trace,midware=_middleware,runstat=_runstat,debug=_debug))
    # admin server
    factory = WebSocketServerFactory("ws://localhost:1815", debug = False)
    factory.protocol = AdminServerProtocol
    factory.protocol.user_trace = _trace
    factory.protocol.midware = _middleware
    factory.protocol.runstat = _runstat
    reactor.listenTCP(1815, factory)
    reactor.run()
