#!/usr/bin/env python
#coding=utf-8

import cyclone.auth
import cyclone.escape
import cyclone.web
import decimal
from toughradius.console import models
from toughradius.console.customer.base import BaseHandler,authenticated
from toughradius.common.permit import permit
from toughradius.common import utils
from toughradius.common.settings import * 

@permit.route(r"/customer/ticket")
class CustomerTicketHandler(BaseHandler):

    @authenticated
    def get(self):
        self.post()

    @authenticated
    def post(self):
        account_number = self.get_argument('account_number',None)
        query_begin_time = self.get_argument('query_begin_time',None)
        query_end_time = self.get_argument('query_end_time',None)
        _query = self.db.query(
            models.TrTicket.id,
            models.TrTicket.account_number,
            models.TrTicket.nas_addr,
            models.TrTicket.acct_session_id,
            models.TrTicket.acct_start_time,
            models.TrTicket.acct_input_octets,
            models.TrTicket.acct_output_octets,
            models.TrTicket.acct_input_gigawords,
            models.TrTicket.acct_output_gigawords,
            models.TrTicket.acct_stop_time,
            models.TrTicket.framed_ipaddr,
            models.TrTicket.mac_addr,
            models.TrTicket.nas_port_id,
            models.TrCustomer.node_id,
            models.TrCustomer.realname
        ).filter(
            models.TrTicket.account_number == models.TrAccount.account_number,
            models.TrCustomer.customer_id == models.TrAccount.customer_id,
            models.TrCustomer.customer_id == self.current_user.cid
        )
        if account_number:
            _query = _query.filter(models.TrTicket.account_number == account_number)
        if query_begin_time:
            _query = _query.filter(models.TrTicket.acct_start_time >= query_begin_time)
        if query_end_time:
            _query = _query.filter(models.TrTicket.acct_stop_time <= query_end_time)

        _query = _query.order_by(models.TrTicket.acct_start_time.desc())
        account = self.db.query(models.TrAccount).filter_by(customer_id=self.current_user.cid)

        return self.render("ticket_list.html", 
            accounts=account, page_data=self.get_page_data(_query), **self.get_params())



