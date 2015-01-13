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
        models.SlcRadAccount.last_pause,
        models.SlcRadAccount.create_time,
        models.SlcRadProduct.product_name,
        models.SlcRadProduct.product_policy
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
        models.SlcMemberOrder.order_desc,
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

@app.get('/member/update',apply=auth_opr)
def member_update(db): 
    member_id = request.params.get("member_id")
    member = db.query(models.SlcMember).get(member_id)
    nodes = [ (n.id,n.node_name) for n in db.query(models.SlcNode)]
    form = forms.member_update_form(nodes)
    form.fill(member)
    return render("base_form",form=form)

@app.post('/member/open',apply=auth_opr)
def member_open(db): 
    nodes = [ (n.id,n.node_name) for n in db.query(models.SlcNode)]
    products = [(p.id,p.product_name) for p in db.query(models.SlcRadProduct)]    
    form = forms.user_open_form(nodes,products)
    if not form.validates(source=request.forms):
        return render("open_form", form=form)

    if db.query(models.SlcRadAccount).filter_by(account_number=form.d.account_number).count()>0:
        return render("open_form", form=form,msg=u"上网账号%s已经存在"%form.d.account_number)

    if db.query(models.SlcRadAccount).filter_by(ip_address=form.d.ip_address).count()>0:
        return render("open_form", form=form,msg=u"ip%s已经被使用"%form.d.ip_address)

    if db.query(models.SlcMember).filter_by(
        member_name=form.d.member_name).count()>0:
        return render("open_form", form=form,msg=u"用户名%s已经存在"%form.d.member_name)

    member = models.SlcMember()
    member.node_id = form.d.node_id
    member.realname = form.d.realname
    member.member_name = form.d.member_name
    member.password = md5(form.d.password.encode()).hexdigest()
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

    accept_log = models.SlcRadAcceptLog()
    accept_log.accept_type = 'open'
    accept_log.accept_source = 'console'
    accept_log.account_number = form.d.account_number
    accept_log.accept_time = member.create_time
    accept_log.operator_name = get_cookie("username")
    accept_log.accept_desc = u"用户新开户：(%s)%s - 上网账号:%s"%(member.member_name,member.realname,form.d.account_number)
    db.add(accept_log)
    db.flush()
    db.refresh(accept_log)    

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
    order.accept_id = accept_log.id
    order.order_source = 'console'
    order.create_time = member.create_time
    order.order_desc = u"用户新开账号"
    db.add(order)

    account = models.SlcRadAccount()
    account.account_number = form.d.account_number
    account.ip_address = form.d.ip_address
    account.member_id = member.member_id
    account.product_id = order.product_id
    account.install_address = member.address
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

    if db.query(models.SlcRadAccount).filter_by(ip_address=form.d.ip_address).count()>0:
        return render("open_form", form=form,msg=u"ip%s已经被使用"%form.d.ip_address)

    accept_log = models.SlcRadAcceptLog()
    accept_log.accept_type = 'open'
    accept_log.accept_source = 'console'
    accept_log.account_number = form.d.account_number
    accept_log.accept_time = utils.get_currtime()
    accept_log.operator_name = get_cookie("username")
    accept_log.accept_desc = u"用户新增账号：(%s)%s - 上网账号:%s"%(member.member_name,member.realname,form.d.account_number)
    db.add(accept_log)
    db.flush()
    db.refresh(accept_log)

    _datetime = utils.get_currtime()
    order_fee = 0
    balance = 0
    expire_date = form.d.expire_date
    product = db.query(models.SlcRadProduct).get(form.d.product_id)
    if product.product_policy == 0:
        order_fee = decimal.Decimal(product.fee_price) * decimal.Decimal(form.d.months)
        order_fee = int(order_fee.to_integral_value())
    elif product.product_policy == 1:
        balance = utils.yuan2fen(form.d.fee_value)
        expire_date = '3000-11-11'

    order = models.SlcMemberOrder()
    order.order_id = utils.gen_order_id()
    order.member_id = form.d.member_id
    order.product_id = product.id
    order.account_number = form.d.account_number
    order.order_fee = order_fee
    order.actual_fee = utils.yuan2fen(form.d.fee_value)
    order.pay_status = 1
    order.accept_id = accept_log.id   
    order.order_source = 'console'
    order.create_time = _datetime
    order.order_desc = u"用户增开账号"
    db.add(order)

    account = models.SlcRadAccount()
    account.account_number = form.d.account_number
    account.ip_address = form.d.ip_address
    account.member_id = int(form.d.member_id)
    account.product_id = order.product_id
    account.install_address = form.d.address
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

