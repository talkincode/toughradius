#!/usr/bin/env python
#coding=utf-8
from twisted.python import log
from store import store
from settings import *
import logging
import json

def process(req=None,trace=None,send=None):
    nas_addr = req.get("nas_addr") 
    if not nas_addr:
        reply = json.dumps({'data':'nas_addr is empty'.code:1})
        return send(reply,False) 
    session_id = req.get("session_id")
    store.unlock_online(nas_addr,session_id,STATUS_TYPE_UNLOCK)
    reply = json.dumps({'data':'unlock ok'.code:0})
    send(reply,False)
