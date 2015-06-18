#!/usr/bin/env python
#coding:utf-8
from bottle import Bottle
from bottle import request
from bottle import response
from bottle import redirect
from bottle import run as runserver
from bottle import static_file
from bottle import abort
from hashlib import md5
from tablib import Dataset
from toughradius.console.base import *
from toughradius.console.libs import utils
from toughradius.console.websock import websock
from toughradius.console import models
import bottle
from toughradius.console.admin import forms
import decimal
import datetime

__prefix__ = "/bus"

decimal.getcontext().prec = 11
decimal.getcontext().rounding = decimal.ROUND_UP

app = Bottle()
app.config['__prefix__'] = __prefix__


###############################################################################
# ajax query
###############################################################################

@app.post('/opencalc',apply=auth_opr)
def opencalc(db,render):
    months = request.params.get('months',0)
    product_id = request.params.get("product_id")
    old_expire = request.params.get("old_expire")
    product = db.query(models.SlcRadProduct).get(product_id)
    # 预付费时长，预付费流量，
    if product.product_policy in (PPTimes,PPFlow):
        return dict(code=0,data=dict(policy=product.product_policy,fee_value=0,expire_date="3000-12-30"))
    # 买断时长 买断流量
    elif product.product_policy in (BOTimes,BOFlows):
        fee_value = utils.fen2yuan(product.fee_price)
        return dict(code=0,data=dict(policy=product.product_policy,fee_value=fee_value,expire_date="3000-12-30"))
    # 预付费包月 
    elif product.product_policy == PPMonth:
        fee = decimal.Decimal(months) * decimal.Decimal(product.fee_price)
        fee_value = utils.fen2yuan(int(fee.to_integral_value()))
        start_expire = datetime.datetime.now()
        if old_expire:
            start_expire = datetime.datetime.strptime(old_expire,"%Y-%m-%d")
        expire_date = utils.add_months(start_expire,int(months))
        expire_date = expire_date.strftime( "%Y-%m-%d")
        return dict(code=0,data=dict(policy=product.product_policy,fee_value=fee_value,expire_date=expire_date))
    # 买断包月
    elif product.product_policy == BOMonth:
        start_expire = datetime.datetime.now()
        if old_expire:
            start_expire = datetime.datetime.strptime(old_expire,"%Y-%m-%d")
        fee_value = utils.fen2yuan(product.fee_price)
        expire_date = utils.add_months(start_expire,product.fee_months)
        expire_date = expire_date.strftime( "%Y-%m-%d")
        return dict(code=0,data=dict(policy=product.product_policy,fee_value=fee_value,expire_date=expire_date))







