#!/usr/bin/env python
#coding:utf-8

from bottle import Bottle
from bottle import request
from bottle import response
from bottle import redirect
from bottle import run as runserver
from bottle import static_file
from bottle import abort
from bottle import mako_template as render
from hashlib import md5
from tablib import Dataset
from libs import sqla_plugin 
from base import (logger,set_cookie,get_cookie,cache)
from libs import utils
from sqlalchemy.sql import exists
import bottle
import models
import forms
import decimal
import datetime
import functools

app = Bottle()

def auth_cus(func):
    @functools.wraps(func)
    def warp(*args,**kargs):
        if not get_cookie("customer"):
            log.msg("user login timeout")
            return redirect('/login')
        else:
            return func(*args,**kargs)
    return warp
    
###############################################################################
# Basic handle         
###############################################################################    
    
@app.error(404)
def error404(error):
    return render("error.html",msg=u"页面不存在 - 请联系管理员!")

@app.error(500)
def error500(error):
    return render("error.html",msg=u"出错了： %s"%error.exception)

@app.route('/static/:path#.+#')
def route_static(path):
    return static_file(path, root='./static')    

###############################################################################
# login handle         
###############################################################################

@app.get('/',apply=auth_cus)
def customer_index(db):
    @cache.cache('customer_index_get_data',expire=300)   
    def get_data(member_name):
        member = db.query(models.SlcMember).filter_by(member_name=member_name).first()
        accounts = db.query(
            models.SlcMember.realname,
            models.SlcRadAccount.member_id,
            models.SlcRadAccount.account_number,
            models.SlcRadAccount.expire_date,
            models.SlcRadAccount.balance,
            models.SlcRadAccount.time_length,
            models.SlcRadAccount.status,
            models.SlcRadAccount.last_pause,
            models.SlcRadAccount.create_time,
            models.SlcRadProduct.product_name,
            models.SlcRadProduct.product_policy
        ).filter(
            models.SlcRadProduct.id == models.SlcRadAccount.product_id,
            models.SlcMember.member_id == models.SlcRadAccount.member_id,
            models.SlcRadAccount.member_id == member.member_id
        )
        orders = db.query(
            models.SlcMemberOrder.order_id,
            models.SlcMemberOrder.order_id,
            models.SlcMemberOrder.product_id,
            models.SlcMemberOrder.account_number,
            models.SlcMemberOrder.order_fee,
            models.SlcMemberOrder.actual_fee,
            models.SlcMemberOrder.pay_status,
            models.SlcMemberOrder.create_time,
            models.SlcMemberOrder.order_desc,
            models.SlcRadProduct.product_name
        ).filter(
            models.SlcRadProduct.id == models.SlcMemberOrder.product_id,
            models.SlcMemberOrder.member_id==member.member_id
        )
        return member,accounts,orders
    member,accounts,orders = get_data(get_cookie('customer'))
    return  render("index",member=member,accounts=accounts,orders=orders)    


@app.get('/login')
def member_login_get(db):
    form = forms.member_login_form()
    form.next.set_value(request.params.get('next','/'))
    return render("login",form=form)

@app.post('/login')
def member_login_post(db):
    next = request.params.get("next", "/")
    form = forms.member_login_form()
    if not form.validates(source=request.params):
        return render("login", form=form)

    member = db.query(models.SlcMember).filter_by(
        member_name=form.d.username,
        password=md5(form.d.password.encode()).hexdigest()
    ).first()

    if not member:
        return render("login", form=form,msg=u"用户名密码不符合")
 
    set_cookie('customer',form.d.username)
    set_cookie('customer_login_time', utils.get_currtime())
    set_cookie('customer_login_ip', request.remote_addr) 
    redirect(next)

@app.get("/logout")
def member_logout():
    set_cookie('customer',None)
    set_cookie('customer_login_time', None)
    set_cookie('customer_login_ip', None)     
    request.cookies.clear()
    redirect('login')


@app.get('/join')
def member_join_get(db):
    nodes = [ (n.id,n.node_name) for n in db.query(models.SlcNode)]
    form = forms.member_join_form(nodes)
    return render("join",form=form)
    
@app.post('/join')
def member_join_post(db):
    nodes = [ (n.id,n.node_name) for n in db.query(models.SlcNode)]
    form = forms.member_join_form(nodes)
    if not form.validates(source=request.params):
        return render("join", form=form)    
        
    if db.query(exists().where(models.SlcMember.member_name == form.d.username)).scalar():
        return render("join",form=form,msg=u"用户{0}已被使用".format(form.d.username))
        
    if db.query(exists().where(models.SlcMember.email == form.d.email)).scalar():
        return render("join",form=form,msg=u"用户邮箱{0}已被使用".format(form.d.email))
    
    member = models.SlcMember()
    member.node_id = form.d.node_id
    member.realname = form.d.realname
    member.member_name = form.d.username
    member.password = md5(form.d.password.encode()).hexdigest()
    member.idcard = form.d.idcard
    member.sex = form.d.sex
    member.age = int(form.d.age)
    member.email = form.d.email
    member.mobile = form.d.mobile
    member.address = form.d.address
    member.create_time = utils.get_currtime()
    member.update_time = utils.get_currtime()
    db.add(member) 
    db.commit()
   
    logger.info(u"新用户注册成功,member_name=%s"%form.d.member_name)
    redirect('/login')
   
@app.get('/account/detail',apply=auth_cus)
def account_detail(db):
    account_number = request.params.get('account_number')  
    user  = db.query(
        models.SlcMember.realname,
        models.SlcRadAccount.member_id,
        models.SlcRadAccount.account_number,
        models.SlcRadAccount.expire_date,
        models.SlcRadAccount.balance,
        models.SlcRadAccount.time_length,
        models.SlcRadAccount.user_concur_number,
        models.SlcRadAccount.status,
        models.SlcRadAccount.mac_addr,
        models.SlcRadAccount.vlan_id,
        models.SlcRadAccount.vlan_id2,
        models.SlcRadAccount.ip_address,
        models.SlcRadAccount.bind_mac,
        models.SlcRadAccount.bind_vlan,
        models.SlcRadAccount.ip_address,
        models.SlcRadAccount.install_address,
        models.SlcRadAccount.create_time,
        models.SlcRadProduct.product_name
    ).filter(
            models.SlcRadProduct.id == models.SlcRadAccount.product_id,
            models.SlcMember.member_id == models.SlcRadAccount.member_id,
            models.SlcRadAccount.account_number == account_number
    ).first()
    if not user:
        return render("error",msg=u"账号不存在")
    return render("account_detail",user=user)
     