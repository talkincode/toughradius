#!/usr/bin/env python
#coding:utf-8
from twisted.python import log
from bottle import request
from bottle import response
from bottle import redirect
from libs.paginator import Paginator
from libs import utils
import functools
import urllib
import models
import json
from beaker.cache import CacheManager

page_size = 20

__cache_timeout__ = 600

cache = CacheManager(cache_regions={'short_term':{ 'type': 'memory', 'expire': __cache_timeout__ }}) 

secret='123321qweasd',
get_cookie = lambda name: request.get_cookie(name,secret=secret)
set_cookie = lambda name,value:response.set_cookie(name,value,secret=secret)

def auth_opr(func):
    @functools.wraps(func)
    def warp(*args,**kargs):
        if not get_cookie("username"):
            log.msg("admin login timeout")
            return redirect('/login')
        else:
            return func(*args,**kargs)
    return warp

@cache.cache('get_account_node_id',expire=3600)   
def account_node_id(db,account_number):
    return  db.query(models.SlcMember.node_id).filter(
        models.SlcMember.member_id == models.SlcRadAccount.member_id,
        models.SlcRadAccount.account_number == account_number).scalar()

@cache.cache('get_member_node_id',expire=3600)   
def member_node_id(db,member_id):
    return  db.query(models.SlcMember.node_id).filter_by(member_id = member_id).scalar()

@cache.cache('get_param_value',expire=3600)   
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
    return json.dumps({c.name: getattr(mdl, c.name) for c in mdl.__table__.columns},ensure_ascii=False)







