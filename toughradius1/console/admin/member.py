#!/usr/bin/env python
# coding=utf-8

from bottle import Bottle
from bottle import abort
from tablib import Dataset
from toughradius.console.base import *
from toughradius.console.libs import utils
from toughradius.console import models
from toughradius.console.admin import member_forms
import decimal
import datetime

__prefix__ = "/member"

decimal.getcontext().prec = 11
decimal.getcontext().rounding = decimal.ROUND_UP

app = Bottle()
app.config['__prefix__'] = __prefix__

member_detail_url_formatter = "/member/detail?account_number={0}".format

###############################################################################
# member query
###############################################################################

@app.route('/', apply=auth_opr, method=['GET', 'POST'])
@app.post('/export', apply=auth_opr)
def member_query(db, render):
    node_id = request.params.get('node_id')
    realname = request.params.get('realname')
    idcard = request.params.get('idcard')
    mobile = request.params.get('mobile')
    user_name = request.params.get('user_name')
    status = request.params.get('status')
    product_id = request.params.get('product_id')
    address = request.params.get('address')
    expire_days = request.params.get('expire_days')
    opr_nodes = get_opr_nodes(db)
    _query = db.query(
        models.SlcMember,
        models.SlcRadAccount,
        models.SlcRadProduct.product_name,
        models.SlcNode.node_desc
    ).filter(
        models.SlcRadProduct.id == models.SlcRadAccount.product_id,
        models.SlcMember.member_id == models.SlcRadAccount.member_id,
        models.SlcNode.id == models.SlcMember.node_id
    )

    _now = datetime.datetime.now()

    if idcard:
        _query = _query.filter(models.SlcMember.idcard == idcard)
    if mobile:
        _query = _query.filter(models.SlcMember.mobile == mobile)
    if node_id:
        _query = _query.filter(models.SlcMember.node_id == node_id)
    else:
        _query = _query.filter(models.SlcMember.node_id.in_([i.id for i in opr_nodes]))
    if realname:
        _query = _query.filter(models.SlcMember.realname.like('%' + realname + '%'))
    if user_name:
        _query = _query.filter(models.SlcRadAccount.account_number.like('%' + user_name + '%'))

    #用户状态判断
    if status:
        if status == '4':
            _query = _query.filter(models.SlcRadAccount.expire_date <= _now.strftime("%Y-%m-%d"))
        elif status == '1':
            _query = _query.filter(
                models.SlcRadAccount.status == status,
                models.SlcRadAccount.expire_date >= _now.strftime("%Y-%m-%d")
            )
        else:
            _query = _query.filter(models.SlcRadAccount.status == status)

    if product_id:
        _query = _query.filter(models.SlcRadAccount.product_id == product_id)
    if address:
        _query = _query.filter(models.SlcMember.address.like('%' + address + '%'))
    if expire_days:
        _days = int(expire_days)
        td = datetime.timedelta(days=_days)
        edate = (_now + td).strftime("%Y-%m-%d")
        _query = _query.filter(models.SlcRadAccount.expire_date <= edate)
        _query = _query.filter(models.SlcRadAccount.expire_date >= _now.strftime("%Y-%m-%d"))

    if request.path == '/':
        return render("bus_member_list",
                      page_data=get_page_data(_query),
                      node_list=opr_nodes,
                      products=db.query(models.SlcRadProduct),
                      **request.params)
    elif request.path == "/export":
        data = Dataset()
        data.append((
            u'区域', u'姓名', u'证件号', u'邮箱', u'联系电话', u'地址',
            u'用户账号', u'密码', u'资费', u'过期时间', u'余额(元)',
            u'时长(小时)', u'流量(MB)', u'并发数', u'ip地址', u'状态', u'创建时间'
        ))
        for i, j, _product_name, _node_desc in _query:
            data.append((
                _node_desc, i.realname, i.idcard, i.email, i.mobile, i.address,
                j.account_number, utils.decrypt(j.password), _product_name,
                j.expire_date, utils.fen2yuan(j.balance),
                utils.sec2hour(j.time_length), utils.kb2mb(j.flow_length), j.user_concur_number, j.ip_address,
                member_forms.user_state[j.status], j.create_time
            ))
        name = u"RADIUS-USER-" + datetime.datetime.now().strftime("%Y%m%d-%H%M%S") + ".xls"
        return export_file(name, data)


