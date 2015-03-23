#!/usr/bin/env python
#coding:utf-8
from __future__ import unicode_literals
import sys,os
from bottle import Bottle
from bottle import request
from bottle import response
from bottle import redirect
from bottle import MakoTemplate
from bottle import static_file
from bottle import abort
from beaker.cache import cache_managers
from toughradius.console.libs.paginator import Paginator
from toughradius.console.libs import utils
from toughradius.console.websock import websock
from toughradius.console import models
from toughradius.console.base import *
from toughradius.console.admin import forms
from hashlib import md5
import bottle
import datetime
import json
import functools

app = Bottle()
render = functools.partial(Render.render_app,app)

##############################################################################
# test handle
##############################################################################
@app.route('/test',apply=auth_opr)
def index(db):    
    form = forms.param_form()
    fparam = {}
    for p in db.query(models.SlcParam):
        fparam[p.param_name] = p.param_value
    form.fill(fparam)
    return render("base_form",form=form)

@app.get('/test/pid',apply=auth_opr)
def product_id(db):
    name = request.params.get("name")   
    product = db.query(models.SlcRadProduct).filter(
        models.SlcRadProduct.product_name == name
    ).first()
    return dict(pid=product.id)
    
@app.get('/test/mid',apply=auth_opr)
def member_id(db):
    name = request.params.get("name")   
    member = db.query(models.SlcMember).filter(
        models.SlcMember.member_name == name
    ).first()
    return dict(mid=member.member_id)
    
@app.route('/mksign',apply=auth_opr)
def index(db):    
    sign_args = request.params.get('sign_args')
    return dict(code=0,sign=utils.mk_sign(sign_args.strip().split(',')))
    
@app.post('/encrypt',apply=auth_opr)
def encrypt_data(db):    
    msg_data = request.params.get('data')
    return dict(code=0,data=utils.encrypt(msg_data))
    
@app.post('/decrypt',apply=auth_opr)
def decrypt_data(db):    
    msg_data = request.params.get('data')
    return dict(code=0,data=utils.decrypt(msg_data))
    
@app.get('/logquery/:name',apply=auth_opr)
def logquery(db,name):   
    def _query(logfile):
        if os.path.exists(logfile):
            with open(logfile) as f:
                f.seek(0,2)
                if f.tell() > 32*1024:
                    f.seek(f.tell()-32*1024)
                else:
                    f.seek(0)
                return f.read().replace('\n','<br>')
    if '%s.logfile'%name in app.config:
        logfile = app.config['%s.logfile'%name]
        return render("sys_logquery",msg=_query(logfile),title="%s logging"%name)
    else:
        return render("sys_logquery",msg="logfile not exists",title="%s logging"%name)
        
permit.add_route("/logquery/radiusd",u"radius系统日志查看",u"系统管理",is_menu=False,order=0.001,is_open=False)
permit.add_route("/logquery/admin",u"管理系统日志查看",u"系统管理",is_menu=False,order=0.002,is_open=False)
permit.add_route("/logquery/customer",u"自助系统日志查看",u"系统管理",is_menu=False,order=0.003,is_open=False)

@app.route('/backup',apply=auth_opr)
def backup(db): 
    backup_path = app.config.get('database.backup_path','/tmp/data')   
    flist = os.listdir(backup_path)
    flist.sort(reverse=True)
    return render("sys_backup_db",backups=flist[:30],backup_path=backup_path)
    
@app.route('/backup/dump',apply=auth_opr)
def backup_dump(db):   
    from toughradius.tools.backup import dumpdb
    from toughradius.tools.config import find_config
    backup_path = app.config.get('database.backup_path','/tmp/data')  
    backup_file = "toughradius_db_%s.json.gz"%utils.gen_backep_id()
    try:
        dumpdb(find_config(),os.path.join(backup_path,backup_file))
        return dict(code=0,msg="backup done!")
    except Exception as err:
        return dict(code=1,msg="backup fail! %s"%(err))
    

@app.post('/backup/restore',apply=auth_opr)
def backup_restore(db):   
    from toughradius.tools.backup import dumpdb,restoredb
    from toughradius.tools.config import find_config
    backup_path = app.config.get('database.backup_path','/tmp/data')  
    backup_file = "toughradius_db_%s.before_restore.json.gz"%utils.gen_backep_id()
    rebakfs = request.params.get("bakfs")
    try:
        dumpdb(find_config(),os.path.join(backup_path,backup_file))
        restoredb(find_config(),os.path.join(backup_path,rebakfs))
        return dict(code=0,msg="restore done!")
    except Exception as err:
        return dict(code=1,msg="restore fail! %s"%(err)) 
        
