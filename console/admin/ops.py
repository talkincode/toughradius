#!/usr/bin/env python
#coding:utf-8

from bottle import Bottle
from bottle import request
from bottle import response
from bottle import redirect
from bottle import static_file
from bottle import mako_template as render
from tablib import Dataset
from websock import websock
import bottle
import models
import forms
import datetime
from libs import utils
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
            models.SlcRadAccount,
            models.SlcMember.realname,
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
        data.append((u'上网账号',u'姓名', u'资费', u'过期时间', u'余额(元)',u'时长(秒)',u'并发数',u'ip地址',u'状态',u'创建时间'))
        for i,_realname,_product_name in result:
            data.append((
                i.account_number, _realname, _product_name, 
                i.expire_date,utils.fen2yuan(i.balance),
                i.time_length,i.user_concur_number,i.ip_address,
                i.status,i.create_time
            ))
        name = u"RADIUS-USER-" + datetime.datetime.now().strftime("%Y%m%d-%H%M%S") + ".xls"
        with open(u'./static/xls/%s' % name, 'wb') as f:
            f.write(data.xls)
        return static_file(name, root='./static/xls',download=True)

@app.get('/user/trace',apply=auth_opr)
def user_trace(db):   
    return render("ops_user_trace", bas_list=db.query(models.SlcRadBas))


                   
@app.get('/user/detail',apply=auth_opr)
def user_detail(db):   
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
        return render("error",msg=u"用户不存在")
    user_attrs = db.query(models.SlcRadAccountAttr).filter_by(account_number=account_number)
    return render("ops_user_detail",user=user,user_attrs=user_attrs)

@app.post('/user/release',apply=auth_opr)
def user_release(db):   
    account_number = request.params.get('account_number')  
    user = db.query(models.SlcRadAccount).filter_by(account_number=account_number)
    user.mac_addr = ''
    user.vlan_id = 0
    user.vlan_id2 = 0

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'释放用户账号（%s）绑定信息'%(account_number,)
    db.add(ops_log)

    db.commit()
    websock.update_cache("account",account_number=account_number)
    return dict(code=0,msg=u"解绑成功")

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
        models.SlcRadOnline.billing_times,
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

    _query = _query.order_by(models.SlcRadOnline.acct_start_time.desc())
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

    _query = _query.order_by(models.SlcRadTicket.acct_start_time.desc())
    return render("ops_ticket_list", page_data = get_page_data(_query),
               node_list=db.query(models.SlcNode),**request.params)


###############################################################################
# ops log manage        
###############################################################################

@app.route('/opslog',apply=auth_opr,method=['GET','POST'])
def opslog_query(db): 
    node_id = request.params.get('node_id')
    operator_name = request.params.get('operator_name')
    query_begin_time = request.params.get('query_begin_time')  
    query_end_time = request.params.get('query_end_time')  
    _query = db.query(
        models.SlcRadOperateLog,
        models.SlcOperator.node_id,
    ).filter(
        models.SlcRadOperateLog.operator_name == models.SlcOperator.operator_name,
    )
    if node_id:
        _query = _query.filter(models.SlcOperator.node_id == node_id)    
    if operator_name:
        _query = _query.filter(models.SlcRadOperateLog.operator_name == operator_name)
    if query_begin_time:
        _query = _query.filter(models.SlcRadOperateLog.operate_time >= query_begin_time)
    if query_end_time:
        _query = _query.filter(models.SlcRadOperateLog.operate_time <= query_end_time)
    _query = _query.order_by(models.SlcRadOperateLog.operate_time.desc())
    return render("ops_log_list", 
        node_list=db.query(models.SlcNode),
        page_data = get_page_data(_query),**request.params)
