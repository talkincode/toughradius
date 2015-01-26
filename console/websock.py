#!/usr/bin/env python
#coding:utf-8
import json
from bottle import MakoTemplate
from twisted.internet import reactor  
from twisted.internet.protocol import ReconnectingClientFactory
from autobahn.twisted.websocket import WebSocketClientProtocol, WebSocketClientFactory 

class WebSockProtocol(WebSocketClientProtocol):  

    messages = []

    def onConnect(self, response):
        print("Radius Admin Server connected: {0}".format(response.peer))

    def onOpen(self): 
        def send_message():
            if self.messages:
                self.sendMessage(self.messages.pop(),False) 
            self.factory.reactor.callLater(1, send_message) 
        send_message()

    def onMessage(self, msg, binary):  
        print "Got: " + msg  

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

    def connect(self,radaddr,adminport):    
        self.factory = WSClientFactory("ws://%s:%s"%(radaddr,adminport), debug = False)  
        reactor.connectTCP(radaddr, int(adminport), self.factory)
        
    def reconnect(self,radaddr,adminport):    
        self.connect(radaddr,adminport)

    def update_cache(self,cache_class,**kwargs):
        message = {
            'process' : "admin_update_cache",
            'cache_class' : cache_class
        }
        message.update(**kwargs)
        self.factory.protocol.messages.append(json.dumps(message).encode("utf-8")) 

    def invoke_admin(self,ops,**kwargs):
        message = {
            'process' : "admin_%s"%ops
        }
        message.update(**kwargs)
        self.factory.protocol.messages.append(json.dumps(message).encode("utf-8")) 


websock = WebSock()




