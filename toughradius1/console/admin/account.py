#!/usr/bin/env python
# coding:utf-8
from bottle import Bottle
from bottle import request
from bottle import response
from bottle import redirect
from bottle import static_file
from bottle import abort
from hashlib import md5
from tablib import Dataset
from toughradius.console.base import *
from toughradius.console.libs import utils
from toughradius.console.websock import websock
from toughradius.console import models
import bottle
from toughradius.console.admin import account_forms
import decimal
import datetime

__prefix__ = "/account"

decimal.getcontext().prec = 11
decimal.getcontext().rounding = decimal.ROUND_UP

app = Bottle()
app.config['__prefix__'] = __prefix__
member_detail_url_formatter = "/member/detail?account_number={0}".format

###############################################################################
# account open
###############################################################################

@app.get('/open', apply=auth_opr)
def account_open(db, render):
    member_id = request.params.get('member_id')
    member = db.query(models.SlcMember).get(member_id)
    products = [(n.id, n.product_name) for n in get_opr_products(db)]
    form = account_forms.account_open_form(products)
    form.member_id.set_value(member_id)
    form.realname.set_value(member.realname)
    form.node_id.set_value(member.node_id)
    return render("bus_open_form", form=form)


@app.post('/open', apply=auth_opr)
def account_open(db, render):
    products = [(n.id, n.product_name) for n in get_opr_products(db)]
    form = account_forms.account_open_form(products)
    if not form.validates(source=request.forms):
        return render("bus_open_form", form=form)

    if db.query(models.SlcRadAccount).filter_by(
        account_number=form.d.account_number).count() > 0:
        return render("bus_open_form", form=form, msg=u"上网账号已经存在")

    if form.d.ip_address and db.query(models.SlcRadAccount).filter_by(ip_address=form.d.ip_address).count() > 0:
        return render("bus_open_form", form=form, msg=u"ip%s已经被使用" % form.d.ip_address)

    accept_log = models.SlcRadAcceptLog()
    accept_log.accept_type = 'open'
    accept_log.accept_source = 'console'
    accept_log.account_number = form.d.account_number
    accept_log.accept_time = utils.get_currtime()
    accept_log.operator_name = get_cookie("username")
    accept_log.accept_desc = u"用户新增账号：上网账号:%s" % (form.d.account_number)
    db.add(accept_log)
    db.flush()
    db.refresh(accept_log)

    _datetime = utils.get_currtime()
    order_fee = 0
    balance = 0
    expire_date = form.d.expire_date
    product = db.query(models.SlcRadProduct).get(form.d.product_id)
    # 预付费包月
    if product.product_policy == 0:
        order_fee = decimal.Decimal(product.fee_price) * decimal.Decimal(form.d.months)
        order_fee = int(order_fee.to_integral_value())
    # 买断包月,买断时长,买断流量
    elif product.product_policy in (2, 3, 5):
        order_fee = int(product.fee_price)
    # 预付费时长,预付费流量
    elif product.product_policy in (1, 4):
        balance = utils.yuan2fen(form.d.fee_value)
        expire_date = '3000-11-11'

    order = models.SlcMemberOrder()
    order.order_id = utils.gen_order_id()
    order.member_id = form.d.member_id
    order.product_id = product.id
    order.account_number = form.d.account_number
    order.order_fee = order_fee
    order.actual_fee = utils.yuan2fen(form.d.fee_value)
    order.pay_status = 1
    order.accept_id = accept_log.id
    order.order_source = 'console'
    order.create_time = _datetime
    order.order_desc = u"用户增开账号"
    db.add(order)

    account = models.SlcRadAccount()
    account.account_number = form.d.account_number
    account.ip_address = form.d.ip_address
    account.member_id = int(form.d.member_id)
    account.product_id = order.product_id
    account.install_address = form.d.address
    account.mac_addr = ''
    account.password = utils.encrypt(form.d.password)
    account.status = form.d.status
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
    account.account_desc = form.d.account_desc
    db.add(account)

    db.commit()
    redirect(member_detail_url_formatter(account.account_number))




###############################################################################
# account update
###############################################################################

@app.get('/update', apply=auth_opr)
def account_update_get(db, render):
    account_number = request.params.get("account_number")
    account = db.query(models.SlcRadAccount).get(account_number)
    form = account_forms.account_update_form()
    form.fill(account)
    return render("base_form", form=form)


