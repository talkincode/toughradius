#!/usr/bin/env python
#coding:utf-8

from bottle import Bottle
from bottle import request
from bottle import response
from bottle import redirect
from bottle import run as runserver
from bottle import static_file
from bottle import abort
from bottle import MakoTemplate
from bottle import mako_template as render
from beaker.middleware import SessionMiddleware
from libs import sqla_plugin 
from libs.paginator import Paginator
from hashlib import md5
import logging
import bottle
import functools
import urllib
import models
import forms
import decimal
import datetime

###############################################################################
# init                
###############################################################################

""" define logging """
logger = logging.getLogger("admin")

app = bottle.app()
app.config.update(dict(
    port = 8080,
    secret='123321qweasd',
    page_size = 20,
))

decimal.getcontext().prec = 11
decimal.getcontext().rounding = decimal.ROUND_UP

def fen2yuan(fen):
    f = decimal.Decimal(fen)
    y = f / decimal.Decimal(100)
    return str(y.quantize(decimal.Decimal('1.00')))

def yuan2fen(yuan):
    y = decimal.Decimal(yuan)
    f = y * decimal.Decimal(100)
    return int(f.to_integral_value())

get_cookie = lambda name: request.get_cookie(name,secret=app.config['secret'])
set_cookie = lambda name,value:response.set_cookie(name,value,secret=app.config['secret'])

MakoTemplate.defaults.update(dict(
    system_name = 'ToughRADIUS Console',
    get_cookie = get_cookie,
    fen2yuan = fen2yuan,
    request = request
))

''' install plugins'''
sqla_pg = sqla_plugin.Plugin(
    models.engine, 
    models.metadata, 
    keyword='db', 
    create=False, 
    commit=False, 
    use_kwargs=False 
)
app.install(sqla_pg)

def auth_opr(func):
    @functools.wraps(func)
    def warp(*args,**kargs):
        if not get_cookie("username"):
            logger.info("admin login timeout")
            return redirect('/login')
        else:
            return func(*args,**kargs)
    return warp

def get_page_data(query):
    def _page_url(page, form_id=None):
        if form_id:return "javascript:goto_page('%s',%s);" %(form_id.strip(),page)
        request.query['page'] = page
        return request.path + '?' + urllib.urlencode(request.query)        
    page_size = app.config.get("page_size",20)
    page = int(request.params.get("page",1))
    offset = (page - 1) * page_size
    page_data = Paginator(_page_url, page, query.count(), page_size)
    page_data.result = query.limit(page_size).offset(offset)
    return page_data

###############################################################################
# Basic handle         
###############################################################################

@app.route('/',apply=auth_opr)
def index(db):    
    return render("index")

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

@app.get('/login')
def admin_login_get(db):
    return render("login")

@app.post('/login')
def admin_login_post(db):
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
    set_cookie('username',uname)
    return dict(code=0,msg="ok")

@app.get("/logout")
def admin_logout():
    request.cookies.clear()
    redirect('/login')

###############################################################################
# param config      
############################################################################### 

@app.get('/param',apply=auth_opr)
def param(db):   
    return render("base_form",form=forms.param_form(db.query(models.SlcParam)))

@app.post('/param',apply=auth_opr)
def param_update(db): 
    params = db.query(models.SlcParam)
    for param in params:
        if param.param_name in request.forms:
            _value = request.forms.get(param.param_name)
            if _value and param.param_value not in _value:
                param.param_value = _value
    db.commit()
    redirect("/param")

###############################################################################
# password update     
###############################################################################

@app.get('/passwd',apply=auth_opr)
def passwd(db):   
    form=forms.passwd_update_form()
    form.fill(operator_name=get_cookie("username"))
    return render("base_form",form=form)

@app.post('/passwd',apply=auth_opr)
def passwd_update(db):  
    form=forms.passwd_update_form() 
    if not form.validates(source=request.forms):
        return render("base_form", form=form)
    if form.d.operator_pass != form.d.operator_pass_chk:
        return render("base_form", form=form,msg=u"确认密码不一致")
    opr = db.query(models.SlcOperator).first()
    opr.operator_pass = md5(form.d.operator_pass).hexdigest()
    db.commit()
    redirect("/passwd")

###############################################################################
# node manage   
###############################################################################

@app.get('/node',apply=auth_opr)
def node(db):   
    return render("node_list", page_data = get_page_data(db.query(models.SlcNode)))

@app.get('/node/add',apply=auth_opr)
def node_add(db):  
    return render("base_form",form=forms.node_add_form())

@app.post('/node/add',apply=auth_opr)
def node_add_post(db): 
    form=forms.node_add_form()
    if not form.validates(source=request.forms):
        return render("base_form", form=form)
    node = models.SlcNode()
    node.node_name = form.d.node_name
    node.node_desc = form.d.node_desc
    db.add(node)
    db.commit()
    redirect("/node")