@app.post('/member/import',apply=auth_opr)
def member_import(db): 
    nodes = [ (n.id,n.node_name) for n in db.query(models.SlcNode)]
    products = [(p.id,p.product_name) for p in db.query(models.SlcRadProduct)]
    iform = forms.user_import_form(nodes,products)
    node_id =   request.params.get('node_id')
    product_id =   request.params.get('product_id')
    upload = request.files.get('import_file')
    impctx = upload.file.read()
    lines = impctx.split("\n")
    _num = 0
    impusers = []
    for line in lines:
        _num += 1
        line = line.strip()
        if not line or u"用户姓名" in line:continue
        attr_array = line.split(",")
        if len(attr_array) < 5:
            return render("import_form",form=iform,msg=u"line %s error: length must 5 "%_num)

        vform = forms.user_import_vform()
        if not vform.validates(dict(
                realname = attr_array[0],
                account_number = attr_array[1], 
                password = attr_array[2],
                expire_date = attr_array[3],
                balance = attr_array[4])):
            return render("import_form",form=iform,msg=u"line %s error: %s"%(_num,vform.errors))

        impusers.append(vform)

    for form in impusers:
        try:
            member = models.SlcMember()
            member.node_id = node_id
            member.realname = form.d.realname
            member.idcard = '123456'
            member.member_name = form.d.account_number
            member.password = form.d.account_number
            member.sex = '1'
            member.age = '0'
            member.email = ''
            member.mobile = '123456'
            member.address = 'address'
            member.create_time = utils.get_currtime()
            member.update_time = utils.get_currtime()
            db.add(member) 
            db.flush()
            db.refresh(member)

            accept_log = models.SlcRadAcceptLog()
            accept_log.accept_type = 'open'
            accept_log.accept_source = 'console'
            accept_log.accept_desc = u"用户导入账号：(%s)%s - 上网账号:%s"%(member.member_name,member.realname,form.d.account_number)
            accept_log.account_number = form.d.account_number
            accept_log.accept_time = member.create_time
            accept_log.operator_name = get_cookie("username")
            db.add(accept_log)
            db.flush()
            db.refresh(accept_log)            

            order_fee = 0
            actual_fee = 0 
            balance = 0
            expire_date = form.d.expire_date
            product = db.query(models.SlcRadProduct).get(product_id)
            if product.product_policy == 1:
                balance = int(form.d.balance)
                expire_date = '3000-11-11'

            order = models.SlcMemberOrder()
            order.order_id = utils.gen_order_id()
            order.member_id = member.member_id
            order.product_id = product.id
            order.account_number = form.d.account_number
            order.order_fee = order_fee
            order.actual_fee = actual_fee
            order.pay_status = 1
            order.accept_id = accept_log.id 
            order.order_source = 'console'
            order.create_time = member.create_time
            order.order_desc = u"用户导入开户"
            db.add(order)

            account = models.SlcRadAccount()
            account.account_number = form.d.account_number
            account.member_id = member.member_id
            account.product_id = order.product_id
            account.install_address = member.address
            account.ip_address = ''
            account.mac_addr = ''
            account.password = utils.encrypt(form.d.password)
            account.status = 1
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

        except Exception as e:
            return render("import_form",form=iform,msg=u"error : %s"%str(e))

    db.commit()
    redirect("/bus/member")

@app.post('/account/pause',apply=auth_opr)
def account_pause(db): 
    account_number = request.params.get("account_number")
    account = db.query(models.SlcRadAccount).get(account_number)

    if account.status != 1:
        return dict(msg=u"用户当前状态不允许停机")

    _datetime = utils.get_currtime()
    account.last_pause = _datetime
    account.status = 2

    accept_log = models.SlcRadAcceptLog()
    accept_log.accept_type = 'pause'
    accept_log.accept_source = 'console'
    accept_log.accept_desc = u"用户停机：上网账号:%s"%(account_number)
    accept_log.account_number = account.account_number
    accept_log.accept_time = _datetime
    accept_log.operator_name = get_cookie("username")
    db.add(accept_log)

    db.commit()
    return dict(msg=u"操作成功")

