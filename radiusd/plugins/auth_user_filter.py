#!/usr/bin/env python
#coding=utf-8
from plugins import error_auth
from store import store
import utils

def process(req=None,resp=None,user=None):

    if not user:
        return error_auth(resp,'user %s not exists'%req.get_user_name())

    if not req.is_valid_pwd(utils.decrypt(user['password'])):
        return error_auth(resp,'user password not match')
        
    if user['status'] == 4:
        resp['Framed-Pool'] = store.get_param("9_expire_addrpool")
        return resp

    if  user['status'] in (0,2,3):
        return error_auth(resp,'user status not ok')

    return resp