@app.get('/detail', apply=auth_opr)
def member_detail(db, render):
    account_number = request.params.get('account_number')
    user = db.query(
        models.SlcMember.realname,
        models.SlcRadAccount.member_id,
        models.SlcRadAccount.account_number,
        models.SlcRadAccount.password,
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
        models.SlcRadAccount.last_pause,
        models.SlcRadAccount.create_time,
        models.SlcRadProduct.product_name,
        models.SlcRadProduct.product_policy
    ).filter(
        models.SlcRadProduct.id == models.SlcRadAccount.product_id,
        models.SlcMember.member_id == models.SlcRadAccount.member_id,
        models.SlcRadAccount.account_number == account_number
    ).first()

    member = db.query(models.SlcMember).get(user.member_id)

    orders = db.query(
        models.SlcMemberOrder.order_id,
        models.SlcMemberOrder.order_id,
        models.SlcMemberOrder.product_id,
        models.SlcMemberOrder.account_number,
        models.SlcMemberOrder.order_fee,
        models.SlcMemberOrder.actual_fee,
        models.SlcMemberOrder.pay_status,
        models.SlcMemberOrder.create_time,
        models.SlcMemberOrder.order_desc,
        models.SlcRadProduct.product_name
    ).filter(
        models.SlcRadProduct.id == models.SlcMemberOrder.product_id,
        models.SlcMemberOrder.account_number == account_number
    ).order_by(models.SlcMemberOrder.create_time.desc())

    historys = db.query(
                models.SlcRadAccountHistory.id,
                models.SlcRadAccountHistory.accept_id,
                models.SlcRadAccountHistory.account_number,
                models.SlcRadAccountHistory.product_id,
                models.SlcRadAccountHistory.new_product_id,
                models.SlcRadAccountHistory.expire_date,
                models.SlcRadAccountHistory.user_concur_number,
                models.SlcRadAccountHistory.bind_mac,
                models.SlcRadAccountHistory.bind_vlan,
                models.SlcRadAccountHistory.create_time,
                models.SlcRadAccountHistory.operate_time,
                models.SlcRadAcceptLog.accept_type,
                models.SlcRadAcceptLog.operator_name,
                models.SlcRadAccountHistory.new_expire_date,
    ).filter(
                models.SlcRadAccountHistory.accept_id == models.SlcRadAcceptLog.id,
                models.SlcRadAccountHistory.member_id == user.member_id
    ).order_by(models.SlcRadAccountHistory.operate_time.desc())

    accepts = db.query(
        models.SlcRadAcceptLog.id,
        models.SlcRadAcceptLog.accept_type,
        models.SlcRadAcceptLog.accept_time,
        models.SlcRadAcceptLog.accept_desc,
        models.SlcRadAcceptLog.operator_name,
        models.SlcRadAcceptLog.accept_source,
        models.SlcRadAcceptLog.account_number,
        models.SlcMember.node_id,
        models.SlcNode.node_name
    ).filter(
        models.SlcRadAcceptLog.account_number == models.SlcRadAccount.account_number,
        models.SlcMember.member_id == models.SlcRadAccount.member_id,
        models.SlcNode.id == models.SlcMember.node_id,
        models.SlcRadAcceptLog.account_number == account_number
    ).order_by(models.SlcRadAcceptLog.accept_time.desc())

    get_orderid = lambda aid: db.query(models.SlcMemberOrder.order_id).filter_by(accept_id=aid).scalar()

    type_map = ACCEPT_TYPES

    return render("bus_member_detail",
                  member=member,
                  user=user,
                  orders=orders,
                  accepts=accepts,
                  type_map=type_map,
                  historys=historys,
                  get_orderid=get_orderid
                  )



