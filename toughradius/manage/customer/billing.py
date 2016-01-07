#!/usr/bin/env python
#coding=utf-8
import cyclone.web
from tablib import Dataset
from toughradius.manage import models
from toughradius.manage.base import BaseHandler
from toughlib.permit import permit
from toughlib import utils
from toughradius.manage.settings import * 
import datetime

@permit.route(r"/admin/customer/billing", u"用户计费日志",MenuUser, order=6.0000, is_menu=True)
class CustomerBillingHandler(BaseHandler):

    @cyclone.web.authenticated
    def get(self):
        self.post()

    @cyclone.web.authenticated
    def post(self):
        node_id = self.get_argument('node_id',None)
        account_number = self.get_argument('account_number',None)
        query_begin_time = self.get_argument('query_begin_time',None)
        query_end_time = self.get_argument('query_end_time',None)
        opr_nodes = self.get_opr_nodes()
        _query = self.db.query(
            models.TrBilling,
            models.TrCustomer.node_id,
            models.TrNode.node_name
        ).filter(
            models.TrBilling.account_number == models.TrAccount.account_number,
            models.TrCustomer.customer_id == models.TrAccount.customer_id,
            models.TrNode.id == models.TrCustomer.node_id
        )
        if node_id:
            _query = _query.filter(models.TrCustomer.node_id == node_id)
        else:
            _query = _query.filter(models.TrCustomer.node_id.in_(i.id for i in opr_nodes))
        if account_number:
            _query = _query.filter(models.TrBilling.account_number.like('%' + account_number + '%'))
        if query_begin_time:
            _query = _query.filter(models.TrBilling.create_time >= query_begin_time + ' 00:00:00')
        if query_end_time:
            _query = _query.filter(models.TrBilling.create_time <= query_end_time + ' 23:59:59')
        _query = _query.order_by(models.TrBilling.create_time.desc())

        if self.request.path == '/admin/customer/billing':
            return self.render("billing_list.html",
                          node_list=opr_nodes,
                          page_data=self.get_page_data(_query), **self.get_params())

        elif request.path == '/admin/customer/billing/export':
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
            return self.export_file(name, data)

@permit.route(r"/admin/customer/billing/export", u"用户计费日志导出",MenuUser, order=3.0001)
class CustomerBillingExportHandler(CustomerBillingHandler):
    pass
