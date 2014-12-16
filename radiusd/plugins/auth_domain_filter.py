#!/usr/bin/env python
#coding=utf-8
from plugins import error_auth

def process(req=None,resp=None,user=None):
    domain = req.get_domain()
    user_domain  = user.get("domain_name")
    if domain and user_domain:
        if domain not in user_domain:
            return error_auth(resp,'user domain %s not match'%domain  )
    return resp    