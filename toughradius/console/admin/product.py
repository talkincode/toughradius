#!/usr/bin/env python
# coding:utf-8

from bottle import Bottle
from bottle import request
from bottle import response
from bottle import redirect
from bottle import static_file
from tablib import Dataset
from toughradius.console import models
from toughradius.console.admin import product_forms
from toughradius.console.websock import websock
from toughradius.console.libs import utils
from toughradius.console.libs.radius_attrs import radius_attrs
from toughradius.console.base import *
from sqlalchemy import func
import datetime
import bottle

__prefix__ = "/product"

app = Bottle()
app.config['__prefix__'] = __prefix__

###############################################################################
# product manage
###############################################################################


@app.get('/json', apply=auth_opr)
def product_get(db,render):
    node_id = request.params.get('node_id')
    if not node_id: return dict(code=1, data=[])
    items = db.query(models.SlcRadProduct).filter_by(node_id=node_id)
    return dict(
        code=0,
        data=[{'code': it.id, 'name': it.product_name} for it in items]
    )


@app.get('/policy/get', apply=auth_opr)
def product_policy_get(db,render):
    product_id = request.params.get("product_id")
    product_policy = db.query(
        models.SlcRadProduct.product_policy
    ).filter_by(id=product_id).scalar()
    return dict(
        code=0,
        data={'id': product_id, 'policy': product_policy}
    )

@app.route('/', apply=auth_opr, method=['GET', 'POST'])
def product(db, render):
    _query = get_opr_products(db)
    return render(
        "sys_product_list",
        product_policys=product_forms.product_policy,
        node_list=db.query(models.SlcNode),
        page_data=get_page_data(_query)
    )


@app.get('/detail', apply=auth_opr)
def product_detail(db, render):
    product_id = request.params.get("product_id")
    product = db.query(models.SlcRadProduct).get(product_id)
    if not product:
        return render("error", msg=u"资费不存在")
    product_attrs = db.query(models.SlcRadProductAttr).filter_by(
        product_id=product_id)
    return render("sys_product_detail",
                  product_policys=product_forms.product_policy,
                  product=product, product_attrs=product_attrs)

@app.get('/add', apply=auth_opr)
def product_add(db, render):
    return render("sys_product_form", form=product_forms.product_add_form())

@app.post('/add', apply=auth_opr)
def product_add_post(db, render):
    form = product_forms.product_add_form()
    if not form.validates(source=request.forms):
        return render("sys_product_form", form=form)
    product = models.SlcRadProduct()
    product.product_name = form.d.product_name
    product.product_policy = form.d.product_policy
    product.product_status = form.d.product_status
    product.fee_months = int(form.d.get("fee_months", 0))
    product.fee_times = utils.hour2sec(form.d.get("fee_times", 0))
    product.fee_flows = utils.mb2kb(form.d.get("fee_flows", 0))
    product.bind_mac = form.d.bind_mac
    product.bind_vlan = form.d.bind_vlan
    product.concur_number = form.d.concur_number
    product.fee_period = form.d.fee_period
    product.fee_price = utils.yuan2fen(form.d.fee_price)
    product.input_max_limit = utils.mbps2bps(form.d.input_max_limit)
    product.output_max_limit = utils.mbps2bps(form.d.output_max_limit)
    _datetime = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    product.create_time = _datetime
    product.update_time = _datetime
    db.add(product)

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)新增资费信息:%s' % (
        get_cookie("username"), product.product_name)
    db.add(ops_log)

    db.commit()
    redirect(__prefix__)

@app.get('/update', apply=auth_opr)
def product_update(db, render):
    product_id = request.params.get("product_id")
    form = product_forms.product_update_form()
    product = db.query(models.SlcRadProduct).get(product_id)
    form.fill(product)
    form.product_policy_name.set_value(
        product_forms.product_policy[product.product_policy])
    form.fee_times.set_value(utils.sec2hour(product.fee_times))
    form.fee_flows.set_value(utils.kb2mb(product.fee_flows))
    form.input_max_limit.set_value(utils.bps2mbps(product.input_max_limit))
    form.output_max_limit.set_value(utils.bps2mbps(product.output_max_limit))
    form.fee_price.set_value(utils.fen2yuan(product.fee_price))
    return render("sys_product_form", form=form)


@app.post('/update', apply=auth_opr)
def product_update(db, render):
    form = product_forms.product_update_form()
    if not form.validates(source=request.forms):
        return render("sys_product_form", form=form)
    product = db.query(models.SlcRadProduct).get(form.d.id)
    product.product_name = form.d.product_name
    product.product_status = form.d.product_status
    product.fee_months = int(form.d.get("fee_months", 0))
    product.fee_times = utils.hour2sec(form.d.get("fee_times", 0))
    product.fee_flows = utils.mb2kb(form.d.get("fee_flows", 0))
    product.bind_mac = form.d.bind_mac
    product.bind_vlan = form.d.bind_vlan
    product.concur_number = form.d.concur_number
    product.fee_period = form.d.fee_period
    product.fee_price = utils.yuan2fen(form.d.fee_price)
    product.input_max_limit = utils.mbps2bps(form.d.input_max_limit)
    product.output_max_limit = utils.mbps2bps(form.d.output_max_limit)
    product.update_time = utils.get_currtime()

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)修改资费信息:%s' % (
        get_cookie("username"), product.product_name)
    db.add(ops_log)

    db.commit()
    websock.update_cache("product", product_id=product.id)
    redirect(__prefix__)

