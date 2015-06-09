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

__prefix__ = "/billing"

decimal.getcontext().prec = 11
decimal.getcontext().rounding = decimal.ROUND_UP

app = Bottle()
app.config['__prefix__'] = __prefix__


###############################################################################
# billing log query
###############################################################################

@app.route('/', apply=auth_opr, method=['GET', 'POST'])
@app.post('/export', apply=auth_opr)
def billing_query(db, render):
    node_id = request.params.get('node_id')
    account_number = request.params.get('account_number')
    query_begin_time = request.params.get('query_begin_time')
    query_end_time = request.params.get('query_end_time')
    opr_nodes = get_opr_nodes(db)
    _query = db.query(
        models.SlcRadBilling,
        models.SlcMember.node_id,
        models.SlcNode.node_name
    ).filter(
        models.SlcRadBilling.account_number == models.SlcRadAccount.account_number,
        models.SlcMember.member_id == models.SlcRadAccount.member_id,
        models.SlcNode.id == models.SlcMember.node_id
    )
    if node_id:
        _query = _query.filter(models.SlcMember.node_id == node_id)
    else:
        _query = _query.filter(models.SlcMember.node_id.in_(i.id for i in opr_nodes))
    if account_number:
        _query = _query.filter(models.SlcRadBilling.account_number.like('%' + account_number + '%'))
    if query_begin_time:
        _query = _query.filter(models.SlcRadBilling.create_time >= query_begin_time + ' 00:00:00')
    if query_end_time:
        _query = _query.filter(models.SlcRadBilling.create_time <= query_end_time + ' 23:59:59')
    _query = _query.order_by(models.SlcRadBilling.create_time.desc())
    if request.path == '/':
        return render("bus_billing_list",
                      node_list=opr_nodes,
                      page_data=get_page_data(_query), **request.params)
    elif request.path == '/export':
        data = Dataset()
        data.append((
            u'区域', u'上网账号', u'BAS地址', u'会话编号', u'记账开始时间', u'会话时长',
            u'已扣时长', u"已扣流量", u'应扣费用', u'实扣费用', u'剩余余额',
            u'剩余时长', u'剩余流量', u'是否扣费', u'扣费时间'
        ))
        _f2y = utils.fen2yuan
        _fms = utils.fmt_second
        _k2m = utils.kb2mb
        _s2h = utils.sec2hour
        for i, _, _node_name in _query:
            data.append((
                _node_name, i.account_number, i.nas_addr, i.acct_session_id,
                i.acct_start_time, _fms(i.acct_session_time), _fms(i.acct_times), _k2m(i.acct_flows),
                _f2y(i.acct_fee), _f2y(i.actual_fee), _f2y(i.balance),
                _s2h(i.time_length), _k2m(i.flow_length),
                (i.is_deduct == 0 and u'否' or u'是'), i.create_time
            ))
        name = u"RADIUS-BILLING-" + datetime.datetime.now().strftime("%Y%m%d-%H%M%S") + ".xls"
        return export_file(name, data)

permit.add_route("/billing", u"用户计费查询", u"营业管理", is_menu=True, order=4)
permit.add_route("/billing/export", u"用户计费导出", u"营业管理", order=4.01)