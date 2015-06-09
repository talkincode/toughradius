#!/usr/bin/env python
# coding:utf-8
import sys, os
from twisted.internet import reactor
from bottle import Bottle
from bottle import request
from bottle import response
from bottle import redirect
from bottle import run as runserver
from bottle import static_file
from bottle import abort
from hashlib import md5
from tablib import Dataset
from toughradius.console.libs import sqla_plugin
from urlparse import urljoin
from toughradius.console.base import *
from toughradius.console.libs import utils
from toughradius.console.libs.validate import vcache
from toughradius.console.libs.smail import mail
from toughradius.console.websock import websock
from toughradius.console import models
from toughradius.console.customer import forms
from sqlalchemy.sql import exists
import time
import bottle
import decimal
import datetime
import functools

__prefix__ = "/open"

app = Bottle()
app.config['__prefix__'] = __prefix__

###############################################################################
# account open
###############################################################################

def check_card(card):
    if not card:
        return dict(code=1, data=u"充值卡不存在")
    if card.card_status == CardInActive:
        return dict(code=1, data=u"充值卡未激活")
    elif card.card_status == CardUsed:
        return dict(code=1, data=u"充值卡已被使用")
    elif card.card_status == CardRecover:
        return dict(code=1, data=u"充值卡已被回收")
    if card.expire_date < utils.get_currdate():
        return dict(code=1, data=u"充值卡已过期")
    return dict(code=0)

@app.get('/querycp', apply=auth_cus)
def query_card_products(db, render):
    ''' query product by card'''
    recharge_card = request.params.get('recharge_card')
    card = db.query(models.SlcRechargerCard).filter_by(card_number=recharge_card).first()

    check_result = check_card(card)
    if check_result['code'] > 0:
        return check_result

    if card.card_type == BalanceCard:
        products = [(n.id, n.product_name) for n in db.query(models.SlcRadProduct).filter(
            models.SlcRadProduct.product_status == 0,
            models.SlcRadProduct.product_policy.in_((PPTimes, PPFlow))
        )]
        return dict(code=0, data={'products': products})
    elif card.card_type == ProductCard:
        product = db.query(models.SlcRadProduct).get(card.product_id)
        return dict(code=0, data={'products': [(product.id, product.product_name)]})


@app.get('/', apply=auth_cus)
def account_open(db, render):
    member = db.query(models.SlcMember).get(get_cookie("customer_id"))
    if member.email_active == 0 and get_param_value(db, "customer_must_active") == "1":
        return render("error", msg=u"激活您的电子邮箱才能使用此功能")

    r = ['0', '1', '2', '3', '4', '5', '6', '7', '8', '9']
    rg = utils.random_generator

    def random_account():
        _num = ''.join([rg.choice(r) for _ in range(9)])
        if db.query(models.SlcRadAccount).filter_by(account_number=_num).count() > 0:
            return random_account()
        else:
            return _num

    form = forms.account_open_form()
    form.recharge_card.set_value('')
    form.recharge_pwd.set_value('')
    form.account_number.set_value(random_account())
    return render('card_open_form', form=form)


@app.post('/', apply=auth_cus)
def account_open(db, render):
    form = forms.account_open_form()
    if not form.validates(source=request.forms):
        return render("card_open_form", form=form)
    if vcache.is_over(get_cookie("customer_id"), form.d.recharge_card):
        return render("card_open_form", form=form, msg=u"该充值卡一小时内密码输入错误超过5次，请一小时后再试")

    card = db.query(models.SlcRechargerCard).filter_by(card_number=form.d.recharge_card).first()
    check_result = check_card(card)
    if check_result['code'] > 0:
        return render('card_open_form', form=form, msg=check_result['data'])

    if utils.decrypt(card.card_passwd) != form.d.recharge_pwd:
        vcache.incr(get_cookie("customer_id"), form.d.recharge_card)
        errs = vcache.errs(get_cookie("customer_id"), form.d.recharge_card)
        return render('card_open_form', form=form, msg=u"充值卡密码错误%s次" % errs)

    vcache.clear(get_cookie("customer_id"), form.d.recharge_card)

    # start open
    accept_log = models.SlcRadAcceptLog()
    accept_log.accept_type = 'open'
    accept_log.accept_source = 'customer'
    accept_log.account_number = form.d.account_number
    accept_log.accept_time = utils.get_currtime()
    accept_log.operator_name = "customer"
    accept_log.accept_desc = u"用户自助开户:%s" % (form.d.account_number)
    db.add(accept_log)
    db.flush()
    db.refresh(accept_log)

    _datetime = utils.get_currtime()
    order_fee = 0
    balance = 0
    expire_date = utils.add_months(datetime.datetime.now(), card.months).strftime("%Y-%m-%d")
    product = db.query(models.SlcRadProduct).get(form.d.product_id)
    # 预付费包月
    if product.product_policy == PPMonth:
        order_fee = decimal.Decimal(product.fee_price) * decimal.Decimal(card.months)
        order_fee = int(order_fee.to_integral_value())
    # 买断包月,买断时长,买断流量
    elif product.product_policy in (BOMonth, BOTimes, BOFlows):
        order_fee = int(product.fee_price)
    # 预付费时长,预付费流量
    elif product.product_policy in (PPTimes, PPFlow):
        balance = card.fee_value
        expire_date = '3000-11-11'

    order = models.SlcMemberOrder()
    order.order_id = utils.gen_order_id()
    order.member_id = get_cookie("customer_id")
    order.product_id = product.id
    order.account_number = form.d.account_number
    order.order_fee = order_fee
    order.actual_fee = card.fee_value
    order.pay_status = 1
    order.accept_id = accept_log.id
    order.order_source = 'customer'
    order.create_time = _datetime
    order.order_desc = u"用户使用充值卡[ %s ]开户" % form.d.recharge_card
    db.add(order)

    account = models.SlcRadAccount()
    account.account_number = form.d.account_number
    account.ip_address = ''
    account.member_id = get_cookie("customer_id")
    account.product_id = order.product_id
    account.install_address = ''
    account.mac_addr = ''
    account.password = utils.encrypt(form.d.password)
    account.status = 1
    account.balance = balance
    account.time_length = int(product.fee_times)
    account.flow_length = int(product.fee_flows)
    account.expire_date = expire_date
    account.user_concur_number = product.concur_number
    account.bind_mac = product.bind_mac
    account.bind_vlan = product.bind_vlan
    account.vlan_id = 0
    account.vlan_id2 = 0
    account.create_time = _datetime
    account.update_time = _datetime
    db.add(account)

    clog = models.SlcRechargeLog()
    clog.member_id = get_cookie("customer_id")
    clog.card_number = card.card_number
    clog.account_number = form.d.account_number
    clog.recharge_status = 0
    clog.recharge_time = _datetime
    db.add(clog)

    card.card_status = CardUsed

    db.commit()
    redirect('/')

