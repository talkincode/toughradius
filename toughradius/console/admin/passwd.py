#!/usr/bin/env python
#coding=utf-8

import sys, os
from bottle import Bottle
from bottle import request
from bottle import response
from bottle import redirect
from bottle import MakoTemplate
from bottle import static_file
from bottle import abort
from beaker.cache import cache_managers
from toughradius.console.libs.paginator import Paginator
from toughradius.console.libs import utils
from toughradius.console.websock import websock
from toughradius.console import models
from toughradius.console.base import *
from toughradius.console.admin import forms
from hashlib import md5
from twisted.python import log
import bottle
import datetime
import json
import functools

__prefix__ = "/passwd"

app = Bottle()
app.config['__prefix__'] = __prefix__


###############################################################################
# password update
###############################################################################

@app.get('/', apply=auth_opr)
def passwd(db, render):
    form = forms.passwd_update_form()
    form.fill(operator_name=get_cookie("username"))
    return render("base_form", form=form)


@app.post('/', apply=auth_opr)
def passwd_update(db, render):
    form = forms.passwd_update_form()
    if not form.validates(source=request.forms):
        return render("base_form", form=form)
    if form.d.operator_pass != form.d.operator_pass_chk:
        return render("base_form", form=form, msg=u"确认密码不一致")
    opr = db.query(models.SlcOperator).filter_by(operator_name=form.d.operator_name).first()
    opr.operator_pass = md5(form.d.operator_pass).hexdigest()

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)修改密码' % (get_cookie("username"),)
    db.add(ops_log)

    db.commit()
    redirect("/passwd")

