#!/usr/bin/env python
# coding:utf-8

from bottle import Bottle
from bottle import request
from bottle import response
from bottle import redirect
from bottle import static_file
from bottle import mako_template as render
from tablib import Dataset
from toughradius.console.websock import websock
from toughradius.console import models
from toughradius.console.libs import utils
from toughradius.console.base import *
from toughradius.console.admin import forms
import bottle
import datetime
from sqlalchemy import func

__prefix__ = "/ticket"

app = Bottle()
app.config['__prefix__'] = __prefix__

###############################################################################
# ticket manage
###############################################################################

@app.route('/', apply=auth_opr, method=['GET', 'POST'])
def ticket_query(db, render):
    node_id = request.params.get('node_id')
    account_number = request.params.get('account_number')
    framed_ipaddr = request.params.get('framed_ipaddr')
    mac_addr = request.params.get('mac_addr')
    query_begin_time = request.params.get('query_begin_time')
    query_end_time = request.params.get('query_end_time')
    opr_nodes = get_opr_nodes(db)
    _query = db.query(
        models.SlcRadTicket.id,
        models.SlcRadTicket.account_number,
        models.SlcRadTicket.nas_addr,
        models.SlcRadTicket.acct_session_id,
        models.SlcRadTicket.acct_start_time,
        models.SlcRadTicket.acct_stop_time,
        models.SlcRadTicket.acct_input_octets,
        models.SlcRadTicket.acct_output_octets,
        models.SlcRadTicket.acct_input_gigawords,
        models.SlcRadTicket.acct_output_gigawords,
        models.SlcRadTicket.framed_ipaddr,
        models.SlcRadTicket.mac_addr,
        models.SlcRadTicket.nas_port_id,
        models.SlcMember.node_id,
        models.SlcMember.realname
    ).filter(
        models.SlcRadTicket.account_number == models.SlcRadAccount.account_number,
        models.SlcMember.member_id == models.SlcRadAccount.member_id
    )
    if node_id:
        _query = _query.filter(models.SlcMember.node_id == node_id)
    else:
        _query = _query.filter(models.SlcMember.node_id.in_([i.id for i in opr_nodes]))
    if account_number:
        _query = _query.filter(models.SlcRadTicket.account_number.like('%' + account_number + '%'))
    if framed_ipaddr:
        _query = _query.filter(models.SlcRadTicket.framed_ipaddr == framed_ipaddr)
    if mac_addr:
        _query = _query.filter(models.SlcRadTicket.mac_addr == mac_addr)
    if query_begin_time:
        _query = _query.filter(models.SlcRadTicket.acct_start_time >= query_begin_time)
    if query_end_time:
        _query = _query.filter(models.SlcRadTicket.acct_stop_time <= query_end_time)

    _query = _query.order_by(models.SlcRadTicket.acct_start_time.desc())
    return render("ops_ticket_list", page_data=get_page_data(_query),
                  node_list=opr_nodes, **request.params)

permit.add_route("/ticket", u"上网日志查询", u"维护管理", is_menu=True, order=3)
