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
        raise abort(403, u"%s not exists" % username)
    if password != app.config['control.passwd']:
        raise abort(403, u"password not match")
    set_cookie('control_admin', username, path="/")
    set_cookie('control_admin_ip', request.remote_addr, path="/")
    redirect("/")


@app.get("/logout")
def ctl_logout(render):
    set_cookie('control_admin', None)
    set_cookie('control_admin_ip', None)
    request.cookies.clear()
    redirect('/auth/login')


