#!/usr/bin/env python
#coding:utf-8

from bottle import Bottle
from bottle import request
from bottle import response
from bottle import redirect
from bottle import run as runserver
from bottle import static_file
from bottle import abort
from bottle import mako_template as render
from hashlib import md5
from tablib import Dataset
from base import *
from libs import utils
import bottle
import models
import forms
import decimal
import datetime

app = Bottle()

###############################################################################
# login handle         
###############################################################################

@app.get('/login')
def member_login_get(db):
    return render("member_login")

@app.post('/login')
def member_login_post(db):
    uname = request.forms.get("username")
    upass = request.forms.get("password")
    if not uname:return dict(code=1,msg=u"请填写用户名")
    if not upass:return dict(code=1,msg=u"请填写密码")
    enpasswd = md5(upass.encode()).hexdigest()
    member = db.query(models.SlcMember).filter_by(
        member_name=uname,
        password=enpasswd
    ).first()
    if not member:return dict(code=1,msg=u"用户名密码不符")
    set_cookie('member',uname)
    set_cookie('member_login_time', utils.get_currtime())
    set_cookie('member_login_ip', request.remote_addr)    
    return dict(code=0,msg="ok")

@app.get("/logout")
def member_logout():
    request.cookies.clear()
    redirect('/member/login')