@app.post('/update', apply=auth_opr)
def account_update_post(db, render):
    form = account_forms.account_update_form()
    if not form.validates(source=request.forms):
        return render("base_form", form=form)

    account = db.query(models.SlcRadAccount).get(form.d.account_number)
    account.ip_address = form.d.ip_address
    account.install_address = form.d.install_address
    account.user_concur_number = form.d.user_concur_number
    account.bind_mac = form.d.bind_mac
    account.bind_vlan = form.d.bind_vlan
    account.account_desc = form.d.account_desc
    if form.d.new_password:
        account.password = utils.encrypt(form.d.new_password)

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)修改上网账号信息:%s' % (get_cookie("username"), account.account_number)
    db.add(ops_log)

    db.commit()
    websock.update_cache("account", account_number=account.account_number)
    redirect(member_detail_url_formatter(account.account_number))





###############################################################################
# account pause
###############################################################################

@app.post('/pause', apply=auth_opr)
def account_pause(db, render):
    account_number = request.params.get("account_number")
    account = db.query(models.SlcRadAccount).get(account_number)

    if account.status != 1:
        return dict(msg=u"用户当前状态不允许停机")

    _datetime = utils.get_currtime()
    account.last_pause = _datetime
    account.status = 2

    accept_log = models.SlcRadAcceptLog()
    accept_log.accept_type = 'pause'
    accept_log.accept_source = 'console'
    accept_log.accept_desc = u"用户停机：上网账号:%s" % (account_number)
    accept_log.account_number = account.account_number
    accept_log.accept_time = _datetime
    accept_log.operator_name = get_cookie("username")
    db.add(accept_log)

    db.commit()
    websock.update_cache("account", account_number=account.account_number)

    onlines = db.query(models.SlcRadOnline).filter_by(account_number=account_number)
    for _online in onlines:
        websock.invoke_admin("coa_request",
                             nas_addr=_online.nas_addr,
                             acct_session_id=_online.acct_session_id,
                             message_type='disconnect')
    return dict(msg=u"操作成功")




###############################################################################
# account resume
###############################################################################

@app.post('/resume', apply=auth_opr)
def account_resume(db,render):
    account_number = request.params.get("account_number")
    account = db.query(models.SlcRadAccount).get(account_number)
    if account.status != 2:
        return dict(msg=u"用户当前状态不允许复机")

    account.status = 1
    _datetime = datetime.datetime.now()
    _pause_time = datetime.datetime.strptime(account.last_pause, "%Y-%m-%d %H:%M:%S")
    _expire_date = datetime.datetime.strptime(account.expire_date + ' 23:59:59', "%Y-%m-%d %H:%M:%S")
    days = (_expire_date - _pause_time).days
    new_expire = (_datetime + datetime.timedelta(days=int(days))).strftime("%Y-%m-%d")
    account.expire_date = new_expire

    accept_log = models.SlcRadAcceptLog()
    accept_log.accept_type = 'resume'
    accept_log.accept_source = 'console'
    accept_log.accept_desc = u"用户复机：上网账号:%s" % (account_number)
    accept_log.account_number = account.account_number
    accept_log.accept_time = utils.get_currtime()
    accept_log.operator_name = get_cookie("username")
    db.add(accept_log)

    db.commit()
    websock.update_cache("account", account_number=account.account_number)
    return dict(msg=u"操作成功")





def query_account(db, account_number):
    return db.query(
        models.SlcMember.realname,
        models.SlcRadAccount.member_id,
        models.SlcRadAccount.product_id,
        models.SlcRadAccount.account_number,
        models.SlcRadAccount.expire_date,
        models.SlcRadAccount.balance,
        models.SlcRadAccount.time_length,
        models.SlcRadAccount.flow_length,
        models.SlcRadAccount.user_concur_number,
        models.SlcRadAccount.status,
        models.SlcRadAccount.mac_addr,
        models.SlcRadAccount.vlan_id,
        models.SlcRadAccount.vlan_id2,
        models.SlcRadAccount.ip_address,
        models.SlcRadAccount.bind_mac,
        models.SlcRadAccount.bind_vlan,
        models.SlcRadAccount.ip_address,
        models.SlcRadAccount.install_address,
        models.SlcRadAccount.create_time,
        models.SlcRadProduct.product_name
    ).filter(
        models.SlcRadProduct.id == models.SlcRadAccount.product_id,
        models.SlcMember.member_id == models.SlcRadAccount.member_id,
        models.SlcRadAccount.account_number == account_number
    ).first()


