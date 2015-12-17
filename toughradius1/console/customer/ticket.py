#!/usr/bin/env python
# coding:utf-8
from bottle import Bottle

from toughradius.console.base import *
from toughradius.console import models

__prefix__ = "/ticket"

app = Bottle()
app.config['__prefix__'] = __prefix__


###############################################################################
# ticket query
###############################################################################

@app.route('/', apply=auth_cus, method=['GET', 'POST'])
def ticket_query(db, render):
    account_number = request.params.get('account_number')
    query_begin_time = request.params.get('query_begin_time')
    query_end_time = request.params.get('query_end_time')
    _query = db.query(
        models.SlcRadTicket.id,
        models.SlcRadTicket.account_number,
        models.SlcRadTicket.nas_addr,
        models.SlcRadTicket.acct_session_id,
        models.SlcRadTicket.acct_start_time,
        models.SlcRadTicket.acct_input_octets,
        models.SlcRadTicket.acct_output_octets,
        models.SlcRadTicket.acct_input_gigawords,
        models.SlcRadTicket.acct_output_gigawords,
        models.SlcRadTicket.acct_stop_time,
        models.SlcRadTicket.framed_ipaddr,
        models.SlcRadTicket.mac_addr,
        models.SlcRadTicket.nas_port_id,
        models.SlcMember.node_id,
        models.SlcMember.realname
    ).filter(
        models.SlcRadTicket.account_number == models.SlcRadAccount.account_number,
        models.SlcMember.member_id == models.SlcRadAccount.member_id,
        models.SlcMember.member_id == get_cookie("customer_id")
    )
    if account_number:
        _query = _query.filter(models.SlcRadTicket.account_number == account_number)
    if query_begin_time:
        _query = _query.filter(models.SlcRadTicket.acct_start_time >= query_begin_time)
    if query_end_time:
        _query = _query.filter(models.SlcRadTicket.acct_stop_time <= query_end_time)

    _query = _query.order_by(models.SlcRadTicket.acct_start_time.desc())

    return render("ticket_list",
                  accounts=db.query(models.SlcRadAccount).filter_by(member_id=get_cookie("customer_id")),
                  page_data=get_page_data(_query),
                  **request.params)