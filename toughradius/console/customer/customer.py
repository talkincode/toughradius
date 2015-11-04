#!/usr/bin/env python
#coding:utf-8

import os
from twisted.internet import reactor
from bottle import Bottle
from bottle import abort
from sqlalchemy.sql import exists
from toughradius.console.base import *
from toughradius.console.libs import utils
from toughradius.console.libs.smail import mail
from toughradius.console.websock import websock
from toughradius.console import models
from toughradius.console.customer import forms

app = Bottle()

###############################################################################
# Basic handle         
###############################################################################

@app.route('/static/:path#.+#')
def route_static(path,render):
    static_path = os.path.join(os.path.split(os.path.split(__file__)[0])[0],'static')
    return static_file(path, root=static_path)

###############################################################################
# index handle         
###############################################################################
@cache.cache('customer_index_get_data',expire=180)  
def get_data(db,member_name):
    member = db.query(models.SlcMember).filter_by(member_name=member_name).first()
    if not member:
        return None,None,None
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
    ).order_by(models.SlcMemberOrder.create_time.desc())
    return member,accounts,orders
        
@app.get('/',apply=auth_cus)
def customer_index(db, render):
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



@app.get("/active/<code>")
def active_user(db,code, render):
    member = db.query(models.SlcMember).filter(
        models.SlcMember.active_code == code,
    ).first()

    if not member:
        return render("error",msg=u"无效的激活码")

    if member.email_active == 1:
        return render("error",msg=u"用户已经激活")

    member.email_active  = 1
    db.commit()
    
    return render("msg",msg=u"恭喜您, 激活成功, 请登录系统")

###############################################################################
# user join        
###############################################################################

@app.get('/join')
def member_join_get(db, render):
    nodes = [ (n.id,n.node_name) for n in db.query(models.SlcNode)]
    form = forms.member_join_form(nodes)
    return render("join",form=form)
    
@app.post('/join')
def member_join_post(db, render):
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
    member.email_active = 0
    member.mobile_active = 0
    member.active_code = utils.get_uuid()
    db.add(member) 
    db.commit()
    
    topic = u'%s,请验证您在%s注册的电子邮件地址'%(member.realname,get_param_value(db,"customer_system_name"))
    ctx = dict(
        username = member.realname,
        customer_name = get_param_value(db,"customer_system_name"),
        customer_url = get_param_value(db,"customer_system_url"),
        active_code = member.active_code
    )
    reactor.callInThread(mail.sendmail,member.email,topic,render("mail",**ctx))
    return render('msg',msg=u"新用户注册成功,请注意查收您的注册邮箱，及时激活账户")
    
###############################################################################
# user password reset
###############################################################################
@app.get('/password/mail')
def password_reset_mail(db, render):
    form = forms.password_mail_form()
    return render("base_form",form=form)

@app.post('/password/mail')
def password_reset_mail(db, render):
    form = forms.password_mail_form()
    if not form.validates(source=request.forms):
        return render("base_form", form=form)

    last_send = get_cookie("last_send_active") 
    if last_send:
        sec = int(time.time()) - int(float(last_send))
        if sec < 60:
            return render("error",msg=u"每间隔60秒才能提交一次,还需等待%s秒"% int(60-sec))
    set_cookie("last_send_active", str(time.time()))
    member_name = request.params.get("member_name")
    member = db.query(models.SlcMember).filter_by(member_name=member_name).first()
    if not member:
        return render("error",msg=u"用户不存在")
    try:
        member.active_code = utils.get_uuid()
        db.commit()
        topic = u'%s,请重置您在%s的密码'%(member.realname,get_param_value(db,"customer_system_name"))
        ctx = dict(
            username = member.realname,
            customer_name = get_param_value(db,"customer_system_name"),
            customer_url = get_param_value(db,"customer_system_url"),
            active_code = member.active_code
        )
        reactor.callInThread(mail.sendmail,member.email,topic,render("pwdmail",**ctx))
        return render("msg",msg=u"激活邮件已经发送置您的邮箱 *****%s,请注意查收。"%member.email[member.email.find('@'):])  
    except :
        return render('error',msg=u"激活邮件发送失败,请稍后再试")  

@app.get("/password/reset/<code>")
def password_reset(db,code, render):
    form = forms.password_reset_form() 
    form.active_code.set_value(code)
    return render("base_form",form=form)

