#!/usr/bin/env python
#coding:utf-8
import sys,os
sys.path.insert(0,os.path.split(__file__)[0])
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
from hashlib import md5
import bottle
import models
import forms
import datetime


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
    set_cookie('login_time', utils.get_currtime())
    set_cookie('login_ip', request.remote_addr)    
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

@app.get('/product/detail',apply=auth_opr)
def product_detail(db):
    product_id = request.params.get("product_id")   
    product = db.query(models.SlcRadProduct).get(product_id)
    if not product:
        return render("error",msg=u"资费不存在")
    product_attrs = db.query(models.SlcRadProductAttr).filter_by(product_id=product_id)
    return render("product_detail",product=product,product_attrs=product_attrs) 


@app.post('/product/add',apply=auth_opr)
def product_add_post(db): 
    form=forms.product_add_form()
    if not form.validates(source=request.forms):
        return render("base_form", form=form)      
    product = models.SlcRadProduct()
    product.product_name = form.d.product_name
    product.product_policy = form.d.product_policy
    product.product_status = form.d.product_status
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
    return render("base_form",form=form)

@app.post('/product/update',apply=auth_opr)
def product_add_update(db): 
    form=forms.product_update_form()
    if not form.validates(source=request.forms):
        return render("base_form", form=form)
    product = db.query(models.SlcRadProduct).get(form.d.id)
    product.product_name = form.d.product_name
    product.product_status = form.d.product_status
    product.bind_mac = form.d.bind_mac
    product.bind_vlan = form.d.bind_vlan
    product.concur_number = form.d.concur_number
    product.fee_period = form.d.fee_period
    product.fee_price = utils.yuan2fen(form.d.fee_price)
    product.input_max_limit = form.d.input_max_limit
    product.output_max_limit = form.d.output_max_limit
    product.update_time = utils.get_currtime()
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

@app.get('/product/attr/add',apply=auth_opr)
def product_attr_add(db): 
    product_id = request.params.get("product_id")
    if db.query(models.SlcRadProduct).filter_by(id=product_id).count()<=0:
        return render("error",msg=u"资费不存在") 
    form = forms.product_attr_add_form()
    form.product_id.set_value(product_id)
    return render("base_form",form=form)

@app.post('/product/attr/add',apply=auth_opr)
def product_attr_add(db): 
    form = forms.product_attr_add_form()
    if not form.validates(source=request.forms):
        return render("base_form", form=form)   
    attr = models.SlcRadProductAttr()
    attr.product_id = form.d.product_id
    attr.attr_name = form.d.attr_name
    attr.attr_value = form.d.attr_value
    attr.attr_desc = form.d.attr_desc
    db.add(attr)
    db.commit()
    redirect("/product/detail?product_id="+form.d.product_id) 

@app.get('/product/attr/update',apply=auth_opr)
def product_attr_update(db): 
    attr_id = request.params.get("attr_id")
    attr = db.query(models.SlcRadProductAttr).get(attr_id)
    form = forms.product_attr_update_form()
    form.fill(attr)
    return render("base_form",form=form)

@app.post('/product/attr/update',apply=auth_opr)
def product_attr_update(db): 
    form = forms.product_attr_update_form()
    if not form.validates(source=request.forms):
        return render("base_form", form=form)   
    attr = db.query(models.SlcRadProductAttr).get(form.d.id)
    attr.attr_name = form.d.attr_name
    attr.attr_value = form.d.attr_value
    attr.attr_desc = form.d.attr_desc
    db.commit()
    redirect("/product/detail?product_id="+form.d.product_id) 

@app.get('/product/attr/delete',apply=auth_opr)
def product_attr_update(db): 
    attr_id = request.params.get("attr_id")
    attr = db.query(models.SlcRadProductAttr).get(attr_id)
    product_id = attr.product_id
    db.query(models.SlcRadProductAttr).filter_by(id=attr_id).delete()
    db.commit()
    redirect("/product/detail?product_id=%s"%product_id)     

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
    group.update_time = utils.get_currtime()
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
    group.update_time = utils.get_currtime()
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

@app.get('/roster/add',apply=auth_opr)
def roster_add(db):  
    return render("roster_form",form=forms.roster_add_form())

@app.post('/roster/add',apply=auth_opr)
def roster_add_post(db): 
    form=forms.roster_add_form()
    if not form.validates(source=request.forms):
        return render("roster_form", form=form)  
    if db.query(models.SlcRadRoster.id).filter_by(mac_addr=form.d.mac_addr).count()>0:
        return render("roster_form", form=form,msg=u"MAC地址已经存在")     
    roster = models.SlcRadRoster()
    roster.mac_addr = form.d.mac_addr
    roster.account_number = form.d.account_number
    roster.begin_time = form.d.begin_time
    roster.end_time = form.d.end_time
    roster.roster_type = form.d.roster_type
    db.add(roster)
    db.commit()
    redirect("/roster")

@app.get('/roster/update',apply=auth_opr)
def roster_update(db):  
    roster_id = request.params.get("roster_id")
    form=forms.roster_update_form()
    form.fill(db.query(models.SlcRadRoster).get(roster_id))
    return render("roster_form",form=form)

@app.post('/roster/update',apply=auth_opr)
def roster_add_update(db): 
    form=forms.roster_update_form()
    if not form.validates(source=request.forms):
        return render("roster_form", form=form)       
    roster = db.query(models.SlcRadRoster).get(form.d.id)
    roster.mac_addr = form.d.mac_addr
    roster.account_number = form.d.account_number
    roster.begin_time = form.d.begin_time
    roster.end_time = form.d.end_time
    roster.roster_type = form.d.roster_type
    db.commit()
    redirect("/roster")    

@app.get('/roster/delete',apply=auth_opr)
def roster_delete(db):     
    roster_id = request.params.get("roster_id")
    db.query(models.SlcRadRoster).filter_by(id=roster_id).delete()
    db.commit() 
    redirect("/roster")        



    
###############################################################################
# run server                                                                 
###############################################################################

def main():
    import argparse,json
    parser = argparse.ArgumentParser()
    parser.add_argument('-http','--httpport', type=int,default=1816,dest='httpport',help='http port')
    parser.add_argument('-admin','--adminport', type=int,default=1815,dest='adminport',help='admin port')
    parser.add_argument('-d','--debug', type=int,default=1815,dest='debug',help='debug')
    parser.add_argument('-c','--conf', type=str,default=None,dest='conf',help='conf file')
    args =  parser.parse_args(sys.argv[1:])
    init_context(adminport=args.adminport)
    if args.conf:
        from sqlalchemy import create_engine
        with open(args.conf) as cf:
            _mysql = json.loads(cf.read())['mysql']
            models.engine = create_engine(
                'mysql://%s:%s@%s:3306/%s?charset=utf8'%(
                    _mysql['user'],_mysql['passwd'],_mysql['host'],_mysql['db']
        )
    )
    init_app()
    runserver(app, host='0.0.0.0', port=args.httpport ,debug=bool(args.debug),reloader=bool(args.debug),server="cherrypy")

if __name__ == "__main__":
    main()
