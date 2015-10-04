#!/usr/bin/env python
#coding:utf-8
from twisted.python import log
from bottle import request
from bottle import response
from bottle import redirect
from bottle import HTTPError
from bottle import static_file
from bottle import mako_template
from toughradius.console.libs.paginator import Paginator
from toughradius.console.libs import utils
from toughradius.console import models
from beaker.cache import CacheManager
from hashlib import md5
import logging
import functools
import urllib
import json
import time
import tempfile

########################################################################
# const define
########################################################################

FEES = (PPMonth,PPTimes,BOMonth,BOTimes,PPFlow,BOFlows) = (0,1,2,3,4,5)

ACCOUNT_STATUS = (UsrPreAuth,UsrNormal,UsrPause,UsrCancel,UsrExpire) = (0,1,2,3,4)

CARD_STATUS = (CardInActive,CardActive,CardUsed,CardRecover) = (0,1,2,3)

CARD_TYPE = (ProductCard,BalanceCard) = (0,1)

ACCEPT_TYPES = {
    'open':u'开户',
    'pause':u'停机',
    'resume':u'复机',
    'cancel':u'销户',
    'next':u'续费',
    'charge':u'充值',
    'change':u'变更'
}

ADMIN_MENUS = (MenuSys,MenuBus,MenuOpt,MenuStat,MenuWlan,MenuMpp) = (
    u"系统管理", u"营业管理", u"维护管理",u"统计分析", u"Wlan管理", u"微信接入")

MENU_ICONS = {
    u"系统管理" : "fa fa-cog",
    u"营业管理" : "fa fa-users",
    u"维护管理" : "fa fa-wrench",
    u"统计分析" : "fa fa-bar-chart"
}

MAX_EXPIRE_DATE = '3000-12-30'

TMPDIR = tempfile.gettempdir()

page_size = 20

__cache_timeout__ = 600

cache = CacheManager(cache_regions={'short_term':{ 'type': 'memory', 'expire': __cache_timeout__ }}) 
   
class Connect:
    def __init__(self, mkdb):
        self.conn = mkdb()

    def __enter__(self):
        return self.conn   

    def __exit__(self, exc_type, exc_value, exc_tb):
        self.conn.close()

class SecureCookie(object):
    
    def setup(self,secret):
        self.secret = secret
    
    def get_cookie(self,name):
        return request.get_cookie(md5(name).hexdigest(),secret=self.secret)
        
    def set_cookie(self,name,value,**options):
        response.set_cookie(md5(name).hexdigest(),value,secret=self.secret,**options)
        
scookie = SecureCookie()
get_cookie = scookie.get_cookie
set_cookie = scookie.set_cookie

        
########################################################################
# permission manage
########################################################################

class Permit():
    routes = {}
    def add_route(self,path,name,category,is_menu=False,order=time.time(),is_open=True):
        if not path:return
        self.routes[path] = dict(
            path=path,
            name=name,
            category=category,
            is_menu=is_menu,
            oprs=[],
            order=order,
            is_open=is_open
        )
    
    def get_route(self,path):
        return self.routes.get(path)    
        
    def bind_super(self,opr):
        for path in self.routes:
            route = self.routes.get(path)    
            route['oprs'].append(opr)         
    
    def bind_opr(self,opr,path):
        if not path or path not in self.routes:
            return
        oprs = self.routes[path]['oprs'] 
        if opr not in oprs:
            oprs.append(opr) 
            
    def unbind_opr(self,opr,path=None):
        if path:
            self.routes[path]['oprs'].remove(opr)
        else:
            for path in self.routes:
                route = self.routes.get(path)    
                if route and opr in route['oprs']:
                    route['oprs'].remove(opr)
                    
    def check_open(self,path):
        route = self.routes[path]
        return 'is_open' in route and route['is_open']
                
    def check_opr_category(self,opr,category):
        for path in self.routes:
            route = self.routes[path]
            if opr in route['oprs'] and route['category'] == category:
                return True
        return False 
        
    def build_menus(self,order_cats=[]):
        menus = [ {'category':_cat,'items':[]} for _cat in order_cats]
        for path in self.routes:
            route = self.routes[path]
            for menu in menus:
                if route['category'] == menu['category']:
                    menu['items'].append(route)
        return menus
        
    def match(self,path):
        if not path:
            return False
        return get_cookie("username") in self.routes[path]['oprs']
     
permit = Permit()       