@app.post("/password/reset")
def password_reset(db, render):
    form = forms.password_reset_form()
    if not form.validates(source=request.forms):
        return render("base_form", form=form)
        
    member = db.query(models.SlcMember).filter(
        models.SlcMember.active_code == form.d.active_code,
    ).first()
    
    if not member:
        return render("error",msg=u"无效的验证码")
        
    if form.d.new_password != form.d.new_password2:
        return render("base_form", form=form,msg=u'确认新密码不匹配')
    
    member.password =  md5(form.d.new_password.encode()).hexdigest()
    member.active_code = utils.get_uuid()
    db.commit()
    return render("msg",msg=u"密码重置成功，请重新登录系统。")
    
###############################################################################
# user update
###############################################################################

@app.get('/user/update',apply=auth_cus)
def member_update(db, render):
    member = db.query(models.SlcMember).get(get_cookie("customer_id"))
    form = forms.member_update_form()
    form.fill(member)
    form.new_password.set_value("")
    return render("base_form",form=form)

@app.post('/user/update',apply=auth_cus)
def member_update(db, render):
    form=forms.member_update_form()
    if not form.validates(source=request.forms):
        return render("base_form", form=form)

    member = db.query(models.SlcMember).get(get_cookie("customer_id"))
    oldemail = member.email
    member.realname = form.d.realname
    if form.d.new_password:
        member.password =  md5(form.d.new_password.encode()).hexdigest()
    member.email = form.d.email
    member.address = form.d.address
    
    if oldemail != member.email:
        member.email_active = 0
        member.active_code = utils.get_uuid()
    
    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'(%s)修改用户信息:%s'%(get_cookie("username"),member.member_name)
    db.add(ops_log)

    db.commit()
    
    if member.email and (oldemail != member.email):
        topic = u'%s,请验证您在%s注册的电子邮件地址'%(member.realname,get_param_value(db,"customer_system_name"))
        ctx = dict(
            username = member.realname,
            customer_name = get_param_value(db,"customer_system_name"),
            customer_url = get_param_value(db,"customer_system_url"),
            active_code = member.active_code
        )
        reactor.callInThread(mail.sendmail,member.email,topic,render("mail",**ctx))
        return render("msg",msg=u"您修改了email地址，系统已发送激活邮件，请重新激活绑定")
    else:
        return redirect("/")
   
@app.post('/email/reactive',apply=auth_cus)
def email_reactive(db, render):
    last_send = get_cookie("last_send_active") 
    if last_send:
        sec = int(time.time()) - int(float(last_send))
        if sec < 60:
            return dict(code=1,msg=u"每间隔60秒才能发送一次,还需等待%s秒"% int(60-sec))

    set_cookie("last_send_active", str(time.time()))
    member = db.query(models.SlcMember).get(get_cookie("customer_id"))
    try:
        topic = u'%s,请验证您在%s注册的电子邮件地址'%(member.realname,get_param_value(db,"customer_system_name"))
        ctx = dict(
            username = member.realname,
            customer_name = get_param_value(db,"customer_system_name"),
            customer_url = get_param_value(db,"customer_system_url"),
            active_code = member.active_code
        )
        reactor.callInThread(mail.sendmail,member.email,topic,render("mail",**ctx))
        return dict(code=0,msg=u"激活邮件已经发送")  
    except :
        return dict(code=0,msg=u"激活邮件发送失败,请稍后再试")   
 
###############################################################################
# account query        
###############################################################################
   
@app.get('/account/detail',apply=auth_cus)
def account_detail(db, render):
    account_number = request.params.get('account_number')  
    user = db.query(
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
def product_list(db, render):
    return render("product_list",products=db.query(models.SlcRadProduct).filter_by(
        product_status = 0
    ))
    
###############################################################################
# billing query        
###############################################################################
    
@app.route('/billing',apply=auth_cus,method=['GET','POST'])
def billing_query(db, render):
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
# password update    
###############################################################################    
    
@app.get('/password/update',apply=auth_cus)    
def password_update_get(db, render):
    form = forms.password_update_form()
    account_number = request.params.get('account_number')      
    form.account_number.set_value(account_number)
    return render("base_form",form=form)
    
@app.post('/password/update',apply=auth_cus)    
def password_update_post(db, render):
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
def portal_auth(db, render):
    user = request.params.get("user")
    token = request.params.get("token")
    if not user:return abort(403,'user is empty')
    if not token:return abort(403,'token is empty')
    account = db.query(models.SlcRadAccount).filter_by(
        account_number=user
    ).first()
    if not account:
        return abort(403,'user not exists')
    secret = get_param_value(db,"portal_secret")
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
        

