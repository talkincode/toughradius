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
from toughradius.console.admin import roster_forms
from hashlib import md5
from twisted.python import log
import bottle
import datetime
import json
import functools

__prefix__ = "/roster"

app = Bottle()
app.config['__prefix__'] = __prefix__

###############################################################################
# roster manage
###############################################################################

@app.route('/', apply=auth_opr, method=['GET', 'POST'])
def roster(db, render):
    _query = db.query(models.SlcRadRoster)
    return render("sys_roster_list",
                  page_data=get_page_data(_query))


@app.get('/add', apply=auth_opr)
def roster_add(db, render):
    return render("sys_roster_form", form=roster_forms.roster_add_form())


@app.post('/add', apply=auth_opr)
def roster_add_post(db, render):
    form = roster_forms.roster_add_form()
    if not form.validates(source=request.forms):
        return render("sys_roster_form", form=form)
    if db.query(models.SlcRadRoster.id).filter_by(mac_addr=form.d.mac_addr).count() > 0:
        return render("sys_roster_form", form=form, msg=u"MAC地址已经存在")
    roster = models.SlcRadRoster()
    roster.mac_addr = form.d.mac_addr.replace("-", ":").upper()
    roster.begin_time = form.d.begin_time
    roster.end_time = form.d.end_time
    roster.roster_type = form.d.roster_type
    db.add(roster)

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)新增黑白名单信息:%s' % (get_cookie("username"), roster.mac_addr)
    db.add(ops_log)

    db.commit()
    websock.update_cache("roster", mac_addr=roster.mac_addr)
    redirect("/roster")

@app.get('/update', apply=auth_opr)
def roster_update(db, render):
    roster_id = request.params.get("roster_id")
    form = roster_forms.roster_update_form()
    form.fill(db.query(models.SlcRadRoster).get(roster_id))
    return render("sys_roster_form", form=form)


@app.post('/update', apply=auth_opr)
def roster_add_update(db, render):
    form = roster_forms.roster_update_form()
    if not form.validates(source=request.forms):
        return render("sys_roster_form", form=form)
    roster = db.query(models.SlcRadRoster).get(form.d.id)
    roster.begin_time = form.d.begin_time
    roster.end_time = form.d.end_time
    roster.roster_type = form.d.roster_type

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)修改黑白名单信息:%s' % (get_cookie("username"), roster.mac_addr)
    db.add(ops_log)

    db.commit()
    websock.update_cache("roster", mac_addr=roster.mac_addr)
    redirect("/roster")

@app.get('/delete', apply=auth_opr)
def roster_delete(db, render):
    roster_id = request.params.get("roster_id")
    mac_addr = db.query(models.SlcRadRoster).get(roster_id).mac_addr
    db.query(models.SlcRadRoster).filter_by(id=roster_id).delete()

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)删除黑白名单信息:%s' % (get_cookie("username"), roster_id)
    db.add(ops_log)

    db.commit()
    websock.update_cache("roster", mac_addr=mac_addr)
    redirect("/roster")


permit.add_route("/roster", u"黑白名单管理", u"系统管理", is_menu=True, order=6)
permit.add_route("/roster/add", u"新增黑白名单", u"系统管理", order=6.01)
permit.add_route("/roster/update", u"修改白黑名单", u"系统管理", order=6.02)
permit.add_route("/roster/delete", u"删除黑白名单", u"系统管理", order=6.03)
