#!/usr/bin/env python
#coding:utf-8

import json
from bottle import MakoTemplate
from twisted.internet import reactor  
from twisted.internet.protocol import ReconnectingClientFactory
from autobahn.twisted.websocket import WebSocketClientProtocol, WebSocketClientFactory 
from toughradius.console.libs import utils

class WebSockProtocol(WebSocketClientProtocol):  

    messages = []
    callbacks = {}

    def onConnect(self, response):
        print("Radius Server connected: {0}".format(response.peer))

    def onOpen(self): 
        def send_message():
            if self.messages:
                self.sendMessage(self.messages.pop(),False) 
            self.factory.reactor.callLater(0.1, send_message) 
        send_message()

    def onMessage(self, msg, binary):  
        print "Got: " + msg  
        resp = json.loads(msg)
        if 'msg_id' in resp:
            callback = self.callbacks.get(resp['msg_id'])
            if callback and callable(callback):callback(resp)
            
    def onClose(self, wasClean, code, reason):
        print("WebSocket connection closed: {0}".format(reason))

class WSClientFactory(WebSocketClientFactory, ReconnectingClientFactory):

   protocol = WebSockProtocol

   def clientConnectionFailed(self, connector, reason):
      print("Client connection failed .. retrying ..")
      self.retry(connector)

   def clientConnectionLost(self, connector, reason):
      print("Client connection lost .. retrying ..")
      self.retry(connector)

class WebSock():    
    
    use_ssl = False

    def connect(self,radaddr,adminport): 
        def _connect_std():
            self.factory = WSClientFactory("ws://%s:%s"%(radaddr,adminport), debug = False)  
            reactor.connectTCP(radaddr, int(adminport), self.factory)
            
        if self.use_ssl:
            try:
                from twisted.internet import ssl
                contextFactory = ssl.ClientContextFactory()
                self.factory = WSClientFactory("wss://%s:%s"%(radaddr,adminport), debug = False)  
                reactor.connectSSL(radaddr, int(adminport), self.factory,contextFactory)
            except:
                _connect_std()
        else:
            _connect_std()
            

    def reconnect(self,radaddr,adminport):    
        self.connect(radaddr,adminport)

    def update_cache(self,cache_class,**kwargs):
        msg_id = utils.CurrentID()
        message = {
            'msg_id'  : msg_id,
            'process' : "admin_update_cache",
            'cache_class' : cache_class
        }
        callback = 'callback' in kwargs and kwargs.pop('callback') or None
        message.update(**kwargs)
        secure_msg = utils.encrypt(json.dumps(message).encode("utf-8"))
        self.factory.protocol.messages.append(secure_msg) 
        if callback:self.factory.protocol.callbacks[msg_id] = callback

    def invoke_admin(self,ops,**kwargs):
        msg_id = utils.CurrentID()
        message = {
            'msg_id'  : msg_id,
            'process' : "admin_%s"%ops
        }
        callback = 'callback' in kwargs and kwargs.pop('callback') or None
        message.update(**kwargs)
        secure_msg = utils.encrypt(json.dumps(message).encode("utf-8"))
        self.factory.protocol.messages.append(secure_msg) 
        if callback:self.factory.protocol.callbacks[msg_id] = callback

websock = WebSock()