@app.post('/account/resume',apply=auth_opr)
def account_resume(db): 
    account_number = request.params.get("account_number")
    account = db.query(models.SlcRadAccount).get(account_number)
    if account.status != 2:
        return dict(msg=u"用户当前状态不允许复机") 

    account.status = 1
    _datetime = datetime.datetime.now()
    _pause_time = datetime.datetime.strptime(account.last_pause,"%Y-%m-%d %H:%M:%S")
    _expire_date = datetime.datetime.strptime(account.expire_date+' 23:59:59',"%Y-%m-%d %H:%M:%S")
    days = (_expire_date - _pause_time).days
    new_expire = (_datetime + datetime.timedelta(days=int(days))).strftime("%Y-%m-%d")
    account.expire_date = new_expire

    accept_log = models.SlcRadAcceptLog()
    accept_log.accept_type = 'resume'
    accept_log.accept_source = 'console'
    accept_log.accept_desc = u"用户复机：上网账号:%s"%(account_number)
    accept_log.account_number = account.account_number
    accept_log.accept_time = utils.get_currtime()
    accept_log.operator_name = get_cookie("username")
    db.add(accept_log)

    db.commit()
    return dict(msg=u"操作成功")


def query_account(db,account_number):
    return db.query(
        models.SlcMember.realname,
        models.SlcRadAccount.member_id,
        models.SlcRadAccount.product_id,
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


@app.get('/account/next',apply=auth_opr)
def account_next(db): 
    account_number = request.params.get("account_number")
    user = query_account(db,account_number)
    form = forms.account_next_form()
    form.account_number.set_value(account_number)
    return render("account_form",user=user,form=form)

@app.post('/account/next',apply=auth_opr)
def account_next(db): 
    account_number = request.params.get("account_number")
    account = db.query(models.SlcRadAccount).get(account_number)
    user = query_account(db,account_number)
    form = forms.account_next_form()
    if account.status not in (1,4):
        return render("account_form", user=user,form=form,msg=u"无效用户状态")    
    if not form.validates(source=request.forms):
        return render("account_form", user=user,form=form)

    accept_log = models.SlcRadAcceptLog()
    accept_log.accept_type = 'next'
    accept_log.accept_source = 'console'
    accept_log.accept_desc = u"用户续费：上网账号:%s，续费%s元"%(account_number,form.d.fee_value)
    accept_log.account_number = form.d.account_number
    accept_log.accept_time = utils.get_currtime()
    accept_log.operator_name = get_cookie("username")
    db.add(accept_log)
    db.flush()
    db.refresh(accept_log)

    order_fee = 0
    product = db.query(models.SlcRadProduct).get(user.product_id)
    order_fee = decimal.Decimal(product.fee_price) * decimal.Decimal(form.d.months)
    order_fee = int(order_fee.to_integral_value())

    order = models.SlcMemberOrder()
    order.order_id = utils.gen_order_id()
    order.member_id = user.member_id
    order.product_id = user.product_id
    order.account_number = form.d.account_number
    order.order_fee = order_fee
    order.actual_fee = utils.yuan2fen(form.d.fee_value)
    order.pay_status = 1
    order.accept_id = accept_log.id
    order.order_source = 'console'
    order.create_time = utils.get_currtime()
    order.order_desc = u"用户续费"
    db.add(order)  

    account.status = 1
    account.expire_date = form.d.expire_date  

    db.commit()
    redirect("/bus/member/detail?member_id={}".format(user.member_id))

@app.get('/account/charge',apply=auth_opr)
def account_charge(db): 
    account_number = request.params.get("account_number")
    user = query_account(db,account_number)
    form = forms.account_charge_form()
    form.account_number.set_value(account_number)
    return render("account_form",user=user,form=form)

@app.post('/account/charge',apply=auth_opr)
def account_charge(db): 
    account_number = request.params.get("account_number")
    account = db.query(models.SlcRadAccount).get(account_number)
    user = query_account(db,account_number)
    form = forms.account_charge_form()
    if account.status !=1 :
        return render("account_form", user=user,form=form,msg=u"无效用户状态")  

    if not form.validates(source=request.forms):
        return render("account_form", user=user,form=form)

    accept_log = models.SlcRadAcceptLog()
    accept_log.accept_type = 'charge'
    accept_log.accept_source = 'console'
    accept_log.account_number = form.d.account_number
    accept_log.accept_time = utils.get_currtime()
    accept_log.operator_name = get_cookie("username")
    _new_fee = account.balance + utils.yuan2fen(form.d.fee_value)
    accept_log.accept_desc = u"用户充值：上网账号:%s，充值前%s(分),充值后%s(分)"%(account_number,account.balance,_new_fee)
    db.add(accept_log)
    db.flush()
    db.refresh(accept_log)

    order = models.SlcMemberOrder()
    order.order_id = utils.gen_order_id()
    order.member_id = user.member_id
    order.product_id = user.product_id
    order.account_number = form.d.account_number
    order.order_fee = 0
    order.actual_fee = utils.yuan2fen(form.d.fee_value)
    order.pay_status = 1
    order.accept_id = accept_log.id
    order.order_source = 'console'
    order.create_time = utils.get_currtime()
    order.order_desc = u"用户充值"
    db.add(order)  

    account.balance += order.actual_fee

    db.commit()
    redirect("/bus/member/detail?member_id={}".format(user.member_id))

@app.get('/account/cancel',apply=auth_opr)
def account_cancel(db): 
    account_number = request.params.get("account_number")
    user = query_account(db,account_number)
    form = forms.account_cancel_form()
    form.account_number.set_value(account_number)
    return render("account_form",user=user,form=form)


@app.post('/account/cancel',apply=auth_opr)
def account_next(db): 
    account_number = request.params.get("account_number")
    account = db.query(models.SlcRadAccount).get(account_number)
    user = query_account(db,account_number)
    form = forms.account_cancel_form()
    if account.status !=1 :
        return render("account_form", user=user,form=form,msg=u"无效用户状态")      
    if not form.validates(source=request.forms):
        return render("account_form", user=user,form=form)

    accept_log = models.SlcRadAcceptLog()
    accept_log.accept_type = 'cancel'
    accept_log.accept_source = 'console'
    accept_log.account_number = form.d.account_number
    accept_log.accept_time = utils.get_currtime()
    accept_log.operator_name = get_cookie("username")
    accept_log.accept_desc = u"用户销户：上网账号:%s，退费%s(元)"%(account_number,form.d.fee_value)
    db.add(accept_log)
    db.flush()
    db.refresh(accept_log)

    order = models.SlcMemberOrder()
    order.order_id = utils.gen_order_id()
    order.member_id = user.member_id
    order.product_id = user.product_id
    order.account_number = form.d.account_number
    order.order_fee = 0
    order.actual_fee = -utils.yuan2fen(form.d.fee_value)
    order.pay_status = 1
    order.order_source = 'console'
    order.accept_id = accept_log.id
    order.create_time = utils.get_currtime()
    order.order_desc = u'用户销户'
    db.add(order)  

    account.status = 3

    db.commit()
    redirect("/bus/member/detail?member_id={}".format(user.member_id))


###############################################################################
# accept log manage        
###############################################################################

@app.route('/acceptlog',apply=auth_opr,method=['GET','POST'])
def opslog_query(db): 
    node_id = request.params.get('node_id')
    accept_type = request.params.get('accept_type')
    account_number = request.params.get('account_number')  
    operator_name = request.params.get('operator_name')
    query_begin_time = request.params.get('query_begin_time')  
    query_end_time = request.params.get('query_end_time')  
    _query = db.query(
        models.SlcRadAcceptLog.id,
        models.SlcRadAcceptLog.accept_type,
        models.SlcRadAcceptLog.accept_time,
        models.SlcRadAcceptLog.accept_desc,
        models.SlcRadAcceptLog.operator_name,
        models.SlcRadAcceptLog.accept_source,
        models.SlcRadAcceptLog.account_number,
        models.SlcMember.node_id,
    ).filter(
            models.SlcRadAcceptLog.account_number == models.SlcRadAccount.account_number,
            models.SlcMember.member_id == models.SlcRadAccount.member_id
    )
    if operator_name:
        _query = _query.filter(models.SlcRadAcceptLog.operator_name == operator_name)
    if node_id:
        _query = _query.filter(models.SlcMember.node_id == node_id)
    if account_number:
        _query = _query.filter(models.SlcRadAcceptLog.account_number == account_number)
    if accept_type:
        _query = _query.filter(models.SlcRadAcceptLog.accept_type == accept_type)
    if query_begin_time:
        _query = _query.filter(models.SlcRadAcceptLog.accept_time >= query_begin_time)
    if query_end_time:
        _query = _query.filter(models.SlcRadAcceptLog.accept_time <= query_end_time)

    type_map = {'open':u'开户','pause':u'停机','resume':u'复机','cancel':u'销户','next':u'续费','charge':u'充值'}   
    return render(
        "bus_acceptlog_list", 
        page_data = get_page_data(_query),
        node_list=db.query(models.SlcNode),
        type_map = type_map,
        get_orderid = lambda aid:db.query(models.SlcMemberOrder.order_id).filter_by(accept_id=aid).scalar(),
        **request.params
    )

