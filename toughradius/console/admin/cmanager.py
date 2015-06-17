#!/usr/bin/env python
#coding:utf-8

from bottle import Bottle
from bottle import request
from bottle import response
from bottle import redirect
from bottle import static_file
from bottle import mako_template as render
from tablib import Dataset
from toughradius.console import models
from toughradius.console.libs import utils
from toughradius.console.base import *
from toughradius.console.admin import cmanager_forms
import bottle
import datetime
from sqlalchemy import func

__prefix__ = "/cmanager"

app = Bottle()
app.config['__prefix__'] = __prefix__

###############################################################################
# customer manager manage        
###############################################################################

@app.get('/list',apply=auth_opr)
def cmanager_list(db, render):
    manager_code = request.params.get('manager_code')
    _query = db.query(models.SlcCustomerManager)
    if manager_code:
        _query = _query.filter_by(manager_code=manager_code)

    return render("sys_cmanager_list", 
        page_data = get_page_data(_query),
        **request.params)

permit.add_route("%s/list"%__prefix__,u"客户经理管理",u"系统管理",is_menu=True,order=3.1)

def get_oprs(db,copr_name=None):
    mers = [mer[0] for mer in db.query(models.SlcCustomerManager.operator_name) if mer[0]]
    oprs = [(opr.operator_name,opr.operator_desc) \
        for opr in db.query(models.SlcOperator).filter_by(operator_status=0) \
        if opr.operator_name not in mers] 
    oprs = [opr_sel for opr_sel in oprs if opr_sel[0] not in 'admin']
    if copr_name:
        oprs.insert(0,(copr_name,copr_name))
    return oprs

@app.get('/add',apply=auth_opr)
def cmanager_add(db, render):
    form = cmanager_forms.cmanage_add_form(oprs=get_oprs(db))
    return render("base_form",form=form)

@app.post('/add',apply=auth_opr)
def cmanager_add_post(db, render):
    form=cmanager_forms.cmanage_add_form(oprs=get_oprs(db))
    if not form.validates(source=request.forms):
        return render("base_form", form=form)    

    if db.query(models.SlcCustomerManager).filter(
            models.SlcCustomerManager.manager_code==form.d.manager_code).count()>0:
        return render("base_form",form=form,msg=u"工号重复")
    
    cmanager = models.SlcCustomerManager()
    cmanager.manager_code = form.d.manager_code
    cmanager.manager_name = form.d.manager_name
    cmanager.manager_mobile = form.d.manager_mobile
    cmanager.manager_email = form.d.manager_email
    cmanager.operator_name = form.d.operator_name
    cmanager.create_time = utils.get_currtime()

    db.add(cmanager)
    db.commit()
    redirect("/cmanager/list")


permit.add_route("%s/add"%__prefix__,u"客户经理创建",u"系统管理",is_menu=False,order=3.11)

@app.get('/update',apply=auth_opr)
def cmanager_update(db, render):
    manager_id = request.params.get("manager_id")
    cmanager = db.query(models.SlcCustomerManager).get(manager_id)
    form = cmanager_forms.cmanage_update_form(
        oprs=get_oprs(db,copr_name=cmanager.operator_name))
    form.fill(cmanager)
    return render("base_form",form=form)

@app.post('/update',apply=auth_opr)
def cmanager_update_post(db, render):
    form=cmanager_forms.cmanage_update_form(oprs=get_oprs(db))
    if not form.validates(source=request.forms):
        form=cmanager_forms.cmanage_update_form(
            oprs=get_oprs(db,copr_name=form.d.operator_name))
        return render("base_form", form=form)    

    cmanager = db.query(models.SlcCustomerManager).get(form.d.id)
    cmanager.manager_name = form.d.manager_name
    cmanager.manager_mobile = form.d.manager_mobile
    cmanager.manager_email = form.d.manager_email
    cmanager.operator_name = form.d.operator_name

    db.commit()
    redirect("/cmanager/list")


permit.add_route("%s/update"%__prefix__,u"客户经理修改",u"系统管理",is_menu=False,order=3.12)


@app.get('/delete',apply=auth_opr)
def cmanager_delete(db, render):
    manager_id = request.params.get("manager_id")
    db.query(models.SlcCustomerManager).filter_by(id=manager_id).delete()
    db.commit()
    redirect("/cmanager/list")

permit.add_route("%s/delete"%__prefix__,u"客户经理删除",u"系统管理",is_menu=False,order=3.13)


@app.get('/detail',apply=auth_opr)
def cmanager_detail(db, render):
    manager_id = request.params.get("manager_id")
    cmanager = db.query(models.SlcCustomerManager).get(manager_id)
    return render("sys_cmanager_detail",cmanager=cmanager)

permit.add_route("%s/detail"%__prefix__,u"客户经理删除",u"系统管理",is_menu=False,order=3.14)





