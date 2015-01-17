#!/usr/bin/env python
#coding=utf-8
from twisted.python import log
import logging
import json
import datetime


def process(req=None,admin=None):
    now_time = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    if not req.get("username"):
        reply = json.dumps({'data':'username is empty','time':now_time,'host':''})
        return admin.sendMessage(reply,False) 
    
    pkts = admin.user_trace.get_user_msg(req['username'])
    reply = json.dumps({'data':'no messages','time':now_time,'host':''})
    if not pkts:
        return admin.sendMessage(reply,False) 

    for pkt in pkts:
        reply = {'data' : pkt.format_str(),'time':pkt.created.strftime("%Y-%m-%d %H:%M:%S"),'host':pkt.source}
        msg = json.dumps(reply)
        msg = msg.replace("\\n","<br>")
        msg = msg.replace("\\t","    ")
        admin.sendMessage(msg, False)    
