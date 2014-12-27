#!/usr/bin/env python
#coding:utf-8

from bottle import Bottle
from bottle import request
from bottle import response
from bottle import redirect
from bottle import static_file
from bottle import mako_template as render
from tablib import Dataset
import bottle
import models
import forms
import datetime
from base import *

app = Bottle()


###############################################################################
# user manage        
###############################################################################
                   
@app.route('/user',apply=auth_opr,method=['GET','POST'])
@app.get('/user/export',apply=auth_opr)
def user_query(db):   
    node_id = request.params.get('node_id')
    product_id = request.params.get('product_id')
    user_name = request.params.get('user_name')
    status = request.params.get('status')
    _query = db.query(
            models.SlcMember.realname,
            models.SlcRadAccount.member_id,
            models.SlcRadAccount.account_number,
            models.SlcRadAccount.expire_date,
            models.SlcRadAccount.balance,
            models.SlcRadAccount.time_length,
            models.SlcRadAccount.status,
            models.SlcRadAccount.create_time,
            models.SlcRadProduct.product_name
        ).filter(
            models.SlcRadProduct.id == models.SlcRadAccount.product_id,
            models.SlcMember.member_id == models.SlcRadAccount.member_id
        )
    if node_id:
        _query = _query.filter(models.SlcMember.node_id == node_id)
    if product_id:
        _query = _query.filter(models.SlcRadAccount.product_id==product_id)
    if user_name:
        _query = _query.filter(models.SlcRadAccount.account_number.like('%'+user_name+'%'))
    if status:
        _query = _query.filter(models.SlcRadAccount.status == status)

    if request.path == '/user':
        return render("ops_user_list", page_data = get_page_data(_query),
                       node_list=db.query(models.SlcNode), 
                       product_list=db.query(models.SlcRadProduct),**request.params)
    elif request.path == "/user/export":
        result = _query.all()
        data = Dataset()
        data.append((u'上网账号',u'姓名', u'套餐', u'过期时间', u'余额',u'时长',u'状态',u'创建时间'))
        for i in result:
            data.append((i.account_number, i.realname, i.product_name, i.expire_date,i.balance,i.time_length,i.status,i.create_time))
        name = u"RADIUS-USER-" + datetime.datetime.now().strftime("%Y%m%d-%H%M%S") + ".xls"
        with open(u'./static/xls/%s' % name, 'wb') as f:
            f.write(data.xls)
        return static_file(name, root='./static/xls',download=True)

@app.get('/user/trace',apply=auth_opr)
def user_trace(db):   
    return render("ops_user_trace", bas_list=db.query(models.SlcRadBas))
    
###############################################################################
# online manage      
###############################################################################
    
@app.route('/online',apply=auth_opr,method=['GET','POST'])
def online_query(db): 
    node_id = request.params.get('node_id')
    account_number = request.params.get('account_number')  
    framed_ipaddr = request.params.get('framed_ipaddr')  
    mac_addr = request.params.get('mac_addr')  
    nas_addr = request.params.get('nas_addr')  
    _query = db.query(
        models.SlcRadOnline.id,
        models.SlcRadOnline.account_number,
        models.SlcRadOnline.nas_addr,
        models.SlcRadOnline.acct_session_id,
        models.SlcRadOnline.acct_start_time,
        models.SlcRadOnline.framed_ipaddr,
        models.SlcRadOnline.mac_addr,
        models.SlcRadOnline.nas_port_id,
        models.SlcRadOnline.start_source,
        models.SlcMember.node_id,
        models.SlcMember.realname
    ).filter(
            models.SlcRadOnline.account_number == models.SlcRadAccount.account_number,
            models.SlcMember.member_id == models.SlcRadAccount.member_id
    )
    if node_id:
        _query = _query.filter(models.SlcMember.node_id == node_id)
    if account_number:
        _query = _query.filter(models.SlcRadOnline.account_number.like('%'+account_number+'%'))
    if framed_ipaddr:
        _query = _query.filter(models.SlcRadOnline.framed_ipaddr == framed_ipaddr)
    if mac_addr:
        _query = _query.filter(models.SlcRadOnline.mac_addr == mac_addr)
    if nas_addr:
        _query = _query.filter(models.SlcRadOnline.nas_addr == nas_addr)

    return render("ops_online_list", page_data = get_page_data(_query),
                   node_list=db.query(models.SlcNode), 
                   bas_list=db.query(models.SlcRadBas),**request.params)

###############################################################################
# ticket manage        
###############################################################################

@app.route('/ticket',apply=auth_opr,method=['GET','POST'])
def ticket_query(db): 
    node_id = request.params.get('node_id')
    account_number = request.params.get('account_number')  
    framed_ipaddr = request.params.get('framed_ipaddr')  
    mac_addr = request.params.get('mac_addr')  
    query_begin_time = request.params.get('query_begin_time')  
    query_end_time = request.params.get('query_end_time')  
    _query = db.query(
        models.SlcRadTicket.id,
        models.SlcRadTicket.account_number,
        models.SlcRadTicket.nas_addr,
        models.SlcRadTicket.acct_session_id,
        models.SlcRadTicket.acct_start_time,
        models.SlcRadTicket.acct_stop_time,
        models.SlcRadTicket.framed_ipaddr,
        models.SlcRadTicket.mac_addr,
        models.SlcRadTicket.nas_port_id,
        models.SlcRadTicket.acct_fee,
        models.SlcRadTicket.is_deduct,
        models.SlcMember.node_id,
        models.SlcMember.realname
    ).filter(
            models.SlcRadTicket.account_number == models.SlcRadAccount.account_number,
            models.SlcMember.member_id == models.SlcRadAccount.member_id
    )
    if node_id:
        _query = _query.filter(models.SlcMember.node_id == node_id)
    if account_number:
        _query = _query.filter(models.SlcRadTicket.account_number.like('%'+account_number+'%'))
    if framed_ipaddr:
        _query = _query.filter(models.SlcRadTicket.framed_ipaddr == framed_ipaddr)
    if mac_addr:
        _query = _query.filter(models.SlcRadTicket.mac_addr == mac_addr)
    if query_begin_time:
        _query = _query.filter(models.SlcRadTicket.acct_start_time >= query_begin_time)
    if query_end_time:
        _query = _query.filter(models.SlcRadTicket.acct_stop_time <= query_end_time)

    return render("ops_ticket_list", page_data = get_page_data(_query),
               node_list=db.query(models.SlcNode),**request.params)

