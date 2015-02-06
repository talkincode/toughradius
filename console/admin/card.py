#!/usr/bin/env python
#coding:utf-8

from bottle import Bottle
from bottle import request
from bottle import response
from bottle import redirect
from bottle import static_file
from bottle import mako_template as render
from tablib import Dataset
from websock import websock
import bottle
import models
import forms
import datetime
from libs import utils
from base import *
from sqlalchemy import func

app = Bottle()

@app.route('/list',apply=auth_opr,method=['GET','POST'])
@app.route('/export',apply=auth_opr)
def card_list(db):   
    product_id = request.params.get('product_id') 
    card_type = request.params.get('card_type')   
    query_begin_time = request.params.get('query_begin_time')
    query_end_time = request.params.get('query_end_time')
    
    _query = db.query(models.SlcRechargerCard)
    if product_id:
        _query.filter(models.SlcRechargerCard.product_id==product_id)
    if card_type:
        _query.filter(models.SlcRechargerCard.card_type==card_type)
    if query_begin_time:
        _query = _query.filter(models.SlcRechargerCard.create_time >= query_begin_time+' 00:00:00')
    if query_end_time:
        _query = _query.filter(models.SlcRechargerCard.create_time <= query_end_time+' 23:59:59')
    
    products =  db.query(models.SlcRadProduct).filter_by(
        product_status = 0
    )
    if request.path == '/list':
        return render("card_list", 
            page_data = get_page_data(_query),
            card_types = forms.card_types,
            products = products,
            **request.params
        )
    elif request.path == '/export':
        pass
    
permit.add_route("/card/list",u"充值卡管理",u"系统管理",is_menu=True,order=6)
permit.add_route("/card/export",u"充值卡导出",u"系统管理",order=6.01)

@app.get('/create',apply=auth_opr)
def card_create(db):
    products = [ (n.id,n.product_name) for n in db.query(models.SlcRadProduct).filter_by(
        product_status = 0
    )]
    form = forms.recharge_card_form(products)
    return render("card_form",form=form)
    
permit.add_route("/card/create",u"充值卡生成",u"系统管理",order=6.02)