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

__prefix__ = "/recharge"

app = Bottle()
app.config['__prefix__'] = __prefix__


###############################################################################
# recharge
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

@app.get('/')
def account_recharge(db, render):
    member = db.query(models.SlcMember).get(get_cookie("customer_id"))
    if member.email_active == 0 and get_param_value(db, "customer_must_active") == "1":
        return render("error", msg=u"激活您的电子邮箱才能使用此功能")
    account_number = request.params.get('account_number')
    form = forms.recharge_form()
    form.recharge_card.set_value('')
    form.recharge_pwd.set_value('')
    form.account_number.set_value(account_number)
    return render('base_form', form=form)


@app.post('/')
def account_recharge(db, render):
    form = forms.recharge_form()
    if not form.validates(source=request.forms):
        return render("base_form", form=form)
    if vcache.is_over(get_cookie("customer_id"), form.d.recharge_card):
        return render("base_form", form=form, msg=u"该充值卡一小时内密码输入错误超过5次，请一小时后再试")

    # 1 check card
    card = db.query(models.SlcRechargerCard).filter_by(card_number=form.d.recharge_card).first()
    check_result = check_card(card)
    if check_result['code'] > 0:
        return render('base_form', form=form, msg=check_result['data'])

    if utils.decrypt(card.card_passwd) != form.d.recharge_pwd:
        vcache.incr(get_cookie("customer_id"), form.d.recharge_card)
        errs = vcache.errs(get_cookie("customer_id"), form.d.recharge_card)
        return render('base_form', form=form, msg=u"充值卡密码错误%s次" % errs)

    vcache.clear(get_cookie("customer_id"), form.d.recharge_card)

    # 2 check account
    account = db.query(models.SlcRadAccount).filter_by(account_number=form.d.account_number).first()
    if not account:
        return render("base_form", form=form, msg=u'没有这个账号')
    if account.member_id != get_cookie("customer_id"):
        return render("base_form", form=form, msg=u'该账号用用户不匹配')
    if account.status not in (UsrNormal, UsrExpire):
        return render("base_form", form=form, msg=u'只有正常状态的用户才能充值')

    # 3 check product
    user_product = db.query(models.SlcRadProduct).get(account.product_id)
    if card.card_type == ProductCard and card.product_id != account.product_id:
        return render("base_form", form=form, msg=u'您使用的是资费卡，但资费套餐与该账号资费不匹配')
    if card.card_type == BalanceCard and user_product.product_policy not in (PPTimes, PPFlow):
        return render("base_form", form=form, msg=u'您使用的是余额卡，只能为预付费时长或预付费流量账号充值')

    # 4 start recharge
    accept_log = models.SlcRadAcceptLog()
    accept_log.accept_type = 'charge'
    accept_log.accept_source = 'customer'
    accept_log.account_number = form.d.account_number
    accept_log.accept_time = utils.get_currtime()
    accept_log.operator_name = "customer"
    accept_log.accept_desc = u"用户自助充值：上网账号:%s，充值卡:%s" % (form.d.account_number, form.d.recharge_card)
    db.add(accept_log)
    db.flush()
    db.refresh(accept_log)

    history = models.SlcRadAccountHistory()
    history.accept_id = accept_log.id
    history.account_number = account.account_number
    history.member_id = account.member_id
    history.product_id = account.product_id
    history.group_id = account.group_id
    history.password = account.password
    history.install_address = account.install_address
    history.expire_date = account.expire_date
    history.user_concur_number = account.user_concur_number
    history.bind_mac = account.bind_mac
    history.bind_vlan = account.bind_vlan
    history.account_desc = account.account_desc
    history.create_time = account.create_time
    history.operate_time = accept_log.accept_time
    db.add(history)

    _datetime = utils.get_currtime()
    order_fee = 0
    balance = 0
    expire_date = account.expire_date
    d_expire_date = datetime.datetime.strptime(expire_date, "%Y-%m-%d")
    # 预付费包月
    if user_product.product_policy == PPMonth:
        expire_date = utils.add_months(d_expire_date, card.months).strftime("%Y-%m-%d")
        order_fee = decimal.Decimal(user_product.fee_price) * decimal.Decimal(card.months)
        order_fee = int(order_fee.to_integral_value())
    # 买断包月,买断时长,买断流量
    if user_product.product_policy in (BOMonth, BOTimes, BOFlows):
        expire_date = utils.add_months(d_expire_date, card.months).strftime("%Y-%m-%d")
        order_fee = user_product.fee_price
    # 预付费时长,预付费流量
    elif user_product.product_policy in (PPTimes, PPFlow):
        balance = card.fee_value

    order = models.SlcMemberOrder()
    order.order_id = utils.gen_order_id()
    order.member_id = get_cookie("customer_id")
    order.product_id = account.product_id
    order.account_number = form.d.account_number
    order.order_fee = order_fee
    order.actual_fee = card.fee_value
    order.pay_status = 1
    order.accept_id = accept_log.id
    order.order_source = 'customer'
    order.create_time = _datetime
    order.order_desc = u"用户自助充值，充值卡[ %s ]" % form.d.recharge_card
    db.add(order)

    account.expire_date = expire_date
    account.balance += balance
    account.time_length += card.times
    account.flow_length += card.flows
    account.status = 1

    card.card_status = CardUsed

    history.new_expire_date = account.expire_date
    history.new_product_id = account.product_id

    db.commit()
    redirect("/")