###############################################################################
# account next
###############################################################################

@app.get('/next', apply=auth_opr)
def account_next(db, render):
    account_number = request.params.get("account_number")
    user = query_account(db, account_number)
    form = account_forms.account_next_form()
    form.account_number.set_value(account_number)
    form.old_expire.set_value(user.expire_date)
    form.product_id.set_value(user.product_id)
    return render("bus_account_next_form", user=user, form=form)


@app.post('/next', apply=auth_opr)
def account_next(db, render):
    account_number = request.params.get("account_number")
    account = db.query(models.SlcRadAccount).get(account_number)
    user = query_account(db, account_number)
    form = account_forms.account_next_form()
    form.product_id.set_value(user.product_id)
    if account.status not in (1, 4):
        return render("bus_account_next_form", user=user, form=form, msg=u"无效用户状态")
    if not form.validates(source=request.forms):
        return render("bus_account_next_form", user=user, form=form)

    accept_log = models.SlcRadAcceptLog()
    accept_log.accept_type = 'next'
    accept_log.accept_source = 'console'
    accept_log.accept_desc = u"用户续费：上网账号:%s，续费%s元;%s" % (account_number, form.d.fee_value,form.d.operate_desc)
    accept_log.account_number = form.d.account_number
    accept_log.accept_time = utils.get_currtime()
    accept_log.operator_name = get_cookie("username")
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

    order_fee = 0
    product = db.query(models.SlcRadProduct).get(user.product_id)

    # 预付费包月
    if product.product_policy == PPMonth:
        order_fee = decimal.Decimal(product.fee_price) * decimal.Decimal(form.d.months)
        order_fee = int(order_fee.to_integral_value())
    # 买断包月,买断流量,买断时长
    elif product.product_policy in (BOMonth, BOTimes, BOFlows):
        order_fee = int(product.fee_price)

    order = models.SlcMemberOrder()
    order.order_id = utils.gen_order_id()
    order.member_id = user.member_id
    order.product_id = user.product_id
    order.account_number = form.d.account_number
    order.order_fee = order_fee
    order.actual_fee = utils.yuan2fen(form.d.fee_value)
    order.pay_status = 1
    order.accept_id = accept_log.id
    order.order_source = 'console'
    order.create_time = utils.get_currtime()


    account.status = 1
    account.expire_date = form.d.expire_date
    if product.product_policy == BOTimes:
        account.time_length += product.fee_times
    elif product.product_policy == BOFlows:
        account.flow_length += product.fee_flows

    history.new_expire_date = account.expire_date
    history.new_product_id = account.product_id
    db.add(history)

    order.order_desc = u"用户续费,续费前到期:%s,续费后到期:%s" % (history.expire_date, history.new_expire_date)
    db.add(order)

    db.commit()
    websock.update_cache("account", account_number=account_number)
    redirect(member_detail_url_formatter(account_number))




###############################################################################
# account charge
###############################################################################

@app.get('/charge', apply=auth_opr)
def account_charge(db, render):
    account_number = request.params.get("account_number")
    user = query_account(db, account_number)
    form = account_forms.account_charge_form()
    form.account_number.set_value(account_number)
    return render("bus_account_form", user=user, form=form)


