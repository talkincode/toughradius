#!/usr/bin/env python
#coding:utf-8
from toughradius.common.bottle import route, request,post

@post('/api/v1/user/add')
def adduser():
    return dict(code=0,msg="success")
