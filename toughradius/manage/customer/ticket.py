#!/usr/bin/env python
#coding=utf-8

import cyclone.auth
import cyclone.escape
import cyclone.web
import decimal
from toughradius import models
from toughradius.manage.base import BaseHandler
from toughlib.permit import permit
from toughlib import utils
from toughradius.manage.settings import * 

@permit.route(r"/admin/customer/ticket", u"上网日志查询",MenuUser, order=5.0000, is_menu=True)
class CustomerOnlineHandler(BaseHandler):

    @cyclone.web.authenticated
    def get(self):
        self.post()

    @cyclone.web.authenticated
    def post(self):
        node_id = self.get_argument('node_id',None)
        account_number = self.get_argument('account_number',None)
        framed_ipaddr = self.get_argument('framed_ipaddr',None)
        mac_addr = self.get_argument('mac_addr',None)
        query_begin_time = self.get_argument('query_begin_time',None)
        query_end_time = self.get_argument('query_end_time',None)
        opr_nodes = self.get_opr_nodes()
        _query = self.db.query(
            models.TrTicket.id,
            models.TrTicket.account_number,
            models.TrTicket.nas_addr,
            models.TrTicket.acct_session_id,
            models.TrTicket.acct_start_time,
            models.TrTicket.acct_stop_time,
            models.TrTicket.acct_input_octets,
            models.TrTicket.acct_output_octets,
            models.TrTicket.acct_input_gigawords,
            models.TrTicket.acct_output_gigawords,
            models.TrTicket.framed_ipaddr,
            models.TrTicket.mac_addr,
            models.TrTicket.nas_port_id,
            models.TrCustomer.node_id,
            models.TrCustomer.realname
        ).filter(
            models.TrTicket.account_number == models.TrAccount.account_number,
            models.TrCustomer.customer_id == models.TrAccount.customer_id
        )
        if node_id:
            _query = _query.filter(models.TrCustomer.node_id == node_id)
        else:
            _query = _query.filter(models.TrCustomer.node_id.in_([i.id for i in opr_nodes]))
        if account_number:
            _query = _query.filter(models.TrTicket.account_number.like('%' + account_number + '%'))
        if framed_ipaddr:
            _query = _query.filter(models.TrTicket.framed_ipaddr == framed_ipaddr)
        if mac_addr:
            _query = _query.filter(models.TrTicket.mac_addr == mac_addr)
        if query_begin_time:
            _query = _query.filter(models.TrTicket.acct_start_time >= query_begin_time)
        if query_end_time:
            _query = _query.filter(models.TrTicket.acct_stop_time <= query_end_time)

        _query = _query.order_by(models.TrTicket.acct_start_time.desc())
        return self.render("ticket_list.html", page_data=self.get_page_data(_query),
                      node_list=opr_nodes, **self.get_params())