@app.post('/charge', apply=auth_opr)
def account_charge(db, render):
    account_number = request.params.get("account_number")
    account = db.query(models.SlcRadAccount).get(account_number)
    user = query_account(db, account_number)
    form = account_forms.account_charge_form()
    if account.status != 1:
        return render("bus_account_form", user=user, form=form, msg=u"无效用户状态")

    if not form.validates(source=request.forms):
        return render("bus_account_form", user=user, form=form)

    accept_log = models.SlcRadAcceptLog()
    accept_log.accept_type = 'charge'
    accept_log.accept_source = 'console'
    accept_log.account_number = form.d.account_number
    accept_log.accept_time = utils.get_currtime()
    accept_log.operator_name = get_cookie("username")
    _new_fee = account.balance + utils.yuan2fen(form.d.fee_value)
    accept_log.accept_desc = u"用户充值：充值前%s元,充值后%s元;%s" % (
        utils.fen2yuan(account.balance),
        utils.fen2yuan(_new_fee),
        form.d.operate_desc
    )
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

    order = models.SlcMemberOrder()
    order.order_id = utils.gen_order_id()
    order.member_id = user.member_id
    order.product_id = user.product_id
    order.account_number = form.d.account_number
    order.order_fee = 0
    order.actual_fee = utils.yuan2fen(form.d.fee_value)
    order.pay_status = 1
    order.accept_id = accept_log.id
    order.order_source = 'console'
    order.create_time = utils.get_currtime()
    order.order_desc = accept_log.accept_desc
    db.add(order)

    account.balance += order.actual_fee

    db.commit()
    websock.update_cache("account", account_number=account_number)
    redirect(member_detail_url_formatter(account_number))




###############################################################################
# account product change
###############################################################################

@app.get('/change', apply=auth_opr)
def account_change(db, render):
    account_number = request.params.get("account_number")
    products = [(n.id, n.product_name) for n in get_opr_products(db)]
    user = query_account(db, account_number)
    form = account_forms.account_change_form(products=products)
    form.expire_date.set_value(user.expire_date)
    form.account_number.set_value(account_number)
    return render("bus_account_change_form", user=user, form=form)


@app.post('/change', apply=auth_opr)
def account_change(db, render):
    account_number = request.params.get("account_number")
    products = [(n.id, n.product_name) for n in get_opr_products(db)]
    form = account_forms.account_change_form(products=products)
    account = db.query(models.SlcRadAccount).get(account_number)
    user = query_account(db, account_number)
    if account.status not in (1, 4):
        return render("bus_account_change_form", user=user, form=form, msg=u"无效用户状态")
    if not form.validates(source=request.forms):
        return render("bus_account_change_form", user=user, form=form)

    product = db.query(models.SlcRadProduct).get(form.d.product_id)

    accept_log = models.SlcRadAcceptLog()
    accept_log.accept_type = 'change'
    accept_log.accept_source = 'console'
    accept_log.account_number = form.d.account_number
    accept_log.accept_time = utils.get_currtime()
    accept_log.operator_name = get_cookie("username")
    accept_log.accept_desc = u"用户资费变更为:%s;%s" % (product.product_name, form.d.operate_desc)
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

    account.product_id = product.id
    # (PPMonth,PPTimes,BOMonth,BOTimes,PPFlow,BOFlows)
    if product.product_policy in (PPMonth, BOMonth):
        account.expire_date = form.d.expire_date
        account.balance = 0
        account.time_length = 0
        account.flow_length = 0
    elif product.product_policy in (PPTimes, PPFlow):
        account.expire_date = MAX_EXPIRE_DATE
        account.balance = utils.yuan2fen(form.d.balance)
        account.time_length = 0
        account.flow_length = 0
    elif product.product_policy == BOTimes:
        account.expire_date = MAX_EXPIRE_DATE
        account.balance = 0
        account.time_length = utils.hour2sec(form.d.time_length)
        account.flow_length = 0
    elif product.product_policy == BOFlows:
        account.expire_date = MAX_EXPIRE_DATE
        account.balance = 0
        account.time_length = 0
        account.flow_length = utils.mb2kb(form.d.flow_length)

    order = models.SlcMemberOrder()
    order.order_id = utils.gen_order_id()
    order.member_id = account.member_id
    order.product_id = account.product_id
    order.account_number = account.account_number
    order.order_fee = 0
    order.actual_fee = utils.yuan2fen(form.d.add_value) - utils.yuan2fen(form.d.back_value)
    order.pay_status = 1
    order.accept_id = accept_log.id
    order.order_source = 'console'
    order.create_time = utils.get_currtime()

    history.new_expire_date = account.expire_date
    history.new_product_id = account.product_id
    db.add(history)

    order.order_desc = u"用户变更资费,变更前到期:%s,变更后到期:%s" % (history.expire_date, history.new_expire_date)
    db.add(order)

    db.commit()
    websock.update_cache("account", account_number=account_number)
    redirect(member_detail_url_formatter(account_number))




