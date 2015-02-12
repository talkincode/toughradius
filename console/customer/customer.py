#!/usr/bin/env python
#coding:utf-8

from bottle import Bottle
from bottle import request
from bottle import response
from bottle import HTTPResponse
from bottle import redirect
from bottle import run as runserver
from bottle import static_file
from bottle import abort
from bottle import mako_template as render
from hashlib import md5
from tablib import Dataset
from libs import sqla_plugin 
from urlparse import urljoin
from base import (
    logger,set_cookie,get_cookie,cache,get_param_value,
    auth_cus,get_member_by_name,get_page_data,
    get_account_by_number,get_online_status
)
from base import (PPMonth,PPTimes,BOMonth,BOTimes,PPFlow,BOFlows)
from base import  (CardInActive,CardActive,CardUsed,CardRecover)
from base import (ProductCard,BalanceCard)
from base import (UsrPreAuth,UsrNormal,UsrPause,UsrCancel,UsrExpire) 
from libs import utils
from sqlalchemy.sql import exists
from websock import websock
import time
import bottle
import models
import forms
import decimal
import datetime
import functools

app = Bottle()

###############################################################################
# login , recharge error times limit    
###############################################################################   

class ValidateCache(object):
    validates = {}
    def incr(self,mid,vid):
        key = "%s_%s"%(mid,vid)
        if key not in self.validates:
            self.validates[key] = [1,time.time()]
        else:
            self.validates[key][0] += 1
            
    def errs(self,mid,vid):
        key = "%s_%s"%(mid,vid)    
        if key in  self.validates:
            return self.validates[key][0] 
        return 0
    
    def clear(self,mid,vid):
        key = "%s_%s"%(mid,vid)    
        if key in  self.validates:
            del self.validates[key]
        
    def is_over(self,mid,vid):
        key = "%s_%s"%(mid,vid)
        if key not in self.validates:
            return False
        elif (time.time() - self.validates[key][1]) > 3600:
            del self.validates[key]
            return False
        else:
            return self.validates[key][0] >= 5 

