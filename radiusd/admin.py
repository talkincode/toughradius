#!/usr/bin/env python
#coding=utf-8
from autobahn.twisted.websocket import WebSocketServerProtocol
from twisted.python import log
import logging
import json
import datetime

now_time = lambda: datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")

class UserTrace():

    def __init__(self,user_size=10,gl_size=20):
        self.user_size = user_size
        self.gl_szie = gl_size
        self.user_cache = {}
        self.global_trace = []
        
    def size_info(self):
        return len(self.user_cache),len(self.global_trace)
        
    def push(self,username,pkt):
        _cache = self.user_cache.get(username)
        if not _cache:
            _cache = self.user_cache[username] = []
            
        if len(_cache) >= self.user_size:
            _cache.pop()
        _cache.insert(0,pkt)
        
        if len(self.global_trace) >= self.gl_szie:
            self.global_trace.pop()
        self.global_trace.insert(0,pkt)
        
    def get_global_msg(self):
        if len(self.global_trace) == 0:return None
        return self.global_trace.pop()
        
    def get_user_msg(self,username):
        return self.user_cache.get(username) or []

     
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
        # log.msg("trace size info %s,%s"%self.user_trace.size_info(),level=logging.DEBUG)
        plugin = req_msg.get("process")
        # trace=self.user_trace,send=self.sendMessage
        self.midware.process(plugin,req=req_msg,admin=self)

    def onClose(self, wasClean, code, reason):
        log.msg("WebSocket connection closed: {0}".format(reason))
