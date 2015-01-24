#!/usr/bin/env python
#coding:utf-8
import sys,os
sys.path.insert(0,os.path.split(__file__)[0])
sys.path.insert(0,os.path.abspath(os.path.pardir))
from bottle import Bottle
from bottle import request
from bottle import response
from bottle import redirect
from bottle import run as runserver
from bottle import static_file
from bottle import abort
from bottle import mako_template as render
from libs import sqla_plugin 
from libs.paginator import Paginator
from libs import utils
from libs.radius_attrs import radius_attrs
from twisted.python import log
from twisted.python.logfile import DailyLogFile
from hashlib import md5
from websock import websock
import bottle
import models
import forms
import datetime
import json

###############################################################################
# init                
###############################################################################

from base import *
from ops import app as ops_app
from business import app as bus_app

APP_DIR = os.path.split(__file__)[0]
print APP_DIR
app = Bottle()

def init_app():
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
    ops_app.install(sqla_pg)
    bus_app.install(sqla_pg)

    app.mount("/ops",ops_app)
    app.mount("/bus",bus_app)

    #create dir
    try:
        os.makedirs(os.path.join(APP_DIR,'static/xls'))
    except:pass


###############################################################################
# Basic handle         
###############################################################################

@app.route('/',apply=auth_opr)
def index(db):    
    online_count = db.query(models.SlcRadOnline.id).count()
    user_total = db.query(models.SlcRadAccount.account_number).filter_by(status=1).count()
    return render("index",**locals())

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
    set_cookie('login_node',opr.node_id)
    set_cookie('login_time', utils.get_currtime())
    set_cookie('login_ip', request.remote_addr)    

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = uname
    ops_log.operate_ip = request.remote_addr
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)登陆'%(uname,)
    db.add(ops_log)
    db.commit()

    return dict(code=0,msg="ok")

@app.get("/logout")
def admin_logout(db):
    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)登出'%(get_cookie("username"),)
    db.add(ops_log)    
    db.commit()
    set_cookie('username',None)
    set_cookie('login_node',None)
    set_cookie('login_time', None)
    set_cookie('login_ip', None)   
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

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)修改参数:%s'%(
        get_cookie("username"),
        json.dumps([ {c.name: getattr(p, c.name) for c in p.__table__.columns} for p in params ],ensure_ascii=False)
    )
    db.add(ops_log)

    db.commit()
    websock.update_cache("param")
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

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)修改密码'%(get_cookie("username"),)
    db.add(ops_log)

    db.commit()
    redirect("/passwd")

###############################################################################
# node manage   
###############################################################################

@app.get('/node',apply=auth_opr)
def node(db):   
    return render("sys_node_list", page_data = get_page_data(db.query(models.SlcNode)))

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

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)新增区域信息:%s'%(get_cookie("username"),serial_json(node))
    db.add(ops_log)

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

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)修改区域信息:%s'%(get_cookie("username"),serial_json(node))
    db.add(ops_log)

    db.commit()
    redirect("/node")    

@app.get('/node/delete',apply=auth_opr)
def node_delete(db):     
    node_id = request.params.get("node_id")
    if db.query(models.SlcMember.member_id).filter_by(node_id=node_id).count()>0:
        return render("error",msg=u"该节点下有用户，不允许删除")
    db.query(models.SlcNode).filter_by(id=node_id).delete()

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)删除区域信息:%s'%(get_cookie("username"),node_id)
    db.add(ops_log)

    db.commit() 
    redirect("/node")  

###############################################################################
# bas manage    
###############################################################################

@app.route('/bas',apply=auth_opr,method=['GET','POST'])
def bas(db):   
    return render("sys_bas_list", 
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
    bas.coa_port = form.d.coa_port
    db.add(bas)

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)新增BAS信息:%s'%(get_cookie("username"),serial_json(bas))
    db.add(ops_log)

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
    bas.coa_port = form.d.coa_port

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)修改BAS信息:%s'%(get_cookie("username"),serial_json(bas))
    db.add(ops_log)

    db.commit()
    websock.update_cache("bas",ip_addr=bas.ip_addr)
    redirect("/bas")    