vcache = ValidateCache() 
              
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
# index handle         
###############################################################################
@cache.cache('customer_index_get_data',expire=180)  
def get_data(db,member_name):
    member = db.query(models.SlcMember).filter_by(member_name=member_name).first()
    accounts = db.query(
        models.SlcMember.realname,
        models.SlcRadAccount.member_id,
        models.SlcRadAccount.account_number,
        models.SlcRadAccount.expire_date,
        models.SlcRadAccount.balance,
        models.SlcRadAccount.time_length,
        models.SlcRadAccount.flow_length,
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
        
@app.get('/',apply=auth_cus)
def customer_index(db):
    member,accounts,orders = get_data(db,get_cookie('customer'))
    status_colors = {0:'',1:'',2:'class="warning"',3:'class="danger"',4:'class="warning"'}
    online_colors = lambda a : get_online_status(db,a) and 'class="success"' or ''
    return  render("index",
        member=member,
        accounts=accounts,
        orders=orders,
        status_colors=status_colors,
        online_colors = online_colors
    )    

###############################################################################
# user login        
###############################################################################

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
    
    if vcache.is_over(form.d.username,'0'):
        return render("error",msg=u"用户一小时内登录错误超过5次，请一小时后再试")

    member = db.query(models.SlcMember).filter_by(
        member_name=form.d.username
    ).first()
    
    if not member:
        return render("login", form=form,msg=u"用户不存在")
    
    if member.password != md5(form.d.password.encode()).hexdigest():
        vcache.incr(form.d.username,'0')
        print vcache.validates
        return render("login", form=form,msg=u"用户名密码错误第%s次"%vcache.errs(form.d.username,'0'))
 
    vcache.clear(form.d.username,'0')
 
    set_cookie('customer_id',member.member_id)
    set_cookie('customer',form.d.username)
    set_cookie('customer_login_time', utils.get_currtime())
    set_cookie('customer_login_ip', request.remote_addr) 
    redirect(next)

@app.get("/logout")
def member_logout():
    set_cookie('customer_id',None)
    set_cookie('customer',None)
    set_cookie('customer_login_time', None)
    set_cookie('customer_login_ip', None)     
    request.cookies.clear()
    redirect('login')

###############################################################################
# user join        
###############################################################################

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
   
    logger.info(u"新用户注册成功,member_name=%s"%member.member_name)
    redirect('/login')
   
###############################################################################
# account query        
###############################################################################
   
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
        return render("error",msg=u"账号不存在")
    return render("account_detail",user=user)
     
@app.get('/product/list',apply=auth_cus)
def product_list(db):
    return render("product_list",products=db.query(models.SlcRadProduct).filter_by(
        product_status = 0
    ))
    
###############################################################################
# billing query        
###############################################################################
    
@app.route('/billing',apply=auth_cus,method=['GET','POST'])
def billing_query(db): 
    account_number = request.params.get('account_number')  
    query_begin_time = request.params.get('query_begin_time')  
    query_end_time = request.params.get('query_end_time')  
    
    @cache.cache('billing_query_result',expire=180)  
    def query_result(account_number,query_begin_time,query_end_time):
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
        return _query.order_by(models.SlcRadBilling.create_time.desc())
        
    query = query_result(account_number,query_begin_time,query_end_time)
    return render("billing_list", 
        accounts=db.query(models.SlcRadAccount).filter_by(member_id=get_cookie("customer_id")),
        page_data=get_page_data(query),**request.params)
        

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
        models.SlcRadTicket.acct_input_gigawords,
        models.SlcRadTicket.acct_output_gigawords,
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
        
###############################################################################
# password update    
###############################################################################    
    
@app.get('/password/update',apply=auth_cus)    
def password_update_get(db):
    form = forms.password_update_form()
    account_number = request.params.get('account_number')      
    form.account_number.set_value(account_number)
    return render("base_form",form=form)
    
@app.post('/password/update',apply=auth_cus)    
def password_update_post(db):
    form = forms.password_update_form()
    if not form.validates(source=request.forms):
        return render("base_form", form=form)
        
    account = db.query(models.SlcRadAccount).filter_by(account_number=form.d.account_number).first()
    if not account:
        return render("base_form", form=form,msg=u'没有这个账号')
        
    if account.member_id != get_cookie("customer_id"):
        return render("base_form", form=form,msg=u'该账号用用户不匹配')
    
    if utils.decrypt(account.password) !=  form.d.old_password:
        return render("base_form", form=form,msg=u'旧密码不正确')
        
    if form.d.new_password != form.d.new_password2:
        return render("base_form", form=form,msg=u'确认新密码不匹配')
    
    account.password =  utils.encrypt(form.d.new_password)
    db.commit()
    websock.update_cache("account",account_number=account.account_number)
    redirect("/")

###############################################################################
# portal auth        
###############################################################################
    
@app.route('/portal/auth')
def portal_auth(db):
    user = request.params.get("user")
    token = request.params.get("token")
    if not user:return abort(403,'user is empty')
    if not token:return abort(403,'token is empty')
    account = db.query(models.SlcRadAccount).filter_by(
        account_number=user
    ).first()
    if not account:
        return abort(403,'user not exists')
    secret = get_param_value(db,"8_portal_secret")
    date = utils.get_currdate()
    _token = md5("%s%s%s%s"%(user,utils.decrypt(account.password),secret,date)).hexdigest()
    if _token == token:
        member = db.query(models.SlcMember).get(account.member_id)
        set_cookie('customer_id',member.member_id,path="/")
        set_cookie('customer',member.member_name,path="/")
        set_cookie('customer_login_time', utils.get_currtime(),path="/")
        set_cookie('customer_login_ip', request.remote_addr,path="/") 
        redirect("/")
    else:
        return abort(403,'token is invalid')
        
###############################################################################
# account open      
###############################################################################

def check_card(card):
    if not card:
        return dict(code=1,data=u"充值卡不存在")
    if card.card_status == CardInActive:
        return dict(code=1,data=u"充值卡未激活")
    elif card.card_status == CardUsed:
        return dict(code=1,data=u"充值卡已被使用")
    elif card.card_status == CardRecover:
        return dict(code=1,data=u"充值卡已被回收")
    if card.expire_date < utils.get_currdate():
        return dict(code=1,data=u"充值卡已过期")
    return dict(code=0)

@app.get('/querycp',apply=auth_cus)
def query_card_products(db):
    ''' query product by card'''
    recharge_card = request.params.get('recharge_card')
    card = db.query(models.SlcRechargerCard).filter_by(card_number=recharge_card).first()  

    check_result = check_card(card)
    if check_result['code'] > 0:
        return check_result
    
    if card.card_type == BalanceCard:
        products = [ (n.id,n.product_name) for n in db.query(models.SlcRadProduct).filter(
            models.SlcRadProduct.product_status == 0,
            models.SlcRadProduct.product_policy.in_((PPTimes,PPFlow))
        )]
        return dict(code=0,data={'products':products})
    elif card.card_type == ProductCard:
        product = db.query(models.SlcRadProduct).get(card.product_id)
        return dict(code=0,data={'products':[(product.id,product.product_name)]})
    

@app.get('/open',apply=auth_cus)
def account_open(db):
    r = ['0','1','2','3','4','5','6','7','8','9']
    rg = utils.random_generator
    def random_account():
        _num = ''.join([rg.choice(r) for _ in range(9)])
        if db.query(models.SlcRadAccount).filter_by(account_number=_num).count() > 0:
            return random_account()
        else:
            return _num

    form = forms.account_open_form()
    form.recharge_card.set_value('')
    form.recharge_pwd.set_value('')
    form.account_number.set_value(random_account())
    return render('card_open_form',form=form)    

@app.post('/open',apply=auth_cus)
def account_open(db):
    form = forms.account_open_form()
    if not form.validates(source=request.forms):
        return render("card_open_form", form=form)
    if vcache.is_over(get_cookie("customer_id"),form.d.recharge_card):
         return render("card_open_form", form=form,msg=u"该充值卡一小时内密码输入错误超过5次，请一小时后再试") 

    card = db.query(models.SlcRechargerCard).filter_by(card_number=form.d.recharge_card).first()  
    check_result = check_card(card)
    if check_result['code'] > 0:
        return render('card_open_form',form=form,msg=check_result['data'])

    if utils.decrypt(card.card_passwd) != form.d.recharge_pwd:
        vcache.incr(get_cookie("customer_id"),form.d.recharge_card)
        errs = vcache.errs(get_cookie("customer_id"),form.d.recharge_card)
        return render('card_open_form',form=form,msg=u"充值卡密码错误%s次"%errs)
    
    vcache.clear(get_cookie("customer_id"),form.d.recharge_card)
    
    # start open
    accept_log = models.SlcRadAcceptLog()
    accept_log.accept_type = 'open'
    accept_log.accept_source = 'customer'
    accept_log.account_number = form.d.account_number
    accept_log.accept_time = utils.get_currtime()
    accept_log.operator_name = "customer"
    accept_log.accept_desc = u"用户自助开户:%s"%(form.d.account_number)
    db.add(accept_log)
    db.flush()
    db.refresh(accept_log)
    
    _datetime = utils.get_currtime()
    order_fee = 0
    balance = 0
    expire_date = utils.add_months(datetime.datetime.now(),card.months).strftime("%Y-%m-%d") 
    product = db.query(models.SlcRadProduct).get(form.d.product_id)
    # 预付费包月
    if product.product_policy == PPMonth:
        order_fee = decimal.Decimal(product.fee_price) * decimal.Decimal(card.months)
        order_fee = int(order_fee.to_integral_value())
    # 买断包月,买断时长,买断流量
    elif product.product_policy in (BOMonth,BOTimes,BOFlows):
        order_fee = int(product.fee_price)
    #预付费时长,预付费流量
    elif product.product_policy in (PPTimes,PPFlow):
        balance = card.fee_value
        expire_date = '3000-11-11'
    
    order = models.SlcMemberOrder()
    order.order_id = utils.gen_order_id()
    order.member_id = get_cookie("customer_id")
    order.product_id = product.id
    order.account_number = form.d.account_number
    order.order_fee = order_fee
    order.actual_fee = card.fee_value
    order.pay_status = 1
    order.accept_id = accept_log.id
    order.order_source = 'customer'
    order.create_time = _datetime
    order.order_desc = u"用户使用充值卡[ %s ]开户"%form.d.recharge_card
    db.add(order)
    
    account = models.SlcRadAccount()
    account.account_number = form.d.account_number
    account.ip_address = ''
    account.member_id = get_cookie("customer_id")
    account.product_id = order.product_id
    account.install_address = ''
    account.mac_addr = ''
    account.password = utils.encrypt(form.d.password)
    account.status = 1
    account.balance = balance
    account.time_length = int(product.fee_times)
    account.flow_length = int(product.fee_flows)
    account.expire_date = expire_date
    account.user_concur_number = product.concur_number
    account.bind_mac = product.bind_mac
    account.bind_vlan = product.bind_vlan
    account.vlan_id = 0
    account.vlan_id2 = 0
    account.create_time = _datetime
    account.update_time = _datetime
    db.add(account)
    
    clog = models.SlcRechargeLog()
    clog.member_id = get_cookie("customer_id")
    clog.card_number = card.card_number
    clog.account_number = form.d.account_number
    clog.recharge_status = 0
    clog.recharge_time = _datetime
    db.add(clog)
    
    card.card_status = 2
    
    db.commit()
    redirect('/')
        

###############################################################################
# recharge         
###############################################################################

@app.get('/recharge')
def account_recharge(db):
    account_number = request.params.get('account_number')
    form = forms.recharge_form()
    form.recharge_card.set_value('')
    form.recharge_pwd.set_value('')
    form.account_number.set_value(account_number)  
    return render('base_form',form=form)      

@app.post('/recharge')
def account_recharge(db):
    form = forms.recharge_form()
    if not form.validates(source=request.forms):
        return render("base_form", form=form)
    if vcache.is_over(get_cookie("customer_id"),form.d.recharge_card):
         return render("base_form", form=form,msg=u"该充值卡一小时内密码输入错误超过5次，请一小时后再试")
    
    # 1 check card     
    card = db.query(models.SlcRechargerCard).filter_by(card_number=form.d.recharge_card).first()  
    check_result = check_card(card)
    if check_result['code'] > 0:
        return render('base_form',form=form,msg=check_result['data'])

    if utils.decrypt(card.card_passwd) != form.d.recharge_pwd:
        vcache.incr(get_cookie("customer_id"),form.d.recharge_card)
        errs = vcache.errs(get_cookie("customer_id"),form.d.recharge_card)
        return render('base_form',form=form,msg=u"充值卡密码错误%s次"%errs)   
        
    vcache.clear(get_cookie("customer_id"),form.d.recharge_card)
        
    # 2 check account
    account = db.query(models.SlcRadAccount).filter_by(account_number=form.d.account_number).first()
    if not account:
        return render("base_form", form=form,msg=u'没有这个账号')
    if account.member_id != get_cookie("customer_id"):
        return render("base_form", form=form,msg=u'该账号用用户不匹配')
    if account.status not in (UsrNormal,UsrExpire):
        return render("base_form", form=form,msg=u'只有正常状态的用户才能充值')
    
    # 3 check product
    user_product = db.query(models.SlcRadProduct).get(account.product_id)    
    if card.card_type == ProductCard and card.product_id != account.product_id:
        return render("base_form", form=form,msg=u'您使用的是资费卡，但资费套餐与该账号资费不匹配')
    if card.card_type == BalanceCard and user_product.product_policy not in (PPTimes,PPFlow):
        return render("base_form", form=form,msg=u'您使用的是余额卡，只能为预付费时长或预付费流量账号充值')
    
    # 4 start recharge
    accept_log = models.SlcRadAcceptLog()
    accept_log.accept_type = 'charge'
    accept_log.accept_source = 'customer'
    accept_log.account_number = form.d.account_number
    accept_log.accept_time = utils.get_currtime()
    accept_log.operator_name = "customer"
    accept_log.accept_desc = u"用户自助充值：上网账号:%s，充值卡:%s"%(form.d.account_number,form.d.recharge_card)
    db.add(accept_log)
    db.flush()
    db.refresh(accept_log) 
    
    _datetime = utils.get_currtime()
    order_fee = 0
    balance = 0
    expire_date = account.expire_date
    d_expire_date = datetime.datetime.strptime(expire_date,"%Y-%m-%d")
    # 预付费包月
    if user_product.product_policy == PPMonth:
        expire_date = utils.add_months(d_expire_date,card.months).strftime("%Y-%m-%d")
        order_fee = decimal.Decimal(user_product.fee_price) * decimal.Decimal(card.months)
        order_fee = int(order_fee.to_integral_value())
    # 买断包月,买断时长,买断流量
    if user_product.product_policy in (BOMonth,BOTimes,BOFlows):
        expire_date = utils.add_months(d_expire_date,card.months).strftime("%Y-%m-%d")
        order_fee = user_product.fee_price
    #预付费时长,预付费流量
    elif user_product.product_policy in (PPTimes,PPFlow):
        balance = card.fee_value
    
    order = models.SlcMemberOrder()
    order.order_id = utils.gen_order_id()
    order.member_id = get_cookie("customer_id")
    order.product_id = account.product_id
    order.account_number = form.d.account_number
    order.order_fee = order_fee
    order.actual_fee = card.fee_value
    order.pay_status = 1
    order.accept_id = accept_log.id
    order.order_source = 'customer'
    order.create_time = _datetime
    order.order_desc = u"用户自助充值，充值卡[ %s ]"%form.d.recharge_card
    db.add(order)
         
    account.expire_date = expire_date
    account.balance += balance
    account.time_length += card.times
    account.flow_length += card.flows
    account.status = 1
    
    db.commit()
    redirect("/") 