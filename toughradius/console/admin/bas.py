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
from toughradius.console.admin import bas_forms
from hashlib import md5
from twisted.python import log
import bottle
import datetime
import json
import functools

__prefix__ = "/bas"

app = Bottle()
app.config['__prefix__'] = __prefix__

###############################################################################
# bas manage
###############################################################################

@app.route('/', apply=auth_opr, method=['GET', 'POST'])
def bas(db, render):
    return render("sys_bas_list",
                  bastype=bas_forms.bastype,
                  bas_list=db.query(models.SlcRadBas))


@app.get('/add', apply=auth_opr)
def bas_add(db, render):
    return render("base_form", form=bas_forms.bas_add_form())


@app.post('/add', apply=auth_opr)
def bas_add_post(db, render):
    form = bas_forms.bas_add_form()
    if not form.validates(source=request.forms):
        return render("base_form", form=form)
    if db.query(models.SlcRadBas.id).filter_by(ip_addr=form.d.ip_addr).count() > 0:
        return render("base_form", form=form, msg=u"Bas地址已经存在")
    bas = models.SlcRadBas()
    bas.ip_addr = form.d.ip_addr
    bas.bas_name = form.d.bas_name
    bas.time_type = form.d.time_type
    bas.vendor_id = form.d.vendor_id
    bas.bas_secret = form.d.bas_secret
    bas.coa_port = form.d.coa_port
    db.add(bas)

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)新增BAS信息:%s' % (get_cookie("username"), bas.ip_addr)
    db.add(ops_log)

    db.commit()
    redirect("/bas")


@app.get('/update', apply=auth_opr)
def bas_update(db, render):
    bas_id = request.params.get("bas_id")
    form = bas_forms.bas_update_form()
    form.fill(db.query(models.SlcRadBas).get(bas_id))
    return render("base_form", form=form)


@app.post('/update', apply=auth_opr)
def bas_add_update(db, render):
    form = bas_forms.bas_update_form()
    if not form.validates(source=request.forms):
        return render("base_form", form=form)
    bas = db.query(models.SlcRadBas).get(form.d.id)
    bas.bas_name = form.d.bas_name
    bas.time_type = form.d.time_type
    bas.vendor_id = form.d.vendor_id
    bas.bas_secret = form.d.bas_secret
    bas.coa_port = form.d.coa_port

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)修改BAS信息:%s' % (get_cookie("username"), bas.ip_addr)
    db.add(ops_log)

    db.commit()
    websock.update_cache("bas", ip_addr=bas.ip_addr)
    redirect("/bas")

@app.get('/delete', apply=auth_opr)
def bas_delete(db, render):
    bas_id = request.params.get("bas_id")
    db.query(models.SlcRadBas).filter_by(id=bas_id).delete()

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)删除BAS信息:%s' % (get_cookie("username"), bas_id)
    db.add(ops_log)

    db.commit()
    redirect("/bas")

permit.add_route("/bas", u"BAS信息管理", u"系统管理", is_menu=True, order=2)
permit.add_route("/bas/add", u"新增BAS", u"系统管理", order=2.01, is_open=False)
permit.add_route("/bas/update", u"修改BAS", u"系统管理", order=2.02, is_open=False)
permit.add_route("/bas/delete", u"删除BAS", u"系统管理", order=2.03, is_open=False)