@app.get('/bas/delete',apply=auth_opr)
def bas_delete(db):     
    bas_id = request.params.get("bas_id")
    db.query(models.SlcRadBas).filter_by(id=bas_id).delete()

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)删除BAS信息:%s'%(get_cookie("username"),bas_id)
    db.add(ops_log)

    db.commit() 
    redirect("/bas")    


###############################################################################
# product manage       
###############################################################################

@app.route('/product',apply=auth_opr,method=['GET','POST'])
def product(db):   
    node_id = request.params.get("node_id")
    _query = db.query(models.SlcRadProduct)
    if node_id:
        _query = _query.filter_by(node_id=node_id)
    return render("sys_product_list", 
        node_list=db.query(models.SlcNode),
        page_data = get_page_data(_query),node_id=node_id)

@app.get('/product/add',apply=auth_opr)
def product_add(db):  
    nodes = [ (n.id,n.node_name) for n in db.query(models.SlcNode)]
    return render("sys_product_form",form=forms.product_add_form(nodes))

@app.get('/product/detail',apply=auth_opr)
def product_detail(db):
    product_id = request.params.get("product_id")   
    product = db.query(models.SlcRadProduct).get(product_id)
    if not product:
        return render("error",msg=u"资费不存在")
    product_attrs = db.query(models.SlcRadProductAttr).filter_by(product_id=product_id)
    return render("sys_product_detail",product=product,product_attrs=product_attrs) 


@app.post('/product/add',apply=auth_opr)
def product_add_post(db): 
    nodes = [ (n.id,n.node_name) for n in db.query(models.SlcNode)]
    form=forms.product_add_form(nodes)
    if not form.validates(source=request.forms):
        return render("sys_product_form", form=form)      
    product = models.SlcRadProduct()
    product.node_id = form.d.node_id
    product.product_name = form.d.product_name
    product.product_policy = form.d.product_policy
    product.product_status = form.d.product_status
    product.fee_months = form.d.get("fee_months",0)
    product.bind_mac = form.d.bind_mac
    product.bind_vlan = form.d.bind_vlan
    product.concur_number = form.d.concur_number
    product.fee_period = form.d.fee_period
    product.fee_price = utils.yuan2fen(form.d.fee_price)
    product.input_max_limit = form.d.input_max_limit
    product.output_max_limit = form.d.output_max_limit
    _datetime = datetime.datetime.now().strftime( "%Y-%m-%d %H:%M:%S")
    product.create_time = _datetime
    product.update_time = _datetime
    db.add(product)

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)新增资费信息:%s'%(get_cookie("username"),serial_json(product))
    db.add(ops_log)

    db.commit()
    redirect("/product")

@app.get('/product/update',apply=auth_opr)
def product_update(db):  
    product_id = request.params.get("product_id")
    form=forms.product_update_form()
    product = db.query(models.SlcRadProduct).get(product_id)
    form.fill(product)
    form.product_policy_name.set_value(forms.product_policy[product.product_policy])
    form.fee_price.set_value(utils.fen2yuan(product.fee_price))
    return render("sys_product_form",form=form)

@app.post('/product/update',apply=auth_opr)
def product_add_update(db): 
    form=forms.product_update_form()
    if not form.validates(source=request.forms):
        return render("sys_product_form", form=form)
    product = db.query(models.SlcRadProduct).get(form.d.id)
    product.product_name = form.d.product_name
    product.product_status = form.d.product_status
    product.fee_months = form.d.get("fee_months",0)
    product.bind_mac = form.d.bind_mac
    product.bind_vlan = form.d.bind_vlan
    product.concur_number = form.d.concur_number
    product.fee_period = form.d.fee_period
    product.fee_price = utils.yuan2fen(form.d.fee_price)
    product.input_max_limit = form.d.input_max_limit
    product.output_max_limit = form.d.output_max_limit
    product.update_time = utils.get_currtime()

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)修改资费信息:%s'%(get_cookie("username"),serial_json(product))
    db.add(ops_log)

    db.commit()
    websock.update_cache("product",product_id=product.id)
    redirect("/product")    

@app.get('/product/delete',apply=auth_opr)
def product_delete(db):     
    product_id = request.params.get("product_id")
    if db.query(models.SlcRadAccount).filter_by(product_id=product_id).count()>0:
        return render("error",msg=u"该套餐有用户使用，不允许删除") 
    db.query(models.SlcRadProduct).filter_by(id=product_id).delete()

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)删除资费信息:%s'%(get_cookie("username"),product_id)
    db.add(ops_log)

    db.commit() 
    websock.update_cache("product",product_id=product_id)
    redirect("/product")   

