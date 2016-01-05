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
from autobahn.twisted.websocket import WebSocketServerProtocol
from autobahn.twisted.websocket import WebSocketServerFactory
from toughradius.radiusd.settings import *
from toughradius.radiusd.pyrad import dictionary
from toughradius.radiusd.pyrad import host
from toughradius.radiusd.pyrad import packet
from toughradius.radiusd.store import Store
from toughradius.radiusd import middleware
from toughradius.radiusd import settings
from toughradius.radiusd import utils
from toughradius.tools.dbengine import get_engine
import datetime
import logging
import pprint
import socket
import time
import json
import six
import os
import ikuai

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
        self.vendor_id = int(self.bas['vendor_id'])
        self.debug=debug
        self.uport = reactor.listenUDP(0, self)

    def close(self):
        self.transport = None
        try:
            self.uport.stopListening()
        except:
            pass
        
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
            print self.vendor_id
            if self.vendor_id == ikuai.VENDOR_ID:
                pkg = ikuai.create_dm_pkg(self.secret,pkt["User-Name"][0])
                log.msg("send ikuai radius Coa Request: %s"%(repr(pkg)),level=logging.INFO)
                self.transport.write(pkg,(self.addr, self.port))
            else:
                self.transport.write(pkt.RequestPacket(),(self.addr, self.port))
        except packet.PacketError as err:
            log.err(err,'::send radius Coa Request error %s: %s'%((host,self.port),str(err)))

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
    def __init__(self, radiusd):
        _dict = dictionary.Dictionary(radiusd.dictfile)
        host.Host.__init__(self,dict=_dict)
        self.debug = radiusd.debug
        self.user_trace = radiusd.trace
        self.midware = radiusd.midware
        self.runstat = radiusd.runstat
        self.coa_clients = radiusd.coa_clients
        self.store = radiusd.store
        self.auth_delay = utils.AuthDelay(int(self.store.get_param("reject_delay") or 0))

    def processPacket(self, pkt):
        pass

    def createPacket(self, **kwargs):
        raise NotImplementedError('Attempted to use a pure base class')

    def datagramReceived(self, datagram, (host, port)):
        bas = self.store.get_bas(host)
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
        user = self.store.get_user(req.get_user_name())
        if user:self.user_trace.push(user['account_number'],req)
        # middleware execute
        for plugin in auth_plugins:
            self.midware.process(plugin,req=req,resp=reply,user=user,radiusd=self)
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
            self.midware.process(plugin,req=req,radiusd=self)

        user = self.store.get_user(req.get_user_name())
        if user:self.user_trace.push(user['account_number'],req)        
          
        reply = req.CreateReply()
        reply.source = req.source
        if user:self.user_trace.push(user['account_number'],reply)   
        req.deferred.callback(reply)
        # middleware execute
        for plugin in acct_plugins:
            self.midware.process(plugin,req=req,user=user,radiusd=self)
        
 
 ###############################################################################
 # admin  Server                                                            ####
 ###############################################################################
 
class AdminServerProtocol(WebSocketServerProtocol):

    radiusd = None

    def onConnect(self, request):
        log.msg("Client connecting: {0}".format(request.peer))

    def onOpen(self):
        log.msg("WebSocket connection open.")

    def onMessage(self, payload, isBinary):
        req_msg = None
        try:
            _msg = utils.decrypt(payload)
            req_msg = json.loads(_msg)
        except:
            log.err('decrypt message error : %s'%payload)
        
        if req_msg:
            log.msg("websocket admin request: %s"%str(req_msg))
            plugin = req_msg.get("process")
            self.radiusd.midware.process(plugin,req=req_msg,admin=self)

    def onClose(self, wasClean, code, reason):
        log.msg("WebSocket connection closed: {0}".format(reason))

###############################################################################
# Radius  Server                                                              ####
###############################################################################    

