#!/usr/bin/env python
#coding:utf-8
import sys,os
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

app = Bottle()

##############################################################################
# test handle
##############################################################################
@app.route('/test',apply=auth_opr)
def test(db, render):
    form = forms.param_form()
    fparam = {}
    for p in db.query(models.SlcParam):
        fparam[p.param_name] = p.param_value
    form.fill(fparam)
    return render("base_form",form=form)

@app.get('/test/pid',apply=auth_opr)
def product_id(db, render):
    name = request.params.get("name")   
    product = db.query(models.SlcRadProduct).filter(
        models.SlcRadProduct.product_name == name
    ).first()
    return dict(pid=product.id)
    
@app.get('/test/mid',apply=auth_opr)
def member_id(db, render):
    name = request.params.get("name")   
    member = db.query(models.SlcMember).filter(
        models.SlcMember.member_name == name
    ).first()
    return dict(mid=member.member_id)
    
@app.route('/mksign',apply=auth_opr)
def mksign(db, render):
    sign_args = request.params.get('sign_args')
    return dict(code=0,sign=utils.mk_sign(sign_args.strip().split(',')))
    
@app.post('/encrypt',apply=auth_opr)
def encrypt_data(db, render):
    msg_data = request.params.get('data')
    return dict(code=0,data=utils.encrypt(msg_data))
    
@app.post('/decrypt',apply=auth_opr)
def decrypt_data(db, render):
    msg_data = request.params.get('data')
    return dict(code=0,data=utils.decrypt(msg_data))
    


###############################################################################
# Basic handle         
###############################################################################

@app.route('/',apply=auth_opr)
def index(db, render):
    online_count = db.query(models.SlcRadOnline.id).count()
    user_total = db.query(models.SlcRadAccount.account_number).filter_by(status=1).count()
    return render("index",**locals())

@app.route('/static/:path#.+#')
def route_static(path,render):
    static_path = os.path.join(os.path.split(os.path.split(__file__)[0])[0],'static')
    return static_file(path, root=static_path)
    
###############################################################################
# update all cache      
###############################################################################    
@app.get('/cache/clean')
def clear_cache(db,render):
    def cbk(resp):
        print 'cbk',resp
    bottle.TEMPLATES.clear()
    for _cache in cache_managers.values():
        _cache.clear()
    websock.update_cache("all",callback=cbk)
    return dict(code=0,msg=u"已刷新缓存")
    
###############################################################################
# login handle         
###############################################################################

@app.get('/login')
def admin_login_get(db, render):
    return render("login")

@app.post('/login')
def admin_login_post(db, render):
    uname = request.forms.get("username")
    upass = request.forms.get("password")
    if not uname:return dict(code=1,msg=u"请填写用户名")
    if not upass:return dict(code=1,msg=u"请填写密码")
    enpasswd = md5(upass.encode()).hexdigest()
    opr = db.query(models.SlcOperator).filter_by(
        operator_name=uname,
        operator_pass=enpasswd
    ).first()
    if not opr:return dict(code=1,msg=u"用户名密码不符")
    if opr.operator_status == 1:return dict(code=1,msg=u"该操作员账号已被停用")
    set_cookie('username',uname)
    set_cookie('opr_type',opr.operator_type)
    set_cookie('login_time', utils.get_currtime())
    set_cookie('login_ip', request.remote_addr)  
    
    if opr.operator_type > 0:
        permit.unbind_opr(uname)
        for rule in db.query(models.SlcOperatorRule).filter_by(operator_name=uname):
            permit.bind_opr(rule.operator_name,rule.rule_path)  

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = uname
    ops_log.operate_ip = request.remote_addr
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)登陆'%(uname,)
    db.add(ops_log)
    db.commit()

    return dict(code=0,msg="ok")

@app.get("/logout")
def admin_logout(db, render):

    if get_cookie("username") and get_cookie("username"):
        ops_log = models.SlcRadOperateLog()
        ops_log.operator_name = get_cookie("username")
        ops_log.operate_ip = get_cookie("login_ip")
        ops_log.operate_time = utils.get_currtime()
        ops_log.operate_desc = u'操作员(%s)登出'%(get_cookie("username"),)
        db.add(ops_log)
        db.commit()
        if get_cookie('opt_type') > 0:
            permit.unbind_opr(get_cookie("username"))
    set_cookie('username',None)
    set_cookie('login_time', None)
    set_cookie('opr_type',None)
    set_cookie('login_ip', None)   
    request.cookies.clear()
    redirect('/login')


@app.route('/dashboard', apply=auth_opr)
def index(db, render):
    online_count = db.query(models.SlcRadOnline.id).count()
    user_total = db.query(models.SlcRadAccount.account_number).filter_by(status=1).count()
    return render("index", **locals())


@app.route('/control', apply=auth_opr)
def index(db, render):
    host = app.config['control.host']
    port = app.config['control.port']
    redirect('//%s:%s'%(host,port))

permit.add_route("/dashboard", u"系统控制面板", u"系统管理", order=0, is_menu=True,is_open=True)
permit.add_route("/control", u"系统管理", u"系统管理", order=0, is_menu=False, is_open=False)