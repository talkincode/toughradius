#!/usr/bin/env python
#coding=utf-8

import cyclone.auth
import cyclone.escape
import cyclone.web
import decimal
from tablib import Dataset
from toughradius.console import models
from toughradius.console.admin.base import BaseHandler
from toughradius.common.permit import permit
from toughradius.common import utils
from toughradius.common.settings import * 

@permit.route(r"/customer/acceptlog", u"用户受理日志",MenuUser, order=3.0000, is_menu=True)
class CustomerAcceptLoggerHandler(BaseHandler):

    @cyclone.web.authenticated
    def get(self):
        self.post()

    @cyclone.web.authenticated
    def post(self):
        node_id = self.get_argument('node_id',None)
        accept_type = self.get_argument('accept_type',None)
        account_number = self.get_argument('account_number',None)
        operator_name = self.get_argument('operator_name',None)
        query_begin_time = self.get_argument('query_begin_time',None)
        query_end_time = self.get_argument('query_end_time',None)
        opr_nodes = self.get_opr_nodes()
        _query = self.db.query(
            models.TrAcceptLog.id,
            models.TrAcceptLog.accept_type,
            models.TrAcceptLog.accept_time,
            models.TrAcceptLog.accept_desc,
            models.TrAcceptLog.operator_name,
            models.TrAcceptLog.accept_source,
            models.TrAcceptLog.account_number,
            models.TrCustomer.node_id,
            models.TrNode.node_name
        ).filter(
            models.TrAcceptLog.account_number == models.TrAccount.account_number,
            models.TrCustomer.customer_id == models.TrAccount.customer_id,
            models.TrNode.id == models.TrCustomer.node_id
        )
        if operator_name:
            _query = _query.filter(models.TrAcceptLog.operator_name == operator_name)
        if node_id:
            _query = _query.filter(models.TrCustomer.node_id == node_id)
        else:
            _query = _query.filter(models.TrCustomer.node_id.in_([i.id for i in opr_nodes]))
        if account_number:
            _query = _query.filter(models.TrAcceptLog.account_number.like('%' + account_number + '%'))
        if accept_type:
            _query = _query.filter(models.TrAcceptLog.accept_type == accept_type)
        if query_begin_time:
            _query = _query.filter(models.TrAcceptLog.accept_time >= query_begin_time + ' 00:00:00')
        if query_end_time:
            _query = _query.filter(models.TrAcceptLog.accept_time <= query_end_time + ' 23:59:59')
        _query = _query.order_by(models.TrAcceptLog.accept_time.desc())
        type_map = ACCEPT_TYPES

        if self.request.path == '/customer/acceptlog':
            return self.render(
                "acceptlog_list.html",
                page_data=self.get_page_data(_query),
                node_list=opr_nodes,
                type_map=type_map,
                get_orderid=lambda aid: self.db.query(models.TrCustomerOrder.order_id).filter_by(accept_id=aid).scalar(),
                **self.get_params()
            )
        elif self.request.path == '/customer/acceptlog/export':
            data = Dataset()
            data.append((u'区域', u'上网账号', u'受理类型', u'受理时间', u'受理渠道', u'操作员', u'受理描述'))
            for i in _query:
                data.append((
                    i.node_name, i.account_number, type_map.get(i.accept_type),
                    i.accept_time, i.accept_source, i.operator_name, i.accept_desc
                ))
            name = u"RADIUS-ACCEPTLOG-" + datetime.datetime.now().strftime("%Y%m%d-%H%M%S") + ".xls"
            return self.export_file(name, data)

@permit.route(r"/customer/acceptlog/export", u"用户受理日志导出",MenuUser, order=3.0001)
class CustomerAcceptLoggerExportHandler(CustomerAcceptLoggerHandler):
    pass






