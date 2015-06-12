#!/usr/bin/env python
#coding=utf-8

# !/usr/bin/env python
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
from toughradius.console.admin import opr_forms
from hashlib import md5
from twisted.python import log
import bottle
import datetime
import json
import functools

__prefix__ = "/opr"

app = Bottle()
app.config['__prefix__'] = __prefix__


###############################################################################
# opr manage
###############################################################################

@app.route('/', apply=auth_opr, method=['GET', 'POST'])
def opr(db, render):
    return render("sys_opr_list",
                  oprtype=opr_forms.opr_type,
                  oprstatus=opr_forms.opr_status_dict,
                  opr_list=db.query(models.SlcOperator))


permit.add_route("/opr", u"操作员管理", u"系统管理", is_menu=True, order=3, is_open=False)


@app.get('/add', apply=auth_opr)
def opr_add(db, render):
    nodes = [(n.node_name, n.node_desc) for n in db.query(models.SlcNode)]
    products = [(p.id,p.product_name) for p in db.query(models.SlcRadProduct)  ]
    form = opr_forms.opr_add_form(nodes, products)
    return render("sys_opr_form", form=form, rules=[])


@app.post('/add', apply=auth_opr)
def opr_add_post(db, render):
    nodes = [(n.node_name, n.node_desc) for n in db.query(models.SlcNode)]
    products = [(p.id, p.product_name) for p in db.query(models.SlcRadProduct)]
    form = opr_forms.opr_add_form(nodes,products)
    if not form.validates(source=request.forms):
        return render("sys_opr_form", form=form, rules=[])
    if db.query(models.SlcOperator.id).filter_by(operator_name=form.d.operator_name).count() > 0:
        return render("sys_opr_form", form=form, rules=[], msg=u"操作员已经存在")

    opr = models.SlcOperator()
    opr.operator_name = form.d.operator_name
    opr.operator_type = 1
    opr.operator_pass = md5(form.d.operator_pass).hexdigest()
    opr.operator_desc = form.d.operator_desc
    opr.operator_status = form.d.operator_status
    db.add(opr)

    for node in request.params.getall("operator_nodes"):
        onode = models.SlcOperatorNodes()
        onode.operator_name = form.d.operator_name
        onode.node_name = node
        db.add(onode)

    for product_id in request.params.getall("operator_products"):
        oproduct = models.SlcOperatorProducts()
        oproduct.operator_name = form.d.operator_name
        oproduct.product_id = product_id
        db.add(oproduct)


    for path in request.params.getall("rule_item"):
        item = permit.get_route(path)
        if not item: continue
        rule = models.SlcOperatorRule()
        rule.operator_name = opr.operator_name
        rule.rule_name = item['name']
        rule.rule_path = item['path']
        rule.rule_category = item['category']
        db.add(rule)

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)新增操作员信息:%s' % (get_cookie("username"), opr.operator_name)
    db.add(ops_log)

    db.commit()
    redirect("/opr")


permit.add_route("/opr/add", u"新增操作员", u"系统管理", order=3.01, is_open=False)


@app.get('/update', apply=auth_opr)
def opr_update(db, render):
    opr_id = request.params.get("opr_id")
    opr = db.query(models.SlcOperator).get(opr_id)
    nodes = [(n.node_name, n.node_desc) for n in db.query(models.SlcNode)]
    products = [(str(p.id), p.product_name) for p in db.query(models.SlcRadProduct)]

    form = opr_forms.opr_update_form(nodes, products)
    form.fill(opr)
    form.operator_pass.set_value('')

    onodes = db.query(models.SlcOperatorNodes).filter_by(operator_name=form.d.operator_name)
    oproducts = db.query(models.SlcOperatorProducts).filter_by(operator_name=form.d.operator_name)

    form.operator_nodes.set_value([ond.node_name for ond in onodes])
    form.operator_products.set_value([str(p.product_id) for p in oproducts])
    rules = db.query(models.SlcOperatorRule.rule_path).filter_by(operator_name=opr.operator_name)
    rules = [r[0] for r in rules]
    return render("sys_opr_form", form=form, rules=rules)


@app.post('/update', apply=auth_opr)
def opr_add_update(db, render):
    nodes = [(n.node_name, n.node_desc) for n in db.query(models.SlcNode)]
    form = opr_forms.opr_update_form(nodes)
    if not form.validates(source=request.forms):
        rules = db.query(models.SlcOperatorRule.rule_path).filter_by(operator_name=opr.operator_name)
        rules = [r[0] for r in rules]
        return render("sys_opr_form", form=form, rules=rules)
    opr = db.query(models.SlcOperator).get(form.d.id)

    if form.d.operator_pass:
        opr.operator_pass = md5(form.d.operator_pass).hexdigest()
    opr.operator_desc = form.d.operator_desc
    opr.operator_status = form.d.operator_status

    db.query(models.SlcOperatorNodes).filter_by(operator_name=opr.operator_name).delete()
    for node in request.params.getall("operator_nodes"):
        onode = models.SlcOperatorNodes()
        onode.operator_name = form.d.operator_name
        onode.node_name = node
        db.add(onode)

    db.query(models.SlcOperatorProducts).filter_by(operator_name=opr.operator_name).delete()
    for product_id in request.params.getall("operator_products"):
        oproduct = models.SlcOperatorProducts()
        oproduct.operator_name = form.d.operator_name
        oproduct.product_id = product_id
        db.add(oproduct)

    # update rules
    db.query(models.SlcOperatorRule).filter_by(operator_name=opr.operator_name).delete()

    for path in request.params.getall("rule_item"):
        item = permit.get_route(path)
        if not item: continue
        rule = models.SlcOperatorRule()
        rule.operator_name = opr.operator_name
        rule.rule_name = item['name']
        rule.rule_path = item['path']
        rule.rule_category = item['category']
        db.add(rule)

    permit.unbind_opr(opr.operator_name)
    for rule in db.query(models.SlcOperatorRule).filter_by(operator_name=opr.operator_name):
        permit.bind_opr(rule.operator_name, rule.rule_path)

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)修改操作员信息:%s' % (get_cookie("username"), opr.operator_name)
    db.add(ops_log)

    db.commit()
    redirect("/opr")


permit.add_route("/opr/update", u"修改操作员", u"系统管理", order=3.02, is_open=False)


@app.get('/delete', apply=auth_opr)
def opr_delete(db, render):
    opr_id = request.params.get("opr_id")
    opr = db.query(models.SlcOperator).get(opr_id)
    db.query(models.SlcOperatorNodes).filter_by(operator_name=opr.operator_name).delete()
    db.query(models.SlcOperatorRule).filter_by(operator_name=opr.operator_name).delete()
    db.query(models.SlcOperator).filter_by(id=opr_id).delete()

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)删除操作员信息:%s' % (get_cookie("username"), opr_id)
    db.add(ops_log)

    db.commit()
    redirect("/opr")


permit.add_route("/opr/delete", u"删除操作员", u"系统管理", order=3.03, is_open=False)