@app.post('/backup/delete',apply=auth_opr)
def backup_delete(db):   
    backup_path = app.config.get('database.backup_path','/tmp/data')  
    bakfs = request.params.get("bakfs")
    try:
        os.remove(os.path.join(backup_path,bakfs))
        return dict(code=0,msg="delete done!")
    except Exception as err:
        return dict(code=1,msg="delete fail! %s"%(err)) 
    

@app.route('/backup/download/:path#.+#',apply=auth_opr)
def backup_download(path): 
    backup_path = app.config.get('database.backup_path','/tmp/data')  
    return static_file(path, root=backup_path,download=True,mimetype="application/x-gzip")


permit.add_route("/backup",u"备份管理",u"系统管理",is_menu=False,order=0.004,is_open=False)

###############################################################################
# Basic handle         
###############################################################################

@app.route('/',apply=auth_opr)
def index(db):    
    online_count = db.query(models.SlcRadOnline.id).count()
    user_total = db.query(models.SlcRadAccount.account_number).filter_by(status=1).count()
    return render("index",**locals())

@app.route('/static/:path#.+#')
def route_static(path):
    static_path = os.path.join(os.path.split(os.path.split(__file__)[0])[0],'static')
    return static_file(path, root=static_path)
    
###############################################################################
# update all cache      
###############################################################################    
@app.get('/cache/clean')
def clear_cache():
    def cbk(resp):
        print 'cbk',resp
    bottle.TEMPLATES.clear()
    for _cache in cache_managers.values():
        _cache.clear()
    websock.update_cache("all",callback=cbk)
    return dict(code=0,msg=u"已刷新缓存")
    
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
    if opr.operator_status == 1:return dict(code=1,msg=u"该操作员账号已被停用")
    set_cookie('username',uname)
    set_cookie('opr_type',opr.operator_type)
    set_cookie('login_time', utils.get_currtime())
    set_cookie('login_ip', request.remote_addr)  
    
    if opr.operator_type > 0:
        permit.unbind_opr(uname)
        for rule in db.query(models.SlcOperatorRule).filter_by(operator_name=uname):
            permit.bind_opr(rule.operator_name,rule.rule_path)  

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
    if get_cookie('opt_type') > 0:
        permit.unbind_opr(get_cookie("username"))
    set_cookie('username',None)
    set_cookie('login_time', None)
    set_cookie('opr_type',None)
    set_cookie('login_ip', None)   
    request.cookies.clear()
    redirect('/login')

###############################################################################
# param config      
############################################################################### 

@app.get('/param',apply=auth_opr)
def param(db):   
    form = forms.param_form()
    fparam = {}
    for p in db.query(models.SlcParam):
        fparam[p.param_name] = p.param_value
    form.fill(fparam)
    return render("sys_param",form=form)

@app.post('/param/update',apply=auth_opr)
def param_update(db): 
    params = db.query(models.SlcParam)
    for param in params:
        if param.param_name in request.forms:
            _value = request.forms.get(param.param_name)
            if _value: 
                param.param_value = _value  
                
    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)修改参数'%(get_cookie("username"))
    db.add(ops_log)
    db.commit()
    
    websock.reconnect(
        request.forms.get('radiusd_address'),
        request.forms.get('radiusd_admin_port'),
    )
        
    is_debug = request.forms.get('is_debug')
    bottle.debug(is_debug == '1')
    
    websock.update_cache("is_debug",is_debug=is_debug)
    websock.update_cache("reject_delay",reject_delay=request.forms.get('reject_delay'))
    websock.update_cache("param")
    redirect("/param")
    
permit.add_route("/param",u"系统参数管理",u"系统管理",is_menu=True,order=0)
permit.add_route("/param/update",u"系统参数修改",u"系统管理",is_menu=False,order=0,is_open=False)

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
    opr = db.query(models.SlcOperator).filter_by(operator_name=form.d.operator_name).first()
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
    
permit.add_route("/node",u"区域信息管理",u"系统管理",is_menu=True,order=1)    

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
    ops_log.operate_desc = u'操作员(%s)新增区域信息:%s'%(get_cookie("username"),node.node_name)
    db.add(ops_log)

    db.commit()
    redirect("/node")
    
permit.add_route("/node/add",u"新增区域",u"系统管理",order=1.01,is_open=False)

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
    ops_log.operate_desc = u'操作员(%s)修改区域信息:%s'%(get_cookie("username"),node.node_name)
    db.add(ops_log)

    db.commit()
    redirect("/node")    
    
