#!/usr/bin/env python
#coding=utf-8

import cyclone.auth
import cyclone.escape
import cyclone.web
import decimal
import datetime
from tablib import Dataset
from toughradius.manage import models
from toughradius.manage.customer import customer_forms
from toughradius.manage.customer.customer import CustomerHandler
from toughlib.permit import permit
from toughlib import utils
from toughradius.manage.settings import * 


@permit.route(r"/admin/customer/order", u"用户交易管理",MenuUser, order=1.5000, is_menu=True)
class CustomerOrderHandler(CustomerHandler):

    @cyclone.web.authenticated
    def get(self):
        self.post()

    @cyclone.web.authenticated
    def post(self):
        node_id = self.get_argument('node_id',None)
        product_id = self.get_argument('product_id',None)
        pay_status = self.get_argument('pay_status',None)
        account_number = self.get_argument('account_number',None)
        query_begin_time = self.get_argument('query_begin_time',None)
        query_end_time = self.get_argument('query_end_time',None)
        opr_nodes = self.get_opr_nodes()
        _query = self.db.query(
            models.TrCustomerOrder,
            models.TrCustomer.node_id,
            models.TrCustomer.realname,
            models.TrProduct.product_name,
            models.TrNode.node_name
        ).filter(
            models.TrCustomerOrder.product_id == models.TrProduct.id,
            models.TrCustomerOrder.customer_id == models.TrCustomer.customer_id,
            models.TrNode.id == models.TrCustomer.node_id
        )
        if node_id:
            _query = _query.filter(models.TrCustomer.node_id == node_id)
        else:
            _query = _query.filter(models.TrCustomer.node_id.in_([i.id for i in opr_nodes]))
        if account_number:
            _query = _query.filter(models.TrCustomerOrder.account_number.like('%' + account_number + '%'))
        if product_id:
            _query = _query.filter(models.TrCustomerOrder.product_id == product_id)
        if pay_status:
            _query = _query.filter(models.TrCustomerOrder.pay_status == pay_status)
        if query_begin_time:
            _query = _query.filter(models.TrCustomerOrder.create_time >= query_begin_time + ' 00:00:00')
        if query_end_time:
            _query = _query.filter(models.TrCustomerOrder.create_time <= query_end_time + ' 23:59:59')
        _query = _query.order_by(models.TrCustomerOrder.create_time.desc())

        if self.request.path == '/admin/customer/order':
            return self.render("order_list.html",
                          node_list=opr_nodes,
                          products=self.db.query(models.TrProduct).filter_by(product_status=0),
                          page_data=self.get_page_data(_query), **self.get_params())
        elif self.request.path == '/admin/customer/order/export':
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
            return self.export_file(name, data)



