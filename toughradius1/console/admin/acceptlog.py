#!/usr/bin/env python
# coding=utf-8

from bottle import Bottle
from bottle import request
from bottle import response
from bottle import redirect
from bottle import static_file
from hashlib import md5
from tablib import Dataset
from toughradius.console.base import *
from toughradius.console.libs import utils
from toughradius.console import models
import bottle
from toughradius.console.admin import forms
import decimal
import datetime

__prefix__ = "/acceptlog"

decimal.getcontext().prec = 11
decimal.getcontext().rounding = decimal.ROUND_UP

app = Bottle()
app.config['__prefix__'] = __prefix__

###############################################################################
# accept log manage
###############################################################################

@app.route('/', apply=auth_opr, method=['GET', 'POST'])
@app.post('/export', apply=auth_opr)
def acceptlog_query(db,render):
    node_id = request.params.get('node_id')
    accept_type = request.params.get('accept_type')
    account_number = request.params.get('account_number')
    operator_name = request.params.get('operator_name')
    query_begin_time = request.params.get('query_begin_time')
    query_end_time = request.params.get('query_end_time')
    opr_nodes = get_opr_nodes(db)
    _query = db.query(
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
        models.SlcNode.id == models.SlcMember.node_id
    )
    if operator_name:
        _query = _query.filter(models.SlcRadAcceptLog.operator_name == operator_name)
    if node_id:
        _query = _query.filter(models.SlcMember.node_id == node_id)
    else:
        _query = _query.filter(models.SlcMember.node_id.in_([i.id for i in opr_nodes]))
    if account_number:
        _query = _query.filter(models.SlcRadAcceptLog.account_number.like('%' + account_number + '%'))
    if accept_type:
        _query = _query.filter(models.SlcRadAcceptLog.accept_type == accept_type)
    if query_begin_time:
        _query = _query.filter(models.SlcRadAcceptLog.accept_time >= query_begin_time + ' 00:00:00')
    if query_end_time:
        _query = _query.filter(models.SlcRadAcceptLog.accept_time <= query_end_time + ' 23:59:59')
    _query = _query.order_by(models.SlcRadAcceptLog.accept_time.desc())
    type_map = ACCEPT_TYPES

    if request.path == '/':
        return render(
            "bus_acceptlog_list",
            page_data=get_page_data(_query),
            node_list=opr_nodes,
            type_map=type_map,
            get_orderid=lambda aid: db.query(models.SlcMemberOrder.order_id).filter_by(accept_id=aid).scalar(),
            **request.params
        )
    elif request.path == '/export':
        data = Dataset()
        data.append((u'区域', u'上网账号', u'受理类型', u'受理时间', u'受理渠道', u'操作员', u'受理描述'))
        for i in _query:
            data.append((
                i.node_name, i.account_number, type_map.get(i.accept_type),
                i.accept_time, i.accept_source, i.operator_name, i.accept_desc
            ))
        name = u"RADIUS-ACCEPTLOG-" + datetime.datetime.now().strftime("%Y%m%d-%H%M%S") + ".xls"
        return export_file(name, data)

permit.add_route("/acceptlog", u"用户受理查询", MenuBus, is_menu=True, order=3)
permit.add_route("/acceptlog/export", u"用户受理导出", MenuBus, order=3.01)