@app.get('/node/update',apply=auth_opr)
def node_update(db):  
    node_id = request.params.get("node_id")
    form=forms.node_update_form()
    form.fill(db.query(models.SlcNode).get(node_id))
    return render("base_form",form=form)

@app.post('/node/update',apply=auth_opr)
def node_add_update(db): 
    form=forms.node_update_form()
    if not form.validates(source=request.forms):
        return render("base_form", form=form)
    node = db.query(models.SlcNode).get(form.d.id)
    node.node_name = form.d.node_name
    node.node_desc = form.d.node_desc
    db.commit()
    redirect("/node")    

@app.get('/node/delete',apply=auth_opr)
def node_delete(db):     
    node_id = request.params.get("node_id")
    if db.query(models.SlcMember.member_id).filter_by(node_id=node_id).count()>0:
        return render("error",msg=u"该节点下有用户，不允许删除")
    db.query(models.SlcNode).filter_by(id=node_id).delete()
    db.commit() 
    redirect("/node")  

###############################################################################
# bas manage    
###############################################################################

@app.get('/bas',apply=auth_opr)
def bas(db):   
    return render("bas_list", 
        bastype = forms.bastype,
        bas_list = db.query(models.SlcRadBas))
    
@app.get('/bas/add',apply=auth_opr)
def bas_add(db):  
    return render("base_form",form=forms.bas_add_form())

@app.post('/bas/add',apply=auth_opr)
def bas_add_post(db): 
    form=forms.bas_add_form()
    if not form.validates(source=request.forms):
        return render("base_form", form=form)
    if db.query(models.SlcRadBas.id).filter_by(ip_addr=form.d.ip_addr).count()>0:
        return render("base_form", form=form,msg=u"Bas地址已经存在")        
    bas = models.SlcRadBas()
    bas.ip_addr = form.d.ip_addr
    bas.bas_name = form.d.bas_name
    bas.time_type = form.d.time_type
    bas.vendor_id = form.d.vendor_id
    bas.bas_secret = form.d.bas_secret
    db.add(bas)
    db.commit()
    redirect("/bas")

@app.get('/bas/update',apply=auth_opr)
def bas_update(db):  
    bas_id = request.params.get("bas_id")
    form=forms.bas_update_form()
    form.fill(db.query(models.SlcRadBas).get(bas_id))
    return render("base_form",form=form)

@app.post('/bas/update',apply=auth_opr)
def bas_add_update(db): 
    form=forms.bas_update_form()
    if not form.validates(source=request.forms):
        return render("base_form", form=form)
    bas = db.query(models.SlcRadBas).get(form.d.id)
    bas.bas_name = form.d.bas_name
    bas.time_type = form.d.time_type
    bas.vendor_id = form.d.vendor_id
    bas.bas_secret = form.d.bas_secret
    db.commit()
    redirect("/bas")    

@app.get('/bas/delete',apply=auth_opr)
def bas_delete(db):     
    bas_id = request.params.get("bas_id")
    db.query(models.SlcRadBas).filter_by(id=bas_id).delete()
    db.commit() 
    redirect("/bas")    


###############################################################################
# product manage       
###############################################################################

@app.get('/product',apply=auth_opr)
def product(db):   
    return render("product_list", page_data = get_page_data(db.query(models.SlcRadProduct)))

@app.get('/product/add',apply=auth_opr)
def product_add(db):  
    return render("base_form",form=forms.product_add_form())

@app.post('/product/add',apply=auth_opr)
def product_add_post(db): 
    form=forms.product_add_form()
    if not form.validates(source=request.forms):
        return render("base_form", form=form)      
    product = models.SlcRadProduct()
    product.product_name = form.d.product_name
    product.product_policy = form.d.product_policy
    product.product_status = form.d.product_status
    product.domain_name = form.d.domain_name
    product.bandwidth_code = form.d.bandwidth_code
    product.bind_mac = form.d.bind_mac
    product.bind_vlan = form.d.bind_vlan
    product.concur_number = form.d.concur_number
    product.fee_period = form.d.fee_period
    product.fee_price = yuan2fen(form.d.fee_price)
    product.input_max_limit = form.d.input_max_limit
    product.output_max_limit = form.d.output_max_limit
    product.input_rate_code = form.d.input_rate_code
    product.output_rate_code = form.d.output_rate_code
    _datetime = datetime.datetime.now().strftime( "%Y-%m-%d %H:%M:%S")
    product.create_time = _datetime
    product.update_time = _datetime
    db.add(product)
    db.commit()
    redirect("/product")

@app.get('/product/update',apply=auth_opr)
def product_update(db):  
    product_id = request.params.get("product_id")
    form=forms.product_update_form()
    product = db.query(models.SlcRadProduct).get(product_id)
    form.fill(product)
    form.product_policy_name.set_value(forms.product_policy[product.product_policy])
    form.fee_price.set_value(fen2yuan(product.fee_price))
    return render("base_form",form=form)

