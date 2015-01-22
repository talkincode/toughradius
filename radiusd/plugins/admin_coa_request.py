#!/usr/bin/env python
#coding=utf-8
from store import store
import json

def process(req=None,admin=None):
    nas_addr = req.get("nas_addr") 
    acct_session_id = req.get("acct_session_id")
    message_type  =req.get("message_type")

    if not nas_addr or not acct_session_id:
        reply = {'code':1,'data':u"nas_addr and acct_session_id Does not allow nulls"}
        return admin.sendMessage(json.dumps(reply),False)

    coa_client = admin.coa_clients.get(nas_addr)
    if not coa_client:
        reply = {'code':1,'data':u"CoA Client instance not exists for %s"%nas_addr}
        return admin.sendMessage(json.dumps(reply),False)

    online = store.get_online(nas_addr,acct_session_id)
    if not online:
        reply = {'code':1,'data':u"online not exists"}
        return admin.sendMessage(json.dumps(reply),False)

    if message_type == 'disconnect':
        attrs = {
            'User-Name' : online['account_number'],
            'Acct-Session-Id' : acct_session_id,
            'NAS-IP-Address' : nas_addr,
            'Framed-IP-Address' : online['framed_ipaddr']
        }
        dmeq = coa_client.createDisconnectPacket(**attrs)
        coa_client.sendCoA(dmeq)
        reply = {'code':0,'data':u"disconnect message send"}
        admin.sendMessage(json.dumps(reply),False)

