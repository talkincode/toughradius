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
from base import *
from libs import utils
import bottle
import models
import forms
import decimal
import datetime

decimal.getcontext().prec = 11
decimal.getcontext().rounding = decimal.ROUND_UP

app = Bottle()

@app.route('/member',apply=auth_opr,method=['GET','POST'])
@app.get('/member/export',apply=auth_opr)
def member_query(db):   
    node_id = request.params.get('node_id')
    realname = request.params.get('realname')
    _query = db.query(
            models.SlcMember.realname,
            models.SlcMember.member_id,
            models.SlcMember.mobile,
            models.SlcMember.address,
            models.SlcMember.create_time,
            models.SlcNode.node_name
        ).filter(
            models.SlcNode.id == models.SlcMember.node_id
        )
    if node_id:
        _query = _query.filter(models.SlcMember.node_id == node_id)
    if realname:
        _query = _query.filter(models.SlcMember.realname.like('%'+realname+'%'))


    if request.path == '/member':
        return render("bus_member_list", page_data = get_page_data(_query),
                       node_list=db.query(models.SlcNode),**request.params)
    elif request.path == "/member/export":
        result = _query.all()
        data = Dataset()
        data.append((u'区域',u'姓名', u'联系电话', u'地址', u'创建时间'))
        for i in result:
            data.append((i.node_name, i.realname, i.mobile, i.address,i.create_time))
        name = u"RADIUS-MEMBER-" + datetime.datetime.now().strftime("%Y%m%d-%H%M%S") + ".xls"
        with open(u'./static/xls/%s' % name, 'wb') as f:
            f.write(data.xls)
        return static_file(name, root='./static/xls',download=True)    


@app.get('/member/detail',apply=auth_opr)
def member_detail(db): 
    member_id =   request.params.get('member_id')
    member = db.query(models.SlcMember).get(member_id)
    accounts = db.query(
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
        models.SlcMember.member_id == models.SlcRadAccount.member_id,
        models.SlcRadAccount.member_id == member_id
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
        models.SlcRadProduct.product_name
    ).filter(
        models.SlcRadProduct.id == models.SlcMemberOrder.product_id,
        models.SlcMemberOrder.member_id==member_id
    )
    return  render("bus_member_detail",member=member,accounts=accounts,orders=orders)

@app.get('/member/open',apply=auth_opr)
def member_open(db): 
    nodes = [ (n.id,n.node_name) for n in db.query(models.SlcNode)]
    products = [(p.id,p.product_name) for p in db.query(models.SlcRadProduct)]
    form = forms.user_open_form(nodes,products)
    return render("open_form",form=form)

@app.post('/member/open',apply=auth_opr)
def member_open(db): 
    nodes = [ (n.id,n.node_name) for n in db.query(models.SlcNode)]
    products = [(p.id,p.product_name) for p in db.query(models.SlcRadProduct)]    
    form = forms.user_open_form(nodes,products)
    if not form.validates(source=request.forms):
        return render("open_form", form=form)

    if db.query(models.SlcRadAccount).filter_by(
        account_number=form.d.account_number).count()>0:
        return render("open_form", form=form,msg=u"上网账号已经存在")

    member = models.SlcMember()
    member.node_id = form.d.node_id
    member.realname = form.d.realname
    member.idcard = form.d.idcard
    member.sex = '1'
    member.age = '0'
    member.email = ''
    member.mobile = form.d.mobile
    member.address = form.d.address
    member.create_time = utils.get_currtime()
    member.update_time = utils.get_currtime()
    db.add(member) 
    db.flush()
    db.refresh(member)

    order_fee = 0
    balance = 0
    expire_date = form.d.expire_date
    product = db.query(models.SlcRadProduct).get(form.d.product_id)
    if product.product_policy == 0:
        order_fee = decimal.Decimal(product.fee_price) * decimal.Decimal(form.d.months)
        order_fee = int(order_fee.to_integral_value())
    elif product.product_policy == 1:
        balance = utils.yuan2fen(fom.d.fee_value)
        expire_date = '3000-11-11'

    order = models.SlcMemberOrder()
    order.order_id = utils.gen_order_id()
    order.member_id = member.member_id
    order.product_id = product.id
    order.account_number = form.d.account_number
    order.order_fee = order_fee
    order.actual_fee = utils.yuan2fen(form.d.fee_value)
    order.pay_status = 1
    order.order_source = 'admin'
    order.create_time = member.create_time
    db.add(order)

    account = models.SlcRadAccount()
    account.account_number = form.d.account_number
    account.member_id = member.member_id
    account.product_id = order.product_id
    account.install_address = member.address
    account.ip_address = ''
    account.mac_addr = ''
    account.password = utils.encrypt(form.d.password)
    account.status = form.d.status
    account.balance = balance
    account.time_length = 0
    account.expire_date = expire_date
    account.user_concur_number = product.concur_number
    account.bind_mac = product.bind_mac
    account.bind_vlan = product.bind_vlan
    account.vlan_id = 0
    account.vlan_id2 = 0
    account.create_time = member.create_time
    account.update_time = member.create_time
    db.add(account)

    accept_log = models.SlcRadAcceptLog()
    accept_log.accept_type = 'open'
    accept_log.accept_source = 'console'
    accept_log.account_number = account.account_number
    accept_log.accept_time = member.create_time
    accept_log.operator_name = get_cookie("username")
    db.add(accept_log)

    db.commit()
    redirect("/bus/member")


