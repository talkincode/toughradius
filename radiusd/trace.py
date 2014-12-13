#!/usr/bin/env python
#coding=utf-8
from autobahn.twisted.websocket import WebSocketServerProtocol
from twisted.python import log
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
        

class TraceServerProtocol(WebSocketServerProtocol):

    user_trace = None

    def onConnect(self, request):
        log.msg("Client connecting: {0}".format(request.peer))

    def onOpen(self):
        log.msg("WebSocket connection open.")

    def onMessage(self, payload, isBinary):
        try:
            req_msg = json.loads(payload)
            log.msg("websocket trace query: %s"%str(req_msg))
            log.msg("trace size info %s,%s"%self.user_trace.size_info())
            if req_msg['scope'] == 'global':
                pkt = self.user_trace.get_global_msg()
                if pkt is None: return
                mtype = int(req_msg.get('type'))
                username = req_msg.get("username")
                basaddr = req_msg.get("bas")
                if mtype:
                    if mtype in (1,) and pkt.code not in (1,2,3):return
                    if mtype in (4,) and pkt.code not in (4,5):return    
                if username:
                    if pkt.code in (1,4) and username not in pkt.get_user_name():return
                    if pkt.code in (2,3,5) and username not in pkt.source_user:return
                if basaddr:
                    if basaddr not in pkt.source[0]:return
                reply = {'data' : pkt.format_str(),'time':pkt.created,'host':pkt.source}
                msg = json.dumps(reply)
                msg = msg.replace("\\n","<br>")
                msg = msg.replace("\\t","    ")
                self.sendMessage(msg, False)
                
            elif req_msg['scope'] == 'user':
                if not req_msg.get("username"):
                    reply = json.dumps({'data':'username is empty','time':now_time(),'host':''})
                    return self.sendMessage(reply, False)
                
                pkts = self.user_trace.get_user_msg(req_msg['username'])
                reply = json.dumps({'data':'no messages','time':now_time(),'host':''})
                if not pkts:self.sendMessage(reply, False)
                for pkt in pkts:
                    reply = {'data' : pkt.format_str(),'time':pkt.created,'host':pkt.source}
                    msg = json.dumps(reply)
                    msg = msg.replace("\\n","<br>")
                    msg = msg.replace("\\t","    ")
                    self.sendMessage(msg, False)

        except Exception as err:
            import traceback
            traceback.print_exc()
            reply = json.dumps({'data':'trace msg error %s'%str(err),'time':now_time(),'host':''})
            self.sendMessage(reply, False)

    def onClose(self, wasClean, code, reason):
        log.msg("WebSocket connection closed: {0}".format(reason))
