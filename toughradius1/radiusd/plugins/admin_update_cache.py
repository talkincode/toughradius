#!/usr/bin/env python
#coding=utf-8
from twisted.python import log
from toughradius.radiusd.settings import *
from toughradius.radiusd import settings
import json

cache_class = []

def process(req=None,admin=None,**kwargs):
    msg_id = req.get("msg_id")
    cache_class = req.get("cache_class") 
    if not cache_class:
        reply = json.dumps({'msg_id':msg_id,'data':u'cache_class is empty','code':1})
        return admin.sendMessage(reply,False) 

    def send_ok(op):
        reply = json.dumps({'msg_id':msg_id,'data':u'%s ok'%op,'code':0})
        admin.sendMessage(reply,False)
    
    if cache_class == 'all':
        admin.radiusd.store.update_all_cache()
        send_ok("all cache update")
    elif cache_class == 'param':
        admin.radiusd.store.update_param_cache()
        send_ok("param cache update")
    elif cache_class == 'account' and req.get("account_number"):
        admin.radiusd.store.update_user_cache(req.get("account_number"))
        send_ok("account cache update")
    elif cache_class == 'bas' and req.get("ip_addr"):
        admin.radiusd.store.update_bas_cache(req.get("ip_addr"))
        send_ok("bas cache update")
    elif cache_class == 'roster' and req.get("mac_addr"):
        admin.radiusd.store.update_roster_cache(req.get("mac_addr"))  
        send_ok("roster cache update")     
    elif cache_class == 'product' and req.get("product_id"):
        admin.radiusd.store.update_product_cache(req.get("product_id"))
        send_ok("product cache update")
    elif cache_class == 'is_debug' and req.get("is_debug"):
        _is_debug = bool(int(req.get("is_debug"))) 
        admin.radiusd.auth_protocol.debug = _is_debug
        admin.radiusd.acct_protocol.debug = _is_debug
        send_ok("radiusd debug mode update")
    elif cache_class == 'reject_delay' and req.get("reject_delay"):
        try:
            _delay = int(req.get("reject_delay"))
            if _delay >= 0 and _delay <= 9:
                admin.radiusd.auth_protocol.auth_delay.reject_delay = _delay
            send_ok("reject_delay update")
        except:
            reply = json.dumps({'msg_id':msg_id,'data':u'error reject_delay param','code':0})
            admin.sendMessage(reply,False)
    else:
        reply = json.dumps({'msg_id':msg_id,'data':u'do nothing','code':0})
        admin.sendMessage(reply,False)



