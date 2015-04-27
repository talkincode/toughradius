#!/usr/bin/env python
#coding=utf-8
from twisted.python import log
import logging
import json

def process(req=None,admin=None,**kwargs):
    msg_id = req.get("msg_id")
    pkt = admin.radiusd.trace.get_global_msg()
    if pkt is None: 
        return
    mtype = int(req.get('type'))
    username = req.get("username")
    basaddr = req.get("bas")
    if mtype:
        if mtype in (1,) and pkt.code not in (1,2,3):return
        if mtype in (4,) and pkt.code not in (4,5):return    
    if username:
        if pkt.code in (1,4) and username not in pkt.get_user_name():return
        if pkt.code in (2,3,5) and username not in pkt.source_user:return
    if basaddr:
        if basaddr not in pkt.source[0]:return
    reply = {'msg_id':msg_id,'data' : pkt.format_str(),'time':pkt.created.strftime("%Y-%m-%d %H:%M:%S"),'host':pkt.source}
    msg = json.dumps(reply)
    msg = msg.replace("\\n","<br>")
    msg = msg.replace("\\t","    ")
    admin.sendMessage(msg,False) 