@app.get('/account/open',apply=auth_opr)
def account_open(db): 
    member_id =   request.params.get('member_id')
    member = db.query(models.SlcMember).get(member_id)
    products = [(p.id,p.product_name) for p in db.query(models.SlcRadProduct)]
    form = forms.account_open_form(products)
    form.member_id.set_value(member_id)
    form.realname.set_value(member.realname)
    return render("open_form",form=form)

@app.post('/account/open',apply=auth_opr)
def account_open(db): 
    products = [(p.id,p.product_name) for p in db.query(models.SlcRadProduct)]    
    form = forms.account_open_form(products)
    if not form.validates(source=request.forms):
        return render("open_form", form=form)

    if db.query(models.SlcRadAccount).filter_by(
        account_number=form.d.account_number).count()>0:
        return render("open_form", form=form,msg=u"上网账号已经存在")

    _datetime = utils.get_currtime()
    order_fee = 0
    balance = 0
    expire_date = form.d.expire_date
    product = db.query(models.SlcRadProduct).get(form.d.product_id)
    if product.product_policy == 0:
        order_fee = decimal.Decimal(product.fee_price) * decimal.Decimal(form.d.months)
        order_fee = int(order_fee.to_integral_value())
    elif product.product_policy == 1:
        balance = utils.yuan2fen(fom.d.fee_value)
        expire_date = '3000-11-11'

    order = models.SlcMemberOrder()
    order.order_id = utils.gen_order_id()
    order.member_id = form.d.member_id
    order.product_id = product.id
    order.account_number = form.d.account_number
    order.order_fee = order_fee
    order.actual_fee = utils.yuan2fen(form.d.fee_value)
    order.pay_status = 1
    order.order_source = 'admin'
    order.create_time = _datetime
    db.add(order)

    account = models.SlcRadAccount()
    account.account_number = form.d.account_number
    account.member_id = int(form.d.member_id)
    account.product_id = order.product_id
    account.install_address = form.d.address
    account.ip_address = ''
    account.mac_addr = ''
    account.password = utils.encrypt(form.d.password)
    account.status = form.d.status
    account.balance = balance
    account.time_length = 0
    account.expire_date = expire_date
    account.user_concur_number = product.concur_number
    account.bind_mac = product.bind_mac
    account.bind_vlan = product.bind_vlan
    account.vlan_id = 0
    account.vlan_id2 = 0
    account.create_time = _datetime
    account.update_time = _datetime
    db.add(account)

    db.commit()
    redirect("/bus/member/detail?member_id={}".format(form.d.member_id))

@app.get('/member/import',apply=auth_opr)
def member_import(db): 
    nodes = [ (n.id,n.node_name) for n in db.query(models.SlcNode)]
    products = [(p.id,p.product_name) for p in db.query(models.SlcRadProduct)]
    form = forms.user_import_form(nodes,products)
    return render("import_form",form=form)



