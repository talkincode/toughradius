#!/usr/bin/env python
#coding=utf-8
from toughradius.radiusd.plugins import error_auth
from toughradius.radiusd import utils

def process(req=None,resp=None,user=None,radiusd=None,**kwargs):
    store = radiusd.store

    if store.is_white_roster(req.get_mac_addr()):
        return resp

    if not user:
        return error_auth(resp,'user %s not exists'%req.get_user_name())

    if store.get_param("radiusd_bypass") in ('1', None, ''):
        if not req.is_valid_pwd(utils.decrypt(user['password'])):
            return error_auth(resp, 'user password not match')
        
    if user['status'] == 4:
        resp['Framed-Pool'] = store.get_param("expire_addrpool")
        return resp

    if  user['status'] in (0,2,3):
        return error_auth(resp,'user status not ok')

    return resp