###############################################################################
# account cancel
###############################################################################

@app.get('/cancel', apply=auth_opr)
def account_cancel(db, render):
    account_number = request.params.get("account_number")
    user = query_account(db, account_number)
    form = account_forms.account_cancel_form()
    form.account_number.set_value(account_number)
    return render("bus_account_form", user=user, form=form)


@app.post('/cancel', apply=auth_opr)
def account_cancel(db, render):
    account_number = request.params.get("account_number")
    account = db.query(models.SlcRadAccount).get(account_number)
    user = query_account(db, account_number)
    form = account_forms.account_cancel_form()
    if account.status != 1:
        return render("bus_account_form", user=user, form=form, msg=u"无效用户状态")
    if not form.validates(source=request.forms):
        return render("bus_account_form", user=user, form=form)

    accept_log = models.SlcRadAcceptLog()
    accept_log.accept_type = 'cancel'
    accept_log.accept_source = 'console'
    accept_log.account_number = form.d.account_number
    accept_log.accept_time = utils.get_currtime()
    accept_log.operator_name = get_cookie("username")
    accept_log.accept_desc = u"用户销户退费%s(元);%s" % (form.d.fee_value, form.d.operate_desc)
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

    order = models.SlcMemberOrder()
    order.order_id = utils.gen_order_id()
    order.member_id = user.member_id
    order.product_id = user.product_id
    order.account_number = form.d.account_number
    order.order_fee = 0
    order.actual_fee = -utils.yuan2fen(form.d.fee_value)
    order.pay_status = 1
    order.order_source = 'console'
    order.accept_id = accept_log.id
    order.create_time = utils.get_currtime()
    order.order_desc = accept_log.accept_desc
    db.add(order)

    account.status = 3

    db.commit()

    websock.update_cache("account", account_number=account_number)
    onlines = db.query(models.SlcRadOnline).filter_by(account_number=account_number)
    for _online in onlines:
        websock.invoke_admin("coa_request",
                             nas_addr=_online.nas_addr,
                             acct_session_id=_online.acct_session_id,
                             message_type='disconnect')
    redirect(member_detail_url_formatter(account_number))


###############################################################################
# member update
###############################################################################
@app.get('/delete', apply=auth_opr)
def account_delete(db, render):
    account_number = request.params.get("account_number")
    if not account_number:
        raise abort(404, 'account_number is empty')
    account = db.query(models.SlcRadAccount).get(account_number)
    member_id = account.member_id

    db.query(models.SlcRadAcceptLog).filter_by(account_number=account.account_number).delete()
    db.query(models.SlcRadAccountAttr).filter_by(account_number=account.account_number).delete()
    db.query(models.SlcRadBilling).filter_by(account_number=account.account_number).delete()
    db.query(models.SlcRadTicket).filter_by(account_number=account.account_number).delete()
    db.query(models.SlcRadOnline).filter_by(account_number=account.account_number).delete()
    db.query(models.SlcRechargeLog).filter_by(account_number=account.account_number).delete()
    db.query(models.SlcRadAccount).filter_by(account_number=account.account_number).delete()
    db.query(models.SlcMemberOrder).filter_by(account_number=account.account_number).delete()

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)删除用户账号%s' % (get_cookie("username"), account_number)
    db.add(ops_log)

    db.commit()
    return redirect("/member")


permit.add_route("/account/open", u"增开用户账号", u"营业管理", order=1.01)
permit.add_route("/account/update", u"修改用户上网账号", u"营业管理", order=1.02)
permit.add_route("/account/pause", u"用户账号停机", u"营业管理", order=2.01)
permit.add_route("/account/resume", u"用户账号复机", u"营业管理", order=2.02)
permit.add_route("/account/next", u"用户账号续费", u"营业管理", order=2.03)
permit.add_route("/account/charge", u"用户账号充值", u"营业管理", order=2.04)
permit.add_route("/account/cancel", u"用户账号销户", u"营业管理", order=2.05)
permit.add_route("/account/change", u"用户资费变更", u"营业管理", order=2.05)
permit.add_route("/account/delete", u"删除用户账号", u"营业管理", order=3.02)