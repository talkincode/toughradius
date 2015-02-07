#!/usr/bin/env python
#coding:utf-8
from twisted.python import log
from bottle import request
from bottle import response
from bottle import redirect
from bottle import HTTPError
from libs.paginator import Paginator
from libs import utils
from hashlib import md5
import logging
import functools
import urllib
import models
import json
import time
from beaker.cache import CacheManager

page_size = 20

__cache_timeout__ = 600

cache = CacheManager(cache_regions={'short_term':{ 'type': 'memory', 'expire': __cache_timeout__ }}) 

secret='123321qweasd',

class SecureCookie(object):
    
    def __init__(self,secret):
        self.secret = secret
    
    def get_cookie(self,name):
        return request.get_cookie(md5(name).hexdigest(),secret=self.secret)
        
    def set_cookie(self,name,value,**options):
        response.set_cookie(md5(name).hexdigest(),value,secret=self.secret,**options)
        
scookie = SecureCookie(secret)
get_cookie = scookie.get_cookie
set_cookie = scookie.set_cookie

def update_secret(secret):
    global scookie
    scookie = SecureCookie(secret)
    
class Logger:
    def info(msg):
        log.msg(msg,level=logging.INFO)
    def debug(msg):
        log.msg(msg,level=logging.DEBUG)
    def error(msg,err=None):
        log.err(msg,err)

logger = Logger()


class Permit():
    '''permission rules'''
    routes = {}
    
    def add_route(self,path,name,category,is_menu=False,order=time.time()):
        if not path:return
        self.routes[path] = dict(path=path,name=name,category=category,is_menu=is_menu,oprs=[],order=order)
    
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
            return redirect('/login')
        else:
            return func(*args,**kargs)
    return warp    

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
    
@cache.cache('get_online_status',expire=60)   
def get_online_status(db,account_number):
    return  db.query(models.SlcRadOnline.id).filter_by(account_number = account_number).count() > 0
    
@cache.cache('get_param_value',expire=600)   
def get_param_value(db,pname):
    return  db.query(models.SlcParam.param_value).filter_by(param_name = pname).scalar()

def get_page_data(query):
    def _page_url(page, form_id=None):
        if form_id:return "javascript:goto_page('%s',%s);" %(form_id.strip(),page)
        request.query['page'] = page
        return request.path + '?' + urllib.urlencode(request.query)        
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







