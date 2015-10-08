#!/usr/bin/env python
#coding:utf-8

from bottle import Bottle
from bottle import request
from bottle import response
from bottle import redirect
from bottle import static_file
from bottle import mako_template as render
from tablib import Dataset
from toughradius.console.websock import websock
from toughradius.console import models
from toughradius.console.libs import utils
from toughradius.console.base import *
from toughradius.console.admin import forms
import bottle
import datetime
from sqlalchemy import func

__prefix__ = "/ops"

app = Bottle()
app.config['__prefix__'] = __prefix__

###############################################################################
# user manage        
###############################################################################
                   
@app.route('/user',apply=auth_opr,method=['GET','POST'])
def user_query(db, render):
    node_id = request.params.get('node_id')
    product_id = request.params.get('product_id')
    user_name = request.params.get('user_name')
    status = request.params.get('status')
    opr_nodes = get_opr_nodes(db)
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
    else:
        _query = _query.filter(models.SlcMember.node_id.in_([i.id for i in opr_nodes]))
    if product_id:
        _query = _query.filter(models.SlcRadAccount.product_id==product_id)
    if user_name:
        _query = _query.filter(models.SlcRadAccount.account_number.like('%'+user_name+'%'))

    # 用户状态判断
    _now = datetime.datetime.now()
    if status:
        if status == '4':
            _query = _query.filter(models.SlcRadAccount.expire_date <= _now.strftime("%Y-%m-%d"))
        elif status == '1':
            _query = _query.filter(
                models.SlcRadAccount.status == status,
                models.SlcRadAccount.expire_date >= _now.strftime("%Y-%m-%d")
            )
        else:
            _query = _query.filter(models.SlcRadAccount.status == status)

    if request.path == '/user':
        return render("ops_user_list", page_data=get_page_data(_query),
                       node_list=opr_nodes, 
                       product_list=db.query(models.SlcRadProduct),**request.params)
                       
permit.add_route("%s/user"%__prefix__,u"用户账号查询",u"维护管理",is_menu=True,order=0)

@app.get('/user/trace',apply=auth_opr)
def user_trace(db, render):
    return render("ops_user_trace", bas_list=db.query(models.SlcRadBas))

permit.add_route("%s/user/trace"%__prefix__,u"用户消息跟踪",u"维护管理",is_menu=True,order=1)
                   
@app.get('/user/detail',apply=auth_opr)
def user_detail(db, render):
    account_number = request.params.get('account_number')  
    user  = db.query(
        models.SlcMember.realname,
        models.SlcRadAccount.member_id,
        models.SlcRadAccount.account_number,
        models.SlcRadAccount.expire_date,
        models.SlcRadAccount.balance,
        models.SlcRadAccount.time_length,
        models.SlcRadAccount.flow_length,
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
    
permit.add_route("%s/user/detail"%__prefix__,u"账号详情",u"维护管理",order=1.01)

@app.post('/user/release',apply=auth_opr)
def user_release(db, render):
    account_number = request.params.get('account_number')  
    user = db.query(models.SlcRadAccount).filter_by(account_number=account_number).first()
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
    
permit.add_route("%s/user/release"%__prefix__,u"用户释放绑定",u"维护管理",order=1.02)    


###############################################################################
# ops log manage        
###############################################################################

@app.route('/opslog',apply=auth_opr,method=['GET','POST'])
def opslog_query(db, render):
    operator_name = request.params.get('operator_name')
    query_begin_time = request.params.get('query_begin_time')  
    query_end_time = request.params.get('query_end_time')  
    keyword = request.params.get('keyword')
    opr_nodes = get_opr_nodes(db)
    _query = db.query(models.SlcRadOperateLog).filter(
        models.SlcRadOperateLog.operator_name == models.SlcOperator.operator_name,
    ) 
    if operator_name:
        _query = _query.filter(models.SlcRadOperateLog.operator_name == operator_name)
    if keyword:
        _query = _query.filter(models.SlcRadOperateLog.operate_desc.like("%"+keyword+"%"))        
    if query_begin_time:
        _query = _query.filter(models.SlcRadOperateLog.operate_time >= query_begin_time+' 00:00:00')
    if query_end_time:
        _query = _query.filter(models.SlcRadOperateLog.operate_time <= query_end_time+' 23:59:59')
    _query = _query.order_by(models.SlcRadOperateLog.operate_time.desc())
    return render("ops_log_list", 
        node_list=opr_nodes,
        page_data = get_page_data(_query),**request.params)


permit.add_route("%s/opslog"%__prefix__,u"操作日志查询",u"维护管理",is_menu=True,order=4)