@app.get('/delete', apply=auth_opr)
def product_delete(db, render):
    product_id = request.params.get("product_id")
    if db.query(models.SlcRadAccount).filter_by(product_id=product_id).count() > 0:
        return render("error", msg=u"该套餐有用户使用，不允许删除")
    if db.query(models.SlcRechargerCard).filter_by(product_id=product_id).count() > 0:
        return render("error", msg=u"该套餐有发行充值卡，不允许删除")

    db.query(models.SlcRadProduct).filter_by(id=product_id).delete()

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)删除资费信息:%s' % (
        get_cookie("username"), product_id)
    db.add(ops_log)

    db.commit()
    websock.update_cache("product", product_id=product_id)
    redirect(__prefix__)

@app.get('/attr/add', apply=auth_opr)
def product_attr_add(db, render):
    product_id = request.params.get("product_id")
    if db.query(models.SlcRadProduct).filter_by(id=product_id).count() <= 0:
        return render("error", msg=u"资费不存在")
    form = product_forms.product_attr_add_form()
    form.product_id.set_value(product_id)
    return render("sys_pattr_form", form=form, pattrs=radius_attrs)

@app.post('/attr/add', apply=auth_opr)
def product_attr_add(db, render):
    form = product_forms.product_attr_add_form()
    if not form.validates(source=request.forms):
        return render("sys_pattr_form", form=form, pattrs=radius_attrs)
    attr = models.SlcRadProductAttr()
    attr.product_id = form.d.product_id
    attr.attr_name = form.d.attr_name
    attr.attr_value = form.d.attr_value
    attr.attr_desc = form.d.attr_desc
    db.add(attr)

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)新增资费属性信息:%s' % (
        get_cookie("username"), attr.attr_name)
    db.add(ops_log)

    db.commit()

    redirect("%s/detail?product_id=%s" % (__prefix__, form.d.product_id))

@app.get('/attr/update', apply=auth_opr)
def product_attr_update(db, render):
    attr_id = request.params.get("attr_id")
    attr = db.query(models.SlcRadProductAttr).get(attr_id)
    form = product_forms.product_attr_update_form()
    form.fill(attr)
    return render("sys_pattr_form", form=form, pattrs=radius_attrs)

@app.post('/attr/update', apply=auth_opr)
def product_attr_update(db, render):
    form = product_forms.product_attr_update_form()
    if not form.validates(source=request.forms):
        return render("pattr_form", form=form, pattrs=radius_attrs)
    attr = db.query(models.SlcRadProductAttr).get(form.d.id)
    attr.attr_name = form.d.attr_name
    attr.attr_value = form.d.attr_value
    attr.attr_desc = form.d.attr_desc

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)修改资费属性信息:%s' % (
        get_cookie("username"), attr.attr_name)
    db.add(ops_log)

    db.commit()
    websock.update_cache("product", product_id=form.d.product_id)
    redirect("%s/detail?product_id=%s" % (__prefix__, form.d.product_id))

@app.get('/attr/delete', apply=auth_opr)
def product_attr_update(db, render):
    attr_id = request.params.get("attr_id")
    attr = db.query(models.SlcRadProductAttr).get(attr_id)
    product_id = attr.product_id
    db.query(models.SlcRadProductAttr).filter_by(id=attr_id).delete()

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)删除资费属性信息:%s' % (
        get_cookie("username"), serial_json(attr))
    db.add(ops_log)

    db.commit()
    websock.update_cache("product", product_id=product_id)
    redirect("%s/detail?product_id=%s" % (__prefix__, product_id))


permit.add_route(__prefix__, u"资费信息管理", u"系统管理", is_menu=True, order=4)
permit.add_route("%s/add" % __prefix__, u"新增资费", u"系统管理", order=4.02)
permit.add_route("%s/update" % __prefix__, u"修改资费", u"系统管理", order=4.03)
permit.add_route("%s/delete" % __prefix__, u"删除资费", u"系统管理", order=4.04)
permit.add_route("%s/detail" % __prefix__, u"资费详情查看", u"系统管理", order=4.01)
permit.add_route("%s/attr/add" % __prefix__, u"新增资费扩展属性", u"系统管理", order=4.05)
permit.add_route("%s/attr/update" %__prefix__, u"修改资费扩展属性", u"系统管理", order=4.06)
permit.add_route("%s/attr/delete" %__prefix__, u"删除资费扩展属性", u"系统管理", order=4.07)
