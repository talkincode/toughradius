#!/usr/bin/env python
#coding=utf-8
# from twisted.internet import kqreactor
# kqreactor.install()
import sys,os
sys.path.insert(0,os.path.split(__file__)[0])
sys.path.insert(0,os.path.abspath(os.path.pardir))
from twisted.internet import task
from twisted.internet.defer import Deferred
from twisted.internet import protocol
from twisted.internet import reactor
from twisted.python import log
from admin import UserTrace,AdminServerProtocol
from settings import auth_plugins,acct_plugins,acct_before_plugins
from pyrad import dictionary
from pyrad import host
from pyrad import packet
from store import store
from plugins import *
import datetime
import middleware
import settings
import statistics
import logging
import six
import pprint
import utils
import json
import os
import socket

        
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
    def __init__(self, dict=None,trace=None,midware=None,runstat=None,debug=False):
        _dict = dictionary.Dictionary(dict)
        host.Host.__init__(self,dict=_dict)
        self.debug = debug
        self.user_trace = trace
        self.midware = midware
        self.runstat = runstat
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
            self.midware.process(plugin,req=req,user=user,runstat=self.runstat)
                

###############################################################################
# Run  Server                                                              ####
###############################################################################     
                 
def main():
    import argparse,json
    from twisted.python.logfile import DailyLogFile
    parser = argparse.ArgumentParser()
    parser.add_argument('-dict','--dictfile', type=str,default=None,dest='dictfile',help='dict file')
    parser.add_argument('-auth','--authport', type=int,default=0,dest='authport',help='auth port')
    parser.add_argument('-acct','--acctport', type=int,default=0,dest='acctport',help='acct port')
    parser.add_argument('-admin','--adminport', type=int,default=0,dest='adminport',help='admin port')
    parser.add_argument('-c','--conf', type=str,default=None,dest='conf',help='conf file')
    parser.add_argument('-d','--debug', nargs='?',type=bool,default=False,dest='debug',help='debug')
    args =  parser.parse_args(sys.argv[1:])

    if not args.conf or not os.path.exists(args.conf):
        print 'no config file user -c or --conf cfgfile'
        return

    _config = json.loads(open(args.conf).read())
    _mysql = _config['mysql']
    _radiusd = _config['radiusd']  

    # init args
    if args.authport:_radiusd['authport'] = args.authport
    if args.acctport:_radiusd['acctport'] = args.acctport
    if args.adminport:_radiusd['adminport'] = args.adminport
    if args.dictfile:_radiusd['dictfile'] = args.dictfile
    if args.debug:_radiusd['debug'] = bool(args.debug)   

    #init dbconfig
    settings.db_config.update(**_config)
    store.__cache_timeout__ = _radiusd['cache_timeout']

    # start logging
    log.startLogging(sys.stdout)

    _trace = UserTrace()
    _runstat = statistics.RunStat()
    _middleware = middleware.Middleware()
    _debug = _radiusd['debug'] or settings.debug

    # init coa clients
    _coa_clients = {}
    for bas in store.list_bas():
        _coa_clients[bas['ip_addr']] = CoAClient(
            bas,dictionary.Dictionary(_radiusd['dictfile']),debug=_debug)

    def start_servers():
        auth_protocol = RADIUSAccess(
            dict=_radiusd['dictfile'],trace=_trace,midware=_middleware,
            runstat=_runstat,debug=_debug
        )
        acct_protocol = RADIUSAccounting(
            dict=_radiusd['dictfile'],trace=_trace,midware=_middleware,
            runstat=_runstat,debug=_debug
        )
        reactor.listenUDP(_radiusd['authport'], auth_protocol)
        reactor.listenUDP(_radiusd['acctport'], acct_protocol)
        _task = task.LoopingCall(auth_protocol.process_delay)
        _task.start(2.7)

        from autobahn.twisted.websocket import WebSocketServerFactory
        factory = WebSocketServerFactory("ws://0.0.0.0:%s"%args.adminport, debug = False)
        factory.protocol = AdminServerProtocol
        factory.protocol.user_trace = _trace
        factory.protocol.midware = _middleware
        factory.protocol.runstat = _runstat
        factory.protocol.coa_clients = _coa_clients
        reactor.listenTCP(_radiusd['adminport'], factory)

    start_servers()
    reactor.run()


if __name__ == '__main__':
    main()