###############################################################################
# member delete
###############################################################################
@app.get('/delete', apply=auth_opr)
def member_delete(db, render):
    member_id = request.params.get("member_id")
    if not member_id:
        raise abort(404, 'member_id is empty')
    db.query(models.SlcMember).filter_by(member_id=member_id).delete()
    for account in db.query(models.SlcRadAccount).filter_by(member_id=member_id):
        db.query(models.SlcRadAcceptLog).filter_by(account_number=account.account_number).delete()
        db.query(models.SlcRadAccountAttr).filter_by(account_number=account.account_number).delete()
        db.query(models.SlcRadBilling).filter_by(account_number=account.account_number).delete()
        db.query(models.SlcRadTicket).filter_by(account_number=account.account_number).delete()
        db.query(models.SlcRadOnline).filter_by(account_number=account.account_number).delete()
        db.query(models.SlcRechargeLog).filter_by(account_number=account.account_number).delete()
    db.query(models.SlcRadAccount).filter_by(member_id=member_id).delete()
    db.query(models.SlcMemberOrder).filter_by(member_id=member_id).delete()

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)删除用户%s' % (get_cookie("username"), member_id)
    db.add(ops_log)

    db.commit()

    return redirect("/member")


###############################################################################
# member update
###############################################################################

@app.get('/update', apply=auth_opr)
def member_update_get(db, render):
    member_id = request.params.get("member_id")
    account_number = request.params.get("account_number")
    member = db.query(models.SlcMember).get(member_id)
    nodes = [(n.id, n.node_name) for n in get_opr_nodes(db)]
    form = member_forms.member_update_form(nodes)
    form.fill(member)
    form.account_number.set_value(account_number)
    return render("base_form", form=form)


@app.post('/update', apply=auth_opr)
def member_update_post(db, render):
    nodes = [(n.id, n.node_name) for n in get_opr_nodes(db)]
    form = member_forms.member_update_form(nodes)
    if not form.validates(source=request.forms):
        return render("base_form", form=form)

    member = db.query(models.SlcMember).get(form.d.member_id)
    member.realname = form.d.realname
    if form.d.new_password:
        member.password = md5(form.d.new_password.encode()).hexdigest()
    member.email = form.d.email
    member.idcard = form.d.idcard
    member.mobile = form.d.mobile
    member.address = form.d.address
    member.member_desc = form.d.member_desc

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)修改用户信息:%s' % (get_cookie("username"), member.member_name)
    db.add(ops_log)

    db.commit()
    redirect(member_detail_url_formatter(form.d.account_number))


###############################################################################
# member open
###############################################################################

@app.get('/open', apply=auth_opr)
def member_open_get(db, render):
    nodes = [(n.id, n.node_desc) for n in get_opr_nodes(db)]
    products = [(n.id, n.product_name) for n in get_opr_products(db)]
    form = member_forms.user_open_form(nodes, products)
    return render("bus_open_form", form=form)