@app.post('/product/update',apply=auth_opr)
def product_add_update(db): 
    form=forms.product_update_form()
    if not form.validates(source=request.forms):
        return render("base_form", form=form)
    product = db.query(models.SlcRadProduct).get(form.d.id)
    product.product_name = form.d.product_name
    product.product_status = form.d.product_status
    product.domain_name = form.d.domain_name
    product.bandwidth_code = form.d.bandwidth_code
    product.bind_mac = form.d.bind_mac
    product.bind_vlan = form.d.bind_vlan
    product.concur_number = form.d.concur_number
    product.fee_period = form.d.fee_period
    product.fee_price = yuan2fen(form.d.fee_price)
    product.input_max_limit = form.d.input_max_limit
    product.output_max_limit = form.d.output_max_limit
    product.input_rate_code = form.d.input_rate_code
    product.output_rate_code = form.d.output_rate_code
    product.update_time = datetime.datetime.now().strftime( "%Y-%m-%d %H:%M:%S")
    db.commit()
    redirect("/product")    

@app.get('/product/delete',apply=auth_opr)
def product_delete(db):     
    product_id = request.params.get("product_id")
    if db.query(models.SlcRadAccount).filter_by(product_id=product_id).count()>0:
        return render("error",msg=u"该套餐有用户使用，不允许删除") 
    db.query(models.SlcRadProduct).filter_by(id=product_id).delete()
    db.commit() 
    redirect("/product")   

###############################################################################
# group manage      
###############################################################################

@app.get('/group',apply=auth_opr)
def group(db):   
    return render("group_list", page_data = get_page_data(db.query(models.SlcRadGroup)))

   
@app.get('/group/add',apply=auth_opr)
def group_add(db):  
    return render("base_form",form=forms.group_add_form())

@app.post('/group/add',apply=auth_opr)
def group_add_post(db): 
    form=forms.group_add_form()
    if not form.validates(source=request.forms):
        return render("base_form", form=form)    
    group = models.SlcRadGroup()
    group.group_name = form.d.group_name
    group.group_desc = form.d.group_desc
    group.bind_mac = form.d.bind_mac
    group.bind_vlan = form.d.bind_vlan
    group.concur_number = form.d.concur_number
    group.update_time = datetime.datetime.now().strftime( "%Y-%m-%d %H:%M:%S")
    db.add(group)
    db.commit()
    redirect("/group")

@app.get('/group/update',apply=auth_opr)
def group_update(db):  
    group_id = request.params.get("group_id")
    form=forms.group_update_form()
    form.fill(db.query(models.SlcRadGroup).get(group_id))
    return render("base_form",form=form)

@app.post('/group/update',apply=auth_opr)
def group_add_update(db): 
    form=forms.group_update_form()
    if not form.validates(source=request.forms):
        return render("base_form", form=form)
    group = db.query(models.SlcRadGroup).get(form.d.id)
    group.group_name = form.d.group_name
    group.group_desc = form.d.group_desc
    group.bind_mac = form.d.bind_mac
    group.bind_vlan = form.d.bind_vlan
    group.concur_number = form.d.concur_number
    group.update_time = datetime.datetime.now().strftime( "%Y-%m-%d %H:%M:%S")
    db.commit()
    redirect("/group")    

@app.get('/group/delete',apply=auth_opr)
def group_delete(db):     
    group_id = request.params.get("group_id")
    db.query(models.SlcRadGroup).filter_by(id=group_id).delete()
    db.commit() 
    redirect("/group")    

###############################################################################
# roster manage    
###############################################################################

@app.get('/roster',apply=auth_opr)
def roster(db):   
    return render("roster_list", page_data = get_page_data(db.query(models.SlcRadRoster)))

###############################################################################
# user manage        
###############################################################################
                   
@app.route('/user',apply=auth_opr,method=['GET','POST'])
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
        _query = _query.filter(models.SlcRadAccount.product_id)
    if user_name:
        _query = _query.filter(models.SlcRadAccount.account_number.like('%'+user_name+'%'))
    if status:
        _query = _query.filter(models.SlcRadAccount.status == status)
        
    return render("user_list", page_data = get_page_data(_query),
                   node_list=db.query(models.SlcNode), 
                   product_list=db.query(models.SlcRadProduct),**request.params)

@app.get('/user/trace',apply=auth_opr)
def user_trace(db):   
    return render("user_trace", bas_list=db.query(models.SlcRadBas))
    
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

    return render("online_list", page_data = get_page_data(_query),
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

    return render("ticket_list", page_data = get_page_data(_query),
               node_list=db.query(models.SlcNode),**request.params)

    
###############################################################################
# run server                                                                 
###############################################################################

if __name__ == "__main__":
    runserver(app, host='0.0.0.0', port=8080 ,debug=True,reloader=True)

