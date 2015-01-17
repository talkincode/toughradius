#!/usr/bin/env python
#coding=utf-8
from twisted.python import log
from store import store
from settings import *
import logging
import json

def process(req=None,admin=None):
    cache_class = req.get("cache_class") 
    if not cache_class:
        reply = json.dumps({'data':u'cache_class is empty','code':1})
        return admin.sendMessage(reply,False) 

    def send_ok(op):
        reply = json.dumps({'data':u'%s ok'%op,'code':0})
        admin.sendMessage(reply,False)
    
    if cache_class == 'param':
        store.update_param_cache()
        send_ok("param cache update")
    elif cache_class == 'account' and req.get("account_number"):
        store.update_user_cache(req.get("account_number"))
        send_ok("account cache update")
    elif cache_class == 'bas' and req.get("ip_addr"):
        store.update_bas_cache(req.get("ip_addr"))
        send_ok("bas cache update")
    elif cache_class == 'group' and req.get("group_id"):
        store.update_group_cache(req.get("group_id"))  
        send_ok("group cache update")      
    elif cache_class == 'product' and req.get("product_id"):
        store.update_product_cache(req.get("product_id"))
        send_ok("product cache update")
    else:
        reply = json.dumps({'data':u'do nothing','code':0})
        admin.sendMessage(reply,False)



