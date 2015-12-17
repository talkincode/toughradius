#!/usr/bin/env python
# coding:utf-8
import sys, os
from bottle import Bottle
from bottle import request
from bottle import redirect
from bottle import abort
from hashlib import md5
from toughradius.console.base import *
from toughradius.console.libs import utils
import time
import bottle
import functools

__prefix__ = "/auth"

app = Bottle()
app.config['__prefix__'] = __prefix__

###############################################################################
# user login
###############################################################################

@app.get('/login')
def ctl_login(render):
    return render("login")


@app.post('/login')
def ctl_login(render):
    username = request.params.get("username")
    password = request.params.get("password")
    if username != app.config['control.user']:
        return dict(code=1,msg="user: %s is not exist" %username)
    if password != app.config['control.passwd']:
        return dict(code=1,msg="password cannot match")
    set_cookie('control_admin', username, path="/")
    set_cookie('control_admin_ip', request.remote_addr, path="/")
    return dict(code=0,msg="OK")


@app.get("/logout")
def ctl_logout(render):
    set_cookie('control_admin', None)
    set_cookie('control_admin_ip', None)
    request.cookies.clear()
    redirect('/auth/login')


