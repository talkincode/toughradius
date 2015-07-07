#!/usr/bin/env python
# coding:utf-8
import sys, os
from bottle import Bottle
from bottle import request
from bottle import redirect
from hashlib import md5
from toughradius.console.base import *
from toughradius.console.libs import utils
from toughradius.console.libs.validate import vcache
from toughradius.console import models
from toughradius.console.customer import forms
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
def member_login_get(db, render):
    form = forms.member_login_form()
    form.next.set_value(request.params.get('next', '/'))
    return render("login", form=form)


@app.post('/login')
def member_login_post(db, render):
    next = request.params.get("next", "/")
    form = forms.member_login_form()
    if not form.validates(source=request.params):
        return render("login", form=form)

    if vcache.is_over(form.d.username, '0'):
        return render("error", msg=u"用户一小时内登录错误超过5次，请一小时后再试")

    member = db.query(models.SlcMember).filter_by(
        member_name=form.d.username
    ).first()

    if not member:
        return render("login", form=form, msg=u"用户不存在")

    if member.password != md5(form.d.password.encode()).hexdigest():
        vcache.incr(form.d.username, '0')
        print vcache.validates
        return render("login", form=form, msg=u"用户名密码错误第%s次" % vcache.errs(form.d.username, '0'))

    vcache.clear(form.d.username, '0')

    set_cookie('customer_id', member.member_id, path="/")
    set_cookie('customer', form.d.username,path="/")
    set_cookie('customer_login_time', utils.get_currtime(), path="/")
    set_cookie('customer_login_ip', request.remote_addr, path="/")
    redirect(next)


@app.get("/logout")
def member_logout(db, render):
    set_cookie('customer_id', None,path="/")
    set_cookie('customer', None, path="/")
    set_cookie('customer_login_time', None, path="/")
    set_cookie('customer_login_ip', None, path="/")
    request.cookies.clear()
    redirect('login')
