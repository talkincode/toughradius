#!/usr/bin/env python
#coding:utf-8

from bottle import Bottle
from bottle import request
from bottle import response
from bottle import redirect
from bottle import static_file
from tablib import Dataset
from toughradius.console import models
from toughradius.console.admin import wlan_forms as forms
from toughradius.console.libs import utils
from toughradius.console.base import *
from sqlalchemy import func
import datetime
import bottle


__prefix__ = "/wlan"

app = Bottle()
app.config['__prefix__'] = __prefix__
render = functools.partial(Render.render_app,app)

###############################################################################
# param config      
############################################################################### 

@app.get('/param',apply=auth_opr)
def param(db):   
    form = forms.param_form()
    fparam = {}
    for p in db.query(models.SlcWlanParam):
        fparam[p.param_name] = p.param_value
    form.fill(fparam)
    return render("wlan_param",form=form)

@app.post('/param/update',apply=auth_opr)
def param_update(db): 
    params = db.query(models.SlcWlanParam)
    for param_name in request.forms:
        if 'submit' in  param_name:
            continue
        param = db.query(models.SlcWlanParam).filter_by(param_name=param_name).first()
        if not param:
            param = models.SlcWlanParam()
            param.param_name = param_name
            param.param_value = request.forms.get(param_name)
            db.add(param)
        else:
            param.param_value = request.forms.get(param_name)
                
    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)修改Wlan参数'%(get_cookie("username"))
    db.add(ops_log)
    db.commit()
    
    redirect("/wlan/param")
    
permit.add_route("/wlan/param",u"Wlan参数管理",u"Wlan管理",is_menu=True,order=0)
permit.add_route("/wlan/param/update",u"Wlan参数修改",u"Wlan管理",is_menu=False,order=0.01,is_open=False)


