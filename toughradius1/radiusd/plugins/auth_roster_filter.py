#!/usr/bin/env python
#coding=utf-8
from toughradius.radiusd.plugins import error_auth

def process(req=None,resp=None,user=None,radiusd=None,**kwargs):
    """check block roster"""
    store = radiusd.store
    if store.is_black_roster(req.get_mac_addr()):
        return error_auth(resp,"user macaddr in blacklist")
    return resp