########################################################################
# web plugin funtion
########################################################################

def export_file(name,data):
    with open(u'%s/%s' % (TMPDIR,name), 'wb') as f:
        f.write(data.xls)
    return static_file(name, root=TMPDIR,download=True)

def auth_opr(func):
    @functools.wraps(func)
    def warp(*args,**kargs):
        if not get_cookie("username"):
            log.msg("admin login timeout")
            return redirect('/login')
        else:
            opr = get_cookie("username")
            rule = permit.get_route(request.fullpath)
            if rule and opr not in rule['oprs']:
                raise HTTPError(403, u'访问被拒绝')
            return func(*args,**kargs)
    return warp
    
def auth_cus(func):
    @functools.wraps(func)
    def warp(*args,**kargs):
        if not get_cookie("customer"):
            log.msg("user login timeout")
            return redirect('/auth/login')
        else:
            return func(*args,**kargs)
    return warp

def auth_ctl(func):
    @functools.wraps(func)
    def warp(*args, **kargs):
        if not get_cookie("control_admin"):
            log.msg("control_admin login timeout")
            return redirect('/auth/login')
        else:
            return func(*args, **kargs)
    return warp


########################################################################
# cache function
########################################################################
def get_opr_nodes(db):
    opr_type = get_cookie('opr_type')
    if opr_type == 0:
        return db.query(models.SlcNode)
    opr_name = get_cookie('username')
    return db.query(models.SlcNode).filter(
        models.SlcNode.node_name == models.SlcOperatorNodes.node_name,
        models.SlcOperatorNodes.operator_name == opr_name
    )

def get_opr_products(db):
    opr_type = get_cookie('opr_type')
    if opr_type == 0:
        return db.query(models.SlcRadProduct).filter(models.SlcRadProduct.product_status == 0)
    else:
        return db.query(models.SlcRadProduct).filter(
            models.SlcRadProduct.id == models.SlcOperatorProducts.product_id,
            models.SlcOperatorProducts.operator_name == get_cookie("username"),
            models.SlcRadProduct.product_status == 0
        )

@cache.cache('get_node_name',expire=3600)   
def get_node_name(db,node_id):
    return  db.query(models.SlcNode.node_name).filter_by(id=node_id).scalar()
    
@cache.cache('get_product_name',expire=3600)   
def get_product_name(db,product_id):
    if not product_id:return
    return  db.query(models.SlcRadProduct.product_name).filter_by(id=product_id).scalar()

@cache.cache('get_account_node_id',expire=3600)   
def account_node_id(db,account_number):
    return  db.query(models.SlcMember.node_id).filter(
        models.SlcMember.member_id == models.SlcRadAccount.member_id,
        models.SlcRadAccount.account_number == account_number).scalar()

@cache.cache('get_member_node_id',expire=3600)   
def member_node_id(db,member_id):
    return  db.query(models.SlcMember.node_id).filter_by(member_id = member_id).scalar()
    
@cache.cache('get_member_by_name',expire=3600)   
def get_member_by_name(db,member_name):
    return  db.query(models.SlcMember).filter_by(member_name = member_name).first()
    
@cache.cache('get_account_by_number',expire=300)   
def get_account_by_number(db,account_number):
    return  db.query(models.SlcRadAccount).filter_by(account_number = account_number).first()
    
# @cache.cache('get_online_status',expire=30)   
def get_online_status(db,account_number):
    return  db.query(models.SlcRadOnline.id).filter_by(account_number = account_number).count() > 0
    
@cache.cache('get_param_value',expire=600)   
def get_param_value(db,pname):
    return  db.query(models.SlcParam.param_value).filter_by(param_name = pname).scalar()

def get_page_data(query):
    def _page_url(page, form_id=None):
        if form_id:return "javascript:goto_page('%s',%s);" %(form_id.strip(),page)
        request.query['page'] = page
        return request.fullpath + '?' + urllib.urlencode(request.query)
    page = int(request.params.get("page",1))
    offset = (page - 1) * page_size
    page_data = Paginator(_page_url, page, query.count(), page_size)
    page_data.result = query.limit(page_size).offset(offset)
    return page_data

def serial_json(mdl):
    if not mdl:return
    if not hasattr(mdl,'__table__'):return
    data = {}
    for c in mdl.__table__.columns:
        data[c.name] = getattr(mdl, c.name)
    return json.dumps(data,ensure_ascii=False)