@app.get('/product/attr/add',apply=auth_opr)
def product_attr_add(db): 
    product_id = request.params.get("product_id")
    if db.query(models.SlcRadProduct).filter_by(id=product_id).count()<=0:
        return render("error",msg=u"资费不存在") 
    form = forms.product_attr_add_form()
    form.product_id.set_value(product_id)
    return render("sys_pattr_form",form=form,pattrs=radius_attrs)

@app.post('/product/attr/add',apply=auth_opr)
def product_attr_add(db): 
    form = forms.product_attr_add_form()
    if not form.validates(source=request.forms):
        return render("sys_pattr_form", form=form,pattrs=radius_attrs)   
    attr = models.SlcRadProductAttr()
    attr.product_id = form.d.product_id
    attr.attr_name = form.d.attr_name
    attr.attr_value = form.d.attr_value
    attr.attr_desc = form.d.attr_desc
    db.add(attr)

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)新增资费属性信息:%s'%(get_cookie("username"),serial_json(attr))
    db.add(ops_log)

    db.commit()

    redirect("/product/detail?product_id="+form.d.product_id) 

@app.get('/product/attr/update',apply=auth_opr)
def product_attr_update(db): 
    attr_id = request.params.get("attr_id")
    attr = db.query(models.SlcRadProductAttr).get(attr_id)
    form = forms.product_attr_update_form()
    form.fill(attr)
    return render("sys_pattr_form",form=form,pattrs=radius_attrs)

@app.post('/product/attr/update',apply=auth_opr)
def product_attr_update(db): 
    form = forms.product_attr_update_form()
    if not form.validates(source=request.forms):
        return render("pattr_form", form=form,pattrs=radius_attrs)   
    attr = db.query(models.SlcRadProductAttr).get(form.d.id)
    attr.attr_name = form.d.attr_name
    attr.attr_value = form.d.attr_value
    attr.attr_desc = form.d.attr_desc

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)修改资费属性信息:%s'%(get_cookie("username"),serial_json(attr))
    db.add(ops_log)

    db.commit()
    websock.update_cache("product",product_id=form.d.product_id)
    redirect("/product/detail?product_id="+form.d.product_id) 

@app.get('/product/attr/delete',apply=auth_opr)
def product_attr_update(db): 
    attr_id = request.params.get("attr_id")
    attr = db.query(models.SlcRadProductAttr).get(attr_id)
    product_id = attr.product_id
    db.query(models.SlcRadProductAttr).filter_by(id=attr_id).delete()

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)删除资费属性信息:%s'%(get_cookie("username"),serial_json(attr))
    db.add(ops_log)

    db.commit()
    websock.update_cache("product",product_id=product_id)
    redirect("/product/detail?product_id=%s"%product_id)     

###############################################################################
# group manage      
###############################################################################

@app.route('/group',apply=auth_opr,method=['GET','POST'])
def group(db):   
    _query = db.query(models.SlcRadGroup)
    return render("sys_group_list", 
        page_data = get_page_data(_query))

   
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
    group.update_time = utils.get_currtime()
    db.add(group)

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)新增用户组信息:%s'%(get_cookie("username"),serial_json(group))
    db.add(ops_log)

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
    group.update_time = utils.get_currtime()

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)修改用户组信息:%s'%(get_cookie("username"),serial_json(group))
    db.add(ops_log)

    db.commit()
    websock.update_cache("group",group_id=group.id)
    redirect("/group")    

@app.get('/group/delete',apply=auth_opr)
def group_delete(db):     
    group_id = request.params.get("group_id")
    db.query(models.SlcRadGroup).filter_by(id=group_id).delete()

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)删除用户组信息:%s'%(get_cookie("username"),group_id)
    db.add(ops_log)

    db.commit() 
    websock.update_cache("group",group_id=group_id)
    redirect("/group")    

###############################################################################
# roster manage    
###############################################################################

@app.route('/roster',apply=auth_opr,method=['GET','POST'])
def roster(db):   
    _query = db.query(models.SlcRadRoster)
    return render("sys_roster_list", 
        page_data = get_page_data(_query))