permit.add_route("/node/update",u"修改区域",u"系统管理",order=1.02,is_open=False)

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
    
permit.add_route("/node/delete",u"删除区域",u"系统管理",order=1.03,is_open=False)

###############################################################################
# bas manage    
###############################################################################

@app.route('/bas',apply=auth_opr,method=['GET','POST'])
def bas(db):   
    return render("sys_bas_list", 
        bastype = forms.bastype,
        bas_list = db.query(models.SlcRadBas))
        
permit.add_route("/bas",u"BAS信息管理",u"系统管理",is_menu=True,order=2)
    
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
    ops_log.operate_desc = u'操作员(%s)新增BAS信息:%s'%(get_cookie("username"),bas.ip_addr)
    db.add(ops_log)

    db.commit()
    redirect("/bas")
    
permit.add_route("/bas/add",u"新增BAS",u"系统管理",order=2.01,is_open=False)

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
    ops_log.operate_desc = u'操作员(%s)修改BAS信息:%s'%(get_cookie("username"),bas.ip_addr)
    db.add(ops_log)

    db.commit()
    websock.update_cache("bas",ip_addr=bas.ip_addr)
    redirect("/bas")    
    
permit.add_route("/bas/update",u"修改BAS",u"系统管理",order=2.02,is_open=False)

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

permit.add_route("/bas/delete",u"删除BAS",u"系统管理",order=2.03,is_open=False)


###############################################################################
# opr manage    
###############################################################################

@app.route('/opr',apply=auth_opr,method=['GET','POST'])
def opr(db):   
    return render("sys_opr_list", 
        oprtype = forms.opr_type,
        oprstatus = forms.opr_status_dict,
        opr_list = db.query(models.SlcOperator))
        
permit.add_route("/opr",u"操作员管理",u"系统管理",is_menu=True,order=3,is_open=False)
    
@app.get('/opr/add',apply=auth_opr)
def opr_add(db):  
    nodes = [ (n.node_name,n.node_desc) for n in db.query(models.SlcNode)]
    form=forms.opr_add_form(nodes)
    return render("sys_opr_form",form=form,rules=[])

@app.post('/opr/add',apply=auth_opr)
def opr_add_post(db): 
    nodes = [ (n.node_name,n.node_desc) for n in db.query(models.SlcNode)]
    form=forms.opr_add_form(nodes)
    if not form.validates(source=request.forms):
        return render("sys_opr_form", form=form,rules=[])
    if db.query(models.SlcOperator.id).filter_by(operator_name=form.d.operator_name).count()>0:
        return render("sys_opr_form", form=form,rules=[],msg=u"操作员已经存在")   
        
    opr = models.SlcOperator()
    opr.operator_name = form.d.operator_name
    opr.operator_type = 1
    opr.operator_pass = md5(form.d.operator_pass).hexdigest()
    opr.operator_desc = form.d.operator_desc
    opr.operator_status = form.d.operator_status
    db.add(opr)
    
    for node in request.params.getall("operator_nodes"):
        onode = models.SlcOperatorNodes()
        onode.operator_name = form.d.operator_name
        onode.node_name = node
        db.add(onode)
        
    for path in request.params.getall("rule_item"):
        item = permit.get_route(path)
        if not item:continue
        rule = models.SlcOperatorRule()
        rule.operator_name = opr.operator_name
        rule.rule_name = item['name']
        rule.rule_path = item['path']
        rule.rule_category = item['category']
        db.add(rule)

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)新增操作员信息:%s'%(get_cookie("username"),opr.operator_name)
    db.add(ops_log)

    db.commit()
    redirect("/opr")
    
permit.add_route("/opr/add",u"新增操作员",u"系统管理",order=3.01,is_open=False)

@app.get('/opr/update',apply=auth_opr)
def opr_update(db):  
    opr_id = request.params.get("opr_id")
    opr = db.query(models.SlcOperator).get(opr_id)
    nodes = [ (n.node_name,n.node_desc) for n in db.query(models.SlcNode)]
    form=forms.opr_update_form(nodes)
    form.fill(opr)
    form.operator_pass.set_value('')
    onodes = db.query(models.SlcOperatorNodes).filter_by(operator_name=form.d.operator_name)
    form.operator_nodes.set_value([ond.node_name for ond in onodes])
    rules = db.query(models.SlcOperatorRule.rule_path).filter_by(operator_name=opr.operator_name)
    rules = [r[0] for r in rules]
    return render("sys_opr_form",form=form,rules=rules)