@app.post('/open', apply=auth_opr)
def member_open_post(db, render):
    nodes = [(n.id, n.node_desc) for n in get_opr_nodes(db)]
    products = [(n.id, n.product_name) for n in get_opr_products(db)]
    form = member_forms.user_open_form(nodes, products)
    if not form.validates(source=request.forms):
        return render("bus_open_form", form=form)

    if db.query(models.SlcRadAccount).filter_by(account_number=form.d.account_number).count() > 0:
        return render("bus_open_form", form=form, msg=u"上网账号%s已经存在" % form.d.account_number)

    if form.d.ip_address and db.query(models.SlcRadAccount).filter_by(ip_address=form.d.ip_address).count() > 0:
        return render("bus_open_form", form=form, msg=u"ip%s已经被使用" % form.d.ip_address)

    if db.query(models.SlcMember).filter_by(
        member_name=form.d.member_name).count() > 0:
        return render("bus_open_form", form=form, msg=u"用户名%s已经存在" % form.d.member_name)

    member = models.SlcMember()
    member.node_id = form.d.node_id
    member.realname = form.d.realname
    member.member_name = form.d.member_name or form.d.account_number
    mpwd = form.d.member_password or form.d.password
    member.password = md5(mpwd.encode()).hexdigest()
    member.idcard = form.d.idcard
    member.sex = '1'
    member.age = '0'
    member.email = ''
    member.mobile = form.d.mobile
    member.address = form.d.address
    member.create_time = utils.get_currtime()
    member.update_time = utils.get_currtime()
    member.email_active = 0
    member.mobile_active = 0
    member.active_code = utils.get_uuid()
    member.member_desc = form.d.member_desc
    db.add(member)
    db.flush()
    db.refresh(member)

    accept_log = models.SlcRadAcceptLog()
    accept_log.accept_type = 'open'
    accept_log.accept_source = 'console'
    accept_log.account_number = form.d.account_number
    accept_log.accept_time = member.create_time
    accept_log.operator_name = get_cookie("username")
    accept_log.accept_desc = u"用户新开户：(%s)%s" % (member.member_name, member.realname)
    db.add(accept_log)
    db.flush()
    db.refresh(accept_log)

    order_fee = 0
    balance = 0
    expire_date = form.d.expire_date
    product = db.query(models.SlcRadProduct).get(form.d.product_id)
    # 预付费包月
    if product.product_policy == 0:
        order_fee = decimal.Decimal(product.fee_price) * decimal.Decimal(form.d.months)
        order_fee = int(order_fee.to_integral_value())
    # 买断包月,买断流量
    elif product.product_policy in (2, 5):
        order_fee = int(product.fee_price)
    # 预付费时长,预付费流量
    elif product.product_policy in (1, 4):
        balance = utils.yuan2fen(form.d.fee_value)
        expire_date = '3000-11-11'

    order = models.SlcMemberOrder()
    order.order_id = utils.gen_order_id()
    order.member_id = member.member_id
    order.product_id = product.id
    order.account_number = form.d.account_number
    order.order_fee = order_fee
    order.actual_fee = utils.yuan2fen(form.d.fee_value)
    order.pay_status = 1
    order.accept_id = accept_log.id
    order.order_source = 'console'
    order.create_time = member.create_time
    order.order_desc = u"用户新开账号"
    db.add(order)

    account = models.SlcRadAccount()
    account.account_number = form.d.account_number
    account.ip_address = form.d.ip_address
    account.member_id = member.member_id
    account.product_id = order.product_id
    account.install_address = member.address
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
    account.create_time = member.create_time
    account.update_time = member.create_time
    account.account_desc = member.member_desc
    db.add(account)

    db.commit()
    redirect(member_detail_url_formatter(account.account_number))


###############################################################################
# account import
###############################################################################

@app.get('/import', apply=auth_opr)
def member_import_get(db, render):
    nodes = [(n.id, n.node_desc) for n in get_opr_nodes(db)]
    products = [(p.id, p.product_name) for p in db.query(models.SlcRadProduct)]
    form = member_forms.user_import_form(nodes, products)
    return render("bus_import_form", form=form)


