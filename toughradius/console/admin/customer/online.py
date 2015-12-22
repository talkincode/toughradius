#!/usr/bin/env python
#coding=utf-8

import cyclone.auth
import cyclone.escape
import cyclone.web
import decimal
from toughradius.console import models
from toughradius.console.admin.base import BaseHandler
from toughradius.common.permit import permit
from toughradius.common import utils
from toughradius.common.settings import * 

@permit.route(r"/admin/customer/online", u"用户在线查询",MenuUser, order=4.0000, is_menu=True)
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
        nas_addr = self.get_argument('nas_addr',None)
        opr_nodes = self.get_opr_nodes()
        _query = self.db.query(
            models.TrOnline.id,
            models.TrOnline.account_number,
            models.TrOnline.nas_addr,
            models.TrOnline.acct_session_id,
            models.TrOnline.acct_start_time,
            models.TrOnline.framed_ipaddr,
            models.TrOnline.mac_addr,
            models.TrOnline.nas_port_id,
            models.TrOnline.start_source,
            models.TrOnline.billing_times,
            models.TrOnline.input_total,
            models.TrOnline.output_total,
            models.TrCustomer.node_id,
            models.TrCustomer.realname
        ).filter(
            models.TrOnline.account_number == models.TrAccount.account_number,
            models.TrCustomer.customer_id == models.TrAccount.customer_id
        )
        if node_id:
            _query = _query.filter(models.TrCustomer.node_id == node_id)
        else:
            _query = _query.filter(models.TrCustomer.node_id.in_([i.id for i in opr_nodes]))

        if account_number:
            _query = _query.filter(models.TrOnline.account_number.like('%' + account_number + '%'))
        if framed_ipaddr:
            _query = _query.filter(models.TrOnline.framed_ipaddr == framed_ipaddr)
        if mac_addr:
            _query = _query.filter(models.TrOnline.mac_addr == mac_addr)
        if nas_addr:
            _query = _query.filter(models.TrOnline.nas_addr == nas_addr)

        _query = _query.order_by(models.TrOnline.acct_start_time.desc())
        return self.render("online_list.html", page_data=self.get_page_data(_query),
                      node_list=opr_nodes,
                      bas_list=self.db.query(models.TrBas), **self.get_params())








