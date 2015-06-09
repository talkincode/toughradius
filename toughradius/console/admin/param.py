#!/usr/bin/env python
# coding=utf-8
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

__prefix__ = "/param"

app = Bottle()
app.config['__prefix__'] = __prefix__


###############################################################################
# param config
###############################################################################

@app.get('/', apply=auth_opr)
def param(db, render):
    form = forms.param_form()
    fparam = {}
    for p in db.query(models.SlcParam):
        fparam[p.param_name] = p.param_value
    form.fill(fparam)
    return render("sys_param", form=form)


@app.post('/update', apply=auth_opr)
def param_update(db, render):
    params = db.query(models.SlcParam)
    for param_name in request.forms:
        if 'submit' in param_name:
            continue
        param = db.query(models.SlcParam).filter_by(param_name=param_name).first()
        if not param:
            param = models.SlcParam()
            param.param_name = param_name
            param.param_value = request.forms.get(param_name)
            db.add(param)
        else:
            param.param_value = request.forms.get(param_name)

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)修改参数' % (get_cookie("username"))
    db.add(ops_log)
    db.commit()

    websock.reconnect(
        request.forms.get('radiusd_address'),
        request.forms.get('radiusd_admin_port'),
    )

    is_debug = request.forms.get('is_debug')
    bottle.debug(is_debug == '1')

    websock.update_cache("is_debug", is_debug=is_debug)
    websock.update_cache("reject_delay", reject_delay=request.forms.get('reject_delay'))
    websock.update_cache("param")
    redirect("/param")


permit.add_route("/param", u"系统参数管理", MenuSys, is_menu=True, order=0.0001)
permit.add_route("/param/update", u"系统参数修改", MenuSys, is_menu=False, order=0.0002, is_open=False)