@app.post('/import', apply=auth_opr)
def member_import_post(db, render):
    nodes = [(n.id, n.node_desc) for n in get_opr_nodes(db)]
    products = [(n.id, n.product_name) for n in get_opr_products(db)]
    iform = member_forms.user_import_form(nodes, products)
    node_id = request.params.get('node_id')
    product_id = request.params.get('product_id')
    upload = request.files.get('import_file')
    impctx = upload.file.read()
    lines = impctx.split("\n")
    _num = 0
    impusers = []
    for line in lines:
        _num += 1
        line = line.strip()
        if not line or "用户姓名" in line: continue
        attr_array = line.split(",")
        if len(attr_array) < 11:
            return render("bus_import_form", form=iform, msg=u"line %s error: length must 11 " % _num)

        vform = member_forms.user_import_vform()
        if not vform.validates(dict(
            realname=attr_array[0],
            idcard=attr_array[1],
            mobile=attr_array[2],
            address=attr_array[3],
            account_number=attr_array[4],
            password=attr_array[5],
            begin_date=attr_array[6],
            expire_date=attr_array[7],
            balance=attr_array[8],
            time_length=utils.hour2sec(attr_array[9]),
            flow_length=utils.mb2kb(attr_array[10]))):
            return render("bus_import_form", form=iform, msg=u"line %s error: %s" % (_num, vform.errors))

        impusers.append(vform)

    for form in impusers:
        try:
            member = models.SlcMember()
            member.node_id = node_id
            member.realname = form.d.realname
            member.idcard = form.d.idcard
            member.member_name = form.d.account_number
            member.password = md5(form.d.password.encode()).hexdigest()
            member.sex = '1'
            member.age = '0'
            member.email = ''
            member.mobile = form.d.mobile
            member.address = form.d.address
            member.create_time = form.d.begin_date + ' 00:00:00'
            member.update_time = utils.get_currtime()
            member.email_active = 0
            member.mobile_active = 0
            member.active_code = utils.get_uuid()
            db.add(member)
            db.flush()
            db.refresh(member)

            accept_log = models.SlcRadAcceptLog()
            accept_log.accept_type = 'open'
            accept_log.accept_source = 'console'
            _desc = u"用户导入账号：%s" % form.d.account_number
            accept_log.accept_desc = _desc
            accept_log.account_number = form.d.account_number
            accept_log.accept_time = member.update_time
            accept_log.operator_name = get_cookie("username")
            db.add(accept_log)
            db.flush()
            db.refresh(accept_log)

            order_fee = 0
            actual_fee = 0
            balance = 0
            time_length = 0
            flow_length = 0
            expire_date = form.d.expire_date
            product = db.query(models.SlcRadProduct).get(product_id)
            # 买断时长
            if product.product_policy == 3:
                time_length = int(form.d.time_length)
            # 买断流量
            elif product.product_policy == 5:
                flow_length = int(form.d.flow_length)
            # 预付费时长,预付费流量
            elif product.product_policy in (1, 4):
                balance = utils.yuan2fen(form.d.balance)
                expire_date = '3000-11-11'

            order = models.SlcMemberOrder()
            order.order_id = utils.gen_order_id()
            order.member_id = member.member_id
            order.product_id = product.id
            order.account_number = form.d.account_number
            order.order_fee = order_fee
            order.actual_fee = actual_fee
            order.pay_status = 1
            order.accept_id = accept_log.id
            order.order_source = 'console'
            order.create_time = member.update_time
            order.order_desc = u"用户导入开户"
            db.add(order)

            account = models.SlcRadAccount()
            account.account_number = form.d.account_number
            account.member_id = member.member_id
            account.product_id = order.product_id
            account.install_address = member.address
            account.ip_address = ''
            account.mac_addr = ''
            account.password = utils.encrypt(form.d.password)
            account.status = 1
            account.balance = balance
            account.time_length = time_length
            account.flow_length = flow_length
            account.expire_date = expire_date
            account.user_concur_number = product.concur_number
            account.bind_mac = product.bind_mac
            account.bind_vlan = product.bind_vlan
            account.vlan_id = 0
            account.vlan_id2 = 0
            account.create_time = member.create_time
            account.update_time = member.update_time
            db.add(account)

        except Exception as e:
            return render("bus_import_form", form=iform, msg=u"error : %s" % str(e))

    db.commit()
    redirect("/member")


permit.add_route("/member/open", u"用户快速开户", MenuBus, is_menu=True, order=1)
permit.add_route("/member/import", u"用户数据导入", MenuBus, is_menu=True, order=2)
permit.add_route("/member", u"用户信息管理", MenuBus, is_menu=True, order=0)
permit.add_route("/member/export", u"用户信息导出", MenuBus, order=0.01)
permit.add_route("/member/detail", u"用户详情查看", MenuBus, order=0.02)
permit.add_route("/member/delete", u"删除用户信息", MenuBus, order=0.03)
permit.add_route("/member/update", u"修改用户资料", MenuBus, order=0.04)