@app.post('/opr/update',apply=auth_opr)
def opr_add_update(db): 
    nodes = [ (n.node_name,n.node_desc) for n in db.query(models.SlcNode)]
    form=forms.opr_update_form(nodes)
    if not form.validates(source=request.forms):
        rules = db.query(models.SlcOperatorRule.rule_path).filter_by(operator_name=opr.operator_name)
        rules = [r[0] for r in rules]
        return render("sys_opr_form", form=form,rules=rules)
    opr = db.query(models.SlcOperator).get(form.d.id)
    
    if form.d.operator_pass:
        opr.operator_pass = md5(form.d.operator_pass).hexdigest()
    opr.operator_desc = form.d.operator_desc
    opr.operator_status = form.d.operator_status
    
    db.query(models.SlcOperatorNodes).filter_by(operator_name=opr.operator_name).delete()
    for node in request.params.getall("operator_nodes"):
        onode = models.SlcOperatorNodes()
        onode.operator_name = form.d.operator_name
        onode.node_name = node
        db.add(onode)
    
    # update rules
    db.query(models.SlcOperatorRule).filter_by(operator_name=opr.operator_name).delete()
    
    for path in request.params.getall("rule_item"):
        item = permit.get_route(path)
        if not item:continue
        rule = models.SlcOperatorRule()
        rule.operator_name = opr.operator_name
        rule.rule_name = item['name']
        rule.rule_path = item['path']
        rule.rule_category = item['category']
        db.add(rule)
        
    permit.unbind_opr(opr.operator_name)
    for rule in db.query(models.SlcOperatorRule).filter_by(operator_name=opr.operator_name):
        permit.bind_opr(rule.operator_name,rule.rule_path)

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)修改操作员信息:%s'%(get_cookie("username"),opr.operator_name)
    db.add(ops_log)

    db.commit()
    redirect("/opr")    
    
permit.add_route("/opr/update",u"修改操作员",u"系统管理",order=3.02,is_open=False)

@app.get('/opr/delete',apply=auth_opr)
def opr_delete(db):     
    opr_id = request.params.get("opr_id")
    opr = db.query(models.SlcOperator).get(opr_id)
    db.query(models.SlcOperatorNodes).filter_by(operator_name=opr.operator_name).delete()
    db.query(models.SlcOperatorRule).filter_by(operator_name=opr.operator_name).delete()
    db.query(models.SlcOperator).filter_by(id=opr_id).delete()

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)删除操作员信息:%s'%(get_cookie("username"),opr_id)
    db.add(ops_log)

    db.commit() 
    redirect("/opr")    

permit.add_route("/opr/delete",u"删除操作员",u"系统管理",order=3.03,is_open=False)



###############################################################################
# roster manage    
###############################################################################

@app.route('/roster',apply=auth_opr,method=['GET','POST'])
def roster(db):   
    _query = db.query(models.SlcRadRoster)
    return render("sys_roster_list", 
        page_data = get_page_data(_query))
        
permit.add_route("/roster",u"黑白名单管理",u"系统管理",is_menu=True,order=6)

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
    roster.mac_addr = form.d.mac_addr.replace("-",":").upper()
    roster.begin_time = form.d.begin_time
    roster.end_time = form.d.end_time
    roster.roster_type = form.d.roster_type
    db.add(roster)

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)新增黑白名单信息:%s'%(get_cookie("username"),roster.mac_addr)
    db.add(ops_log)

    db.commit()
    websock.update_cache("roster",mac_addr=roster.mac_addr)
    redirect("/roster")
    
permit.add_route("/roster/add",u"新增黑白名单",u"系统管理",order=6.01)    

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
    roster.begin_time = form.d.begin_time
    roster.end_time = form.d.end_time
    roster.roster_type = form.d.roster_type

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)修改黑白名单信息:%s'%(get_cookie("username"),roster.mac_addr)
    db.add(ops_log)

    db.commit()
    websock.update_cache("roster",mac_addr=roster.mac_addr)
    redirect("/roster")    
    
permit.add_route("/roster/update",u"修改白黑名单",u"系统管理",order=6.02)    

@app.get('/roster/delete',apply=auth_opr)
def roster_delete(db):     
    roster_id = request.params.get("roster_id")
    mac_addr = db.query(models.SlcRadRoster).get(roster_id).mac_addr
    db.query(models.SlcRadRoster).filter_by(id=roster_id).delete()

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)删除黑白名单信息:%s'%(get_cookie("username"),roster_id)
    db.add(ops_log)

    db.commit() 
    websock.update_cache("roster",mac_addr=mac_addr)
    redirect("/roster")        

permit.add_route("/roster/delete",u"删除黑白名单",u"系统管理",order=6.03)
