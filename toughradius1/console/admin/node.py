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
from toughradius.console.admin import node_forms
from hashlib import md5
from twisted.python import log
import bottle
import datetime
import json
import functools

__prefix__ = "/node"

app = Bottle()
app.config['__prefix__'] = __prefix__

###############################################################################
# node manage
###############################################################################

@app.get('/', apply=auth_opr)
def node(db, render):
    return render("sys_node_list", page_data=get_page_data(db.query(models.SlcNode)))


permit.add_route("/node", u"区域信息管理", u"系统管理", is_menu=True, order=1)


@app.get('/add', apply=auth_opr)
def node_add(db, render):
    return render("base_form", form=node_forms.node_add_form())


@app.post('/add', apply=auth_opr)
def node_add_post(db, render):
    form = node_forms.node_add_form()
    if not form.validates(source=request.forms):
        return render("base_form", form=form)
    node = models.SlcNode()
    node.node_name = form.d.node_name
    node.node_desc = form.d.node_desc
    db.add(node)

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)新增区域信息:%s' % (get_cookie("username"), node.node_name)
    db.add(ops_log)

    db.commit()
    redirect("/node")


permit.add_route("/node/add", u"新增区域", u"系统管理", order=1.01, is_open=False)


@app.get('/update', apply=auth_opr)
def node_update(db, render):
    node_id = request.params.get("node_id")
    form = node_forms.node_update_form()
    form.fill(db.query(models.SlcNode).get(node_id))
    return render("base_form", form=form)


@app.post('/update', apply=auth_opr)
def node_add_update(db, render):
    form = node_forms.node_update_form()
    if not form.validates(source=request.forms):
        return render("base_form", form=form)
    node = db.query(models.SlcNode).get(form.d.id)
    node.node_name = form.d.node_name
    node.node_desc = form.d.node_desc

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)修改区域信息:%s' % (get_cookie("username"), node.node_name)
    db.add(ops_log)

    db.commit()
    redirect("/node")


permit.add_route("/node/update", u"修改区域", u"系统管理", order=1.02, is_open=False)


@app.get('/delete', apply=auth_opr)
def node_delete(db, render):
    node_id = request.params.get("node_id")
    if db.query(models.SlcMember.member_id).filter_by(node_id=node_id).count() > 0:
        return render("error", msg=u"该节点下有用户，不允许删除")
    db.query(models.SlcNode).filter_by(id=node_id).delete()

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)删除区域信息:%s' % (get_cookie("username"), node_id)
    db.add(ops_log)

    db.commit()
    redirect("/node")


permit.add_route("/node/delete", u"删除区域", u"系统管理", order=1.03, is_open=False)