class RadiusServer(object):
    
    def __init__(self,config,db_engine=None,daemon=False):
        self.config = config
        self.db_engine = db_engine
        self.daemon = daemon
        self.tasks = {}
        self.init_config()
        self.init_timezone()
        self.init_db_engine()
        self.init_protocol()
        self.init_task()
        
    def _check_ssl_config(self):
        self.use_ssl = False
        self.privatekey = None
        self.certificate = None
        if self.config.has_option('DEFAULT','ssl') and self.config.getboolean('DEFAULT','ssl'):
            self.privatekey = self.config.get('DEFAULT','privatekey')
            self.certificate = self.config.get('DEFAULT','certificate')
            if os.path.exists(self.privatekey) and os.path.exists(self.certificate):
                self.use_ssl = True
        
    def init_config(self):
        self.logfile = self.config.get('radiusd','logfile')
        self.standalone = self.config.has_option('DEFAULT','standalone') and \
            self.config.getboolean('DEFAULT','standalone') or False
        self.secret = self.config.get('DEFAULT','secret')
        self.timezone = self.config.has_option('DEFAULT','tz') and self.config.get('DEFAULT','tz') or "CST-8"
        self.debug = self.config.getboolean('DEFAULT','debug')
        self.authport = self.config.getint('radiusd','authport')
        self.acctport = self.config.getint('radiusd','acctport')
        self.adminport = self.config.getint('radiusd','adminport')
        self.radiusd_host = self.config.has_option('radiusd','host') \
            and self.config.get('radiusd','host') or  '0.0.0.0'
        #parse dictfile
        self.dictfile = os.path.join(os.path.split(__file__)[0],'dicts/dictionary')
        if self.config.has_option('radiusd','dictfile'):
            if os.path.exists(self.config.get('radiusd','dictfile')):
                self.dictfile = self.config.get('radiusd','dictfile')

        # update aescipher
        utils.aescipher.setup(self.secret)
        self.encrypt = utils.aescipher.encrypt
        self.decrypt = utils.aescipher.decrypt
        #parse ssl
        self._check_ssl_config()
        
    def init_timezone(self):
        try:
            os.environ["TZ"] = self.timezone
            time.tzset()
        except:pass
    
    def init_db_engine(self):
        if not self.db_engine:
            self.db_engine = get_engine(self.config)
        self.store = Store(self.config,self.db_engine)

    def reload_coa_clients(self):
        for bas in self.store.list_bas():

            if bas['ip_addr'] in self.coa_clients:
                self.coa_clients[bas['ip_addr']].close()

            self.coa_clients[bas['ip_addr']] = CoAClient(
                bas,
                dictionary.Dictionary(self.dictfile),
                debug=self.debug
            )
        
    def init_protocol(self):
        # rundata
        self.trace = utils.UserTrace()
        self.runstat = utils.RunStat()
        self.midware = middleware.Middleware()
        # init coa clients
        self.coa_clients = {}
        for bas in self.store.list_bas():
            self.coa_clients[bas['ip_addr']] = CoAClient(
                bas,
                dictionary.Dictionary(self.dictfile),
                debug=self.debug
            )
        self.auth_protocol = RADIUSAccess(self)
        self.acct_protocol = RADIUSAccounting(self)
        
        ws_url = "ws://%s:%s"%(self.radiusd_host,self.adminport)
        if self.use_ssl:
            ws_url = "wss://%s:%s"%(self.radiusd_host,self.adminport)

        self.admin_factory = WebSocketServerFactory(ws_url, debug = False)
        self.admin_factory.protocol = AdminServerProtocol
        self.admin_factory.setProtocolOptions(allowHixie76=True)
        self.admin_factory.protocol.radiusd = self
        
    def _check_online_over(self):
        reactor.callInThread(self.store.check_online_over)

    def init_task(self):
        _task = task.LoopingCall(self.auth_protocol.process_delay)
        _task.start(2.7)
        _online_task = task.LoopingCall(self._check_online_over)
        _online_task.start(3600*4)
        _msg_stat_task = task.LoopingCall(self.runstat.run_stat)
        _msg_stat_task.start(60)
        self.tasks['process_delay'] = _task
        self.tasks['check_online_over'] = _online_task
        
    def run_normal(self):
        if self.debug:
            log.startLogging(sys.stdout)
        else:
            log.startLogging(DailyLogFile.fromFullPath(self.logfile))
        log.msg('server listen %s'%self.radiusd_host)  
        reactor.listenUDP(self.authport, self.auth_protocol,interface=self.radiusd_host)
        reactor.listenUDP(self.acctport, self.acct_protocol,interface=self.radiusd_host)
        if self.use_ssl:
            log.msg('WS SSL Enable!')
            from twisted.internet import ssl
            sslContext = ssl.DefaultOpenSSLContextFactory(self.privatekey, self.certificate)
            reactor.listenSSL(
                self.adminport,
                self.admin_factory,
                contextFactory = sslContext,
                interface=self.radiusd_host
            )
        else:
            reactor.listenTCP(self.adminport, self.admin_factory,interface=self.radiusd_host)
        if not self.standalone:
            reactor.run()

    def get_coa_client(self,nasaddr):
        cli = self.coa_clients.get(nasaddr)
        if not cli:
            bas = self.store.get_bas(nasaddr)
            if bas:
                cli = CoAClient(
                    bas,
                    dictionary.Dictionary(self.dictfile),
                    debug=self.debug
                )
                self.coa_clients[nasaddr] = cli
        return cli


    def get_service(self):
        from twisted.application import service, internet
        top_service = service.MultiService()
        
        internet.UDPServer(
            self.authport,self.auth_protocol,interface=self.radiusd_host
        ).setServiceParent(top_service)
        
        internet.UDPServer(
            self.acctport, self.acct_protocol,interface=self.radiusd_host
        ).setServiceParent(top_service)
        
        if self.use_ssl:
            log.msg('WS SSL Enable!')
            from twisted.internet import ssl
            sslContext = ssl.DefaultOpenSSLContextFactory(self.privatekey, self.certificate)
            internet.SSLServer(
                self.adminport,
                self.admin_factory,
                contextFactory = sslContext,
                interface=self.radiusd_host
            ).setServiceParent(top_service)
        else:
            log.msg('WS SSL Disable!')       
            internet.TCPServer(
                self.adminport,
                self.admin_factory,
                interface=self.radiusd_host
            ).setServiceParent(top_service)
        return top_service


def run(config,db_engine=None,is_serrvice=False):
    print 'running radiusd server...'
    radiusd = RadiusServer(config,db_engine,daemon=is_serrvice)
    if is_serrvice:
        return radiusd.get_service()
    else:
        radiusd.run_normal()