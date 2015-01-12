#!/usr/bin/env python
#coding=utf-8
from plugins import error_auth
import utils

def process(req=None,resp=None,user=None):

    if not user:
        return error_auth(resp,'user %s not exists'%req.get_user_name())

    if not req.is_valid_pwd(utils.decrypt(user['password'])):
        return error_auth(resp,'user password not match')

    if not user['status'] == 1:
        return error_auth(resp,'user status not ok')

    return resp