@app.get('/roster/add',apply=auth_opr)
def roster_add(db):  
    return render("sys_roster_form",form=forms.roster_add_form())

@app.post('/roster/add',apply=auth_opr)
def roster_add_post(db): 
    form=forms.roster_add_form()
    if not form.validates(source=request.forms):
        return render("sys_roster_form", form=form)  
    if db.query(models.SlcRadRoster.id).filter_by(mac_addr=form.d.mac_addr).count()>0:
        return render("sys_roster_form", form=form,msg=u"MAC地址已经存在")     
    roster = models.SlcRadRoster()
    roster.mac_addr = form.d.mac_addr
    roster.account_number = form.d.account_number
    roster.begin_time = form.d.begin_time
    roster.end_time = form.d.end_time
    roster.roster_type = form.d.roster_type
    db.add(roster)

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)新增黑白名单信息:%s'%(get_cookie("username"),serial_json(roster))
    db.add(ops_log)

    db.commit()
    redirect("/roster")

@app.get('/roster/update',apply=auth_opr)
def roster_update(db):  
    roster_id = request.params.get("roster_id")
    form=forms.roster_update_form()
    form.fill(db.query(models.SlcRadRoster).get(roster_id))
    return render("sys_roster_form",form=form)

@app.post('/roster/update',apply=auth_opr)
def roster_add_update(db): 
    form=forms.roster_update_form()
    if not form.validates(source=request.forms):
        return render("sys_roster_form", form=form)       
    roster = db.query(models.SlcRadRoster).get(form.d.id)
    roster.mac_addr = form.d.mac_addr
    roster.account_number = form.d.account_number
    roster.begin_time = form.d.begin_time
    roster.end_time = form.d.end_time
    roster.roster_type = form.d.roster_type

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)修改黑白名单信息:%s'%(get_cookie("username"),serial_json(roster))
    db.add(ops_log)

    db.commit()
    websock.update_cache("roster",roster_id=roster.id)
    redirect("/roster")    

@app.get('/roster/delete',apply=auth_opr)
def roster_delete(db):     
    roster_id = request.params.get("roster_id")
    db.query(models.SlcRadRoster).filter_by(id=roster_id).delete()

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)删除黑白名单信息:%s'%(get_cookie("username"),roster_id)
    db.add(ops_log)

    db.commit() 
    websock.update_cache("roster",roster_id=roster_id)
    redirect("/roster")        


###############################################################################
# run server                                                                 
###############################################################################

def main():
    import argparse,json
    

    parser = argparse.ArgumentParser()
    parser.add_argument('-http','--httpport', type=int,default=0,dest='httpport',help='http port')
    parser.add_argument('-raddr','--radaddr', type=str,default=None,dest='radaddr',help='raduis address')
    parser.add_argument('-admin','--adminport', type=int,default=0,dest='adminport',help='admin port')
    parser.add_argument('-d','--debug', nargs='?',type=bool,default=False,dest='debug',help='debug')
    parser.add_argument('-c','--conf', type=str,default="../config.json",dest='conf',help='conf file')
    args =  parser.parse_args(sys.argv[1:])

    if not args.conf or not os.path.exists(args.conf):
        print 'no config file user -c or --conf cfgfile'
        return

    _config = json.loads(open(args.conf).read())
    _mysql = _config['mysql']
    _console = _config['console']

    if args.httpport:
        _console['httpport'] = args.httpport
    if args.radaddr:
        _console['radaddr'] = args.radaddr
    if args.adminport:
        _console['adminport'] = args.adminport
    if args.debug:
        _console['debug'] = bool(args.debug)

    init_context(radaddr=_console['radaddr'],adminport=_console['adminport'])
    websock.connect(_console['radaddr'],_console['adminport'])

    from sqlalchemy import create_engine
    models.engine = create_engine('mysql://%s:%s@%s:3306/%s?charset=utf8'%(
            _mysql['user'],_mysql['passwd'],_mysql['host'],_mysql['db']))

    init_app()
    
    log.startLogging(sys.stdout)    
    runserver(
        app, host='0.0.0.0', 
        port=_console['httpport'] ,
        debug=bool(_console['debug']),
        reloader=bool(_console['debug']),
        server="twisted"
    )

if __name__ == "__main__":
    main()
