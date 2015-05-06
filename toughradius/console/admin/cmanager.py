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
from toughradius.console.libs.mpsapi import mpsapi
from toughradius.console.base import *
from toughradius.console.admin import cmanager_forms
import bottle
import datetime
from sqlalchemy import func

__prefix__ = "/cmanager"

app = Bottle()
app.config['__prefix__'] = __prefix__
render = functools.partial(Render.render_app,app)

###############################################################################
# customer manager manage        
###############################################################################

@app.get('/list',apply=auth_opr)
def cmanager_list(db):   
    manager_code = request.params.get('manager_code')
    _query = db.query(models.SlcCustomerManager)
    if manager_code:
        _query = _query.filter_by(manager_code=manager_code)

    return render("sys_cmanager_list", 
        page_data = get_page_data(_query),
        **request.params)

permit.add_route("%s/list"%__prefix__,u"客户经理管理",u"系统管理",is_menu=True,order=3.1)

@app.get('/add',apply=auth_opr)
def cmanager_add(db):   
    form = cmanager_forms.cmanage_add_form()
    return render("base_form",form=form)

@app.post('/add',apply=auth_opr)
def cmanager_add_post(db): 
    form=cmanager_forms.cmanage_add_form()
    if not form.validates(source=request.forms):
        return render("base_form", form=form)    

    if db.query(models.SlcCustomerManager).filter(
            models.SlcCustomerManager.manager_code==form.d.manager_code).count()>0:
        return render("base_form",form=form,msg=u"工号重复")
    
    if db.query(models.SlcCustomerManager).filter(
            models.SlcCustomerManager.active_code==form.d.active_code).count()>0:
        return render("base_form",form=form,msg=u"激活码重复")


    cmanager = models.SlcCustomerManager()
    cmanager.manager_code = form.d.manager_code
    cmanager.manager_name = form.d.manager_name
    cmanager.manager_mobile = form.d.manager_mobile
    cmanager.manager_email = form.d.manager_email
    cmanager.active_code = form.d.active_code
    cmanager.active_status = 0
    cmanager.create_time = utils.get_currtime()

    db.add(cmanager)
    db.commit()
    redirect("/cmanager/list")


permit.add_route("%s/add"%__prefix__,u"客户经理创建",u"系统管理",is_menu=False,order=3.11)

@app.get('/update',apply=auth_opr)
def cmanager_update(db):   
    manager_id = request.params.get("manager_id")
    cmanager = db.query(models.SlcCustomerManager).get(manager_id)
    form = cmanager_forms.cmanage_update_form()
    form.fill(cmanager)
    return render("base_form",form=form)

@app.post('/update',apply=auth_opr)
def cmanager_update_post(db): 
    form=cmanager_forms.cmanage_update_form()
    if not form.validates(source=request.forms):
        return render("base_form", form=form)    

    cmanager = db.query(models.SlcCustomerManager).get(form.d.id)
    cmanager.manager_name = form.d.manager_name
    cmanager.manager_mobile = form.d.manager_mobile
    cmanager.manager_email = form.d.manager_email

    db.commit()
    redirect("/cmanager/list")


permit.add_route("%s/update"%__prefix__,u"客户经理修改",u"系统管理",is_menu=False,order=3.12)


@app.get('/delete',apply=auth_opr)
def cmanager_delete(db):   
    manager_id = request.params.get("manager_id")
    manager = db.query(models.SlcCustomerManager).get(manager_id)
    if db.query(models.SlcOperator.id).filter_by(manager_code=manager.manager_code).count()>0:
        return render("error", msg=u"客户经理已经被关联到操作员，不允许删除")  

    db.query(models.SlcCustomerManager).filter_by(id=manager_id).delete()
    db.commit()
    redirect("/cmanager/list")

permit.add_route("%s/delete"%__prefix__,u"客户经理删除",u"系统管理",is_menu=False,order=3.13)

def get_mp_nickname(openid):
    pass

@app.get('/detail',apply=auth_opr)
def cmanager_detail(db):   
    manager_id = request.params.get("manager_id")
    cmanager = db.query(models.SlcCustomerManager).get(manager_id)
    return render("sys_cmanager_detail",cmanager=cmanager,get_mp_nickname=get_mp_nickname)

permit.add_route("%s/detail"%__prefix__,u"客户经理删除",u"系统管理",is_menu=False,order=3.14)

@app.post("/qrcode/new")
def post(db):
    manager_code = request.params.get("manager_code")
    is_wlan = request.params.get("is_wlan",None)

    cmanager = db.query(models.SlcCustomerManager).filter_by(
        manager_code = manager_code
    ).first()

    if not cmanager:
        return dict(code=1,msg=u'客户经理不存在')
         
    scene_str = is_wlan=='true' and  'wlan_cmqr_%s'%manager_code or 'cmqr_%s'%manager_code
    resp = mpsapi.create_limit_qrcode(scene_str)
    if 'errcode' in resp > 0:
        return resp

    qrcode_url = "https://mp.weixin.qq.com/cgi-bin/showqrcode?ticket=%s"%resp["ticket"]

    if is_wlan=='true':
        cmanager.manager_wlan_qrcode = qrcode_url
    else:
        cmanager.manager_qrcode = qrcode_url

    db.commit()
    return dict(code=0,msg="ok")

permit.add_route("%s/qrcode/new"%__prefix__,u"创建客户经理二维码",u"系统管理",is_menu=False,order=3.15)



