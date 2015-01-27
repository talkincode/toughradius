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
from base import (
    logger,set_cookie,get_cookie,cache,
    auth_cus,get_member_by_name,get_page_data
)
from libs import utils
from sqlalchemy.sql import exists
import bottle
import models
import forms
import decimal
import datetime
import functools

app = Bottle()

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
 
    set_cookie('customer_id',member.member_id)
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
     
@app.get('/product/list',apply=auth_cus)
def product_list(db):
    return render("product_list",products=db.query(models.SlcRadProduct).filter_by(
        product_status = 0
    ))
    
@app.route('/billing',apply=auth_cus,method=['GET','POST'])
def billing_query(db): 
    account_number = request.params.get('account_number')  
    query_begin_time = request.params.get('query_begin_time')  
    query_end_time = request.params.get('query_end_time')  
    _query = db.query(
        models.SlcRadBilling,
        models.SlcMember.node_id,
    ).filter(
        models.SlcRadBilling.account_number == models.SlcRadAccount.account_number,
        models.SlcMember.member_id == models.SlcRadAccount.member_id,
        models.SlcMember.member_id == get_cookie("customer_id")
    )
    if account_number:
        _query = _query.filter(models.SlcRadBilling.account_number.like('%'+account_number+'%'))
    if query_begin_time:
        _query = _query.filter(models.SlcRadBilling.create_time >= query_begin_time)
    if query_end_time:
        _query = _query.filter(models.SlcRadBilling.create_time <= query_end_time)
    _query = _query.order_by(models.SlcRadBilling.create_time.desc())
    return render("billing_list", 
        accounts=db.query(models.SlcRadAccount).filter_by(member_id=get_cookie("customer_id")),
        page_data=get_page_data(_query),**request.params)
        

###############################################################################
# ticket query        
###############################################################################

@app.route('/ticket',apply=auth_cus,method=['GET','POST'])
def ticket_query(db): 
    account_number = request.params.get('account_number')  
    query_begin_time = request.params.get('query_begin_time')  
    query_end_time = request.params.get('query_end_time')  
    _query = db.query(
        models.SlcRadTicket.id,
        models.SlcRadTicket.account_number,
        models.SlcRadTicket.nas_addr,
        models.SlcRadTicket.acct_session_id,
        models.SlcRadTicket.acct_start_time,
        models.SlcRadTicket.acct_input_octets,
        models.SlcRadTicket.acct_output_octets,
        models.SlcRadTicket.acct_stop_time,
        models.SlcRadTicket.framed_ipaddr,
        models.SlcRadTicket.mac_addr,
        models.SlcRadTicket.nas_port_id,
        models.SlcMember.node_id,
        models.SlcMember.realname
    ).filter(
        models.SlcRadTicket.account_number == models.SlcRadAccount.account_number,
        models.SlcMember.member_id == models.SlcRadAccount.member_id,
        models.SlcMember.member_id == get_cookie("customer_id")
    )
    if account_number:
        _query = _query.filter(models.SlcRadTicket.account_number == account_number)
    if query_begin_time:
        _query = _query.filter(models.SlcRadTicket.acct_start_time >= query_begin_time)
    if query_end_time:
        _query = _query.filter(models.SlcRadTicket.acct_stop_time <= query_end_time)

    _query = _query.order_by(models.SlcRadTicket.acct_start_time.desc())
    return render("ticket_list", 
        accounts=db.query(models.SlcRadAccount).filter_by(member_id=get_cookie("customer_id")),
        page_data = get_page_data(_query),
        **request.params)    
    
@app.get('/password/update',apply=auth_cus)    
def password_update_get(db):
    form = forms.password_update_form()
    account_number = request.params.get('account_number')  
    account = db.query(models.SlcRadAccount).get(account_number)
    if not account:
        return render("base_form", form=form,msg=u'没有这个账号')
        
    if account.member_id != get_cookie("customer_id"):
        return render("base_form", form=form,msg=u'该账号用用户不匹配')  
          
    form.account_number.set_value(account_number)
    return render("base_form",form=form)
    
@app.post('/password/update',apply=auth_cus)    
def password_update_post(db):
    form = forms.password_update_form()
    if not form.validates(source=request.forms):
        return render("base_form", form=form)
        
    account = db.query(models.SlcRadAccount).get(form.d.account_number)
    if not account:
        return render("base_form", form=form,msg=u'没有这个账号')
        
    if account.member_id != get_cookie("customer_id"):
        return render("base_form", form=form,msg=u'该账号用用户不匹配')
    
    if account.password !=  form.d.old_password:
        return render("base_form", form=form,msg=u'旧密码不正确')
        
    if form.d.new_password != form.d.new_password2:
        return render("base_form", form=form,msg=u'确认新密码不匹配')
    
    account.password =  utils.encrypt(form.d.new_password)
    db.commit()
    redirect("/")