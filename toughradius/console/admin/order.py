#!/usr/bin/env python
#coding=utf-8

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

__prefix__ = "/orders"

decimal.getcontext().prec = 11
decimal.getcontext().rounding = decimal.ROUND_UP

app = Bottle()
app.config['__prefix__'] = __prefix__

###############################################################################
# order query
###############################################################################

@app.route('/', apply=auth_opr, method=['GET', 'POST'])
@app.post('/export', apply=auth_opr)
def order_query(db, render):
    node_id = request.params.get('node_id')
    product_id = request.params.get('product_id')
    pay_status = request.params.get('pay_status')
    account_number = request.params.get('account_number')
    query_begin_time = request.params.get('query_begin_time')
    query_end_time = request.params.get('query_end_time')
    opr_nodes = get_opr_nodes(db)
    _query = db.query(
        models.SlcMemberOrder,
        models.SlcMember.node_id,
        models.SlcMember.realname,
        models.SlcRadProduct.product_name,
        models.SlcNode.node_name
    ).filter(
        models.SlcMemberOrder.product_id == models.SlcRadProduct.id,
        models.SlcMemberOrder.member_id == models.SlcMember.member_id,
        models.SlcNode.id == models.SlcMember.node_id
    )
    if node_id:
        _query = _query.filter(models.SlcMember.node_id == node_id)
    else:
        _query = _query.filter(models.SlcMember.node_id.in_([i.id for i in opr_nodes]))
    if account_number:
        _query = _query.filter(models.SlcMemberOrder.account_number.like('%' + account_number + '%'))
    if product_id:
        _query = _query.filter(models.SlcMemberOrder.product_id == product_id)
    if pay_status:
        _query = _query.filter(models.SlcMemberOrder.pay_status == pay_status)
    if query_begin_time:
        _query = _query.filter(models.SlcMemberOrder.create_time >= query_begin_time + ' 00:00:00')
    if query_end_time:
        _query = _query.filter(models.SlcMemberOrder.create_time <= query_end_time + ' 23:59:59')
    _query = _query.order_by(models.SlcMemberOrder.create_time.desc())

    if request.path == '/':
        return render("bus_order_list",
                      node_list=opr_nodes,
                      products=db.query(models.SlcRadProduct).filter_by(product_status=0),
                      page_data=get_page_data(_query), **request.params)
    elif request.path == '/export':
        data = Dataset()
        data.append((
            u'区域', u"用户姓名", u'上网账号', u'资费', u"订购时间",
            u'订单费用', u'实缴费用', u'支付状态', u'订购渠道', u'订单描述'
        ))
        _f2y = utils.fen2yuan
        _fms = utils.fmt_second
        _pst = {0: u'未支付', 1: u'已支付', 2: u'已取消'}
        for i, _, _realname, _product_name, _node_name in _query:
            data.append((
                _node_name, _realname, i.account_number, _product_name,
                i.create_time, _f2y(i.order_fee), _f2y(i.actual_fee),
                _pst.get(i.pay_status), i.order_source, i.order_desc
            ))
        name = u"RADIUS-ORDERS-" + datetime.datetime.now().strftime("%Y%m%d-%H%M%S") + ".xls"
        return export_file(name, data)


permit.add_route("/orders", u"用户交易查询", u"营业管理", is_menu=True, order=5)
permit.add_route("/orders/export", u"用户交易导出", u"营业管理", order=5.01)