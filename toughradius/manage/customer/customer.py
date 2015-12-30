#!/usr/bin/env python
#coding=utf-8

import cyclone.auth
import cyclone.escape
import cyclone.web
import decimal
import datetime
from tablib import Dataset
from toughradius.manage import models
from toughradius.manage.base import BaseHandler
from toughradius.manage.customer import customer_forms
from toughlib.permit import permit
from toughlib import utils
from toughradius.manage.settings import * 

class CustomerHandler(BaseHandler):

    detail_url_fmt = "/admin/customer/detail?account_number={0}".format


@permit.route(r"/admin/customer", u"用户信息管理",MenuUser, order=1.0000, is_menu=True)
class CustomerListHandler(CustomerHandler):

    @cyclone.web.authenticated
    def get(self):
        self.post()

    @cyclone.web.authenticated
    def post(self):
        node_id = self.get_argument('node_id', None)
        realname = self.get_argument('realname', None)
        idcard = self.get_argument('idcard', None)
        mobile = self.get_argument('mobile', None)
        user_name = self.get_argument('user_name', None)
        status = self.get_argument('status', None)
        product_id = self.get_argument('product_id', None)
        address = self.get_argument('address', None)
        expire_days = self.get_argument('expire_days', None)
        opr_nodes = self.get_opr_nodes()
        _query = self.db.query(
            models.TrCustomer,
            models.TrAccount,
            models.TrProduct.product_name,
            models.TrNode.node_desc
        ).filter(
            models.TrProduct.id == models.TrAccount.product_id,
            models.TrCustomer.customer_id == models.TrAccount.customer_id,
            models.TrNode.id == models.TrCustomer.node_id
        )

        _now = datetime.datetime.now()

        if idcard:
            _query = _query.filter(models.TrCustomer.idcard == idcard)
        if mobile:
            _query = _query.filter(models.TrCustomer.mobile == mobile)
        if node_id:
            _query = _query.filter(models.TrCustomer.node_id == node_id)
        else:
            _query = _query.filter(models.TrCustomer.node_id.in_([i.id for i in opr_nodes]))
        if realname:
            _query = _query.filter(models.TrCustomer.realname.like('%' + realname + '%'))
        if user_name:
            _query = _query.filter(models.TrAccount.account_number.like('%' + user_name + '%'))

        #用户状态判断
        if status:
            if status == '4':
                _query = _query.filter(models.TrAccount.expire_date <= _now.strftime("%Y-%m-%d"))
            elif status == '1':
                _query = _query.filter(
                    models.TrAccount.status == status,
                    models.TrAccount.expire_date >= _now.strftime("%Y-%m-%d")
                )
            else:
                _query = _query.filter(models.TrAccount.status == status)

        if product_id:
            _query = _query.filter(models.TrAccount.product_id == product_id)
        if address:
            _query = _query.filter(models.TrCustomer.address.like('%' + address + '%'))
        if expire_days:
            _days = int(expire_days)
            td = datetime.timedelta(days=_days)
            edate = (_now + td).strftime("%Y-%m-%d")
            _query = _query.filter(models.TrAccount.expire_date <= edate)
            _query = _query.filter(models.TrAccount.expire_date >= _now.strftime("%Y-%m-%d"))

        if self.request.path == '/admin/customer':
            return self.render("customer_list.html",
                          page_data=self.get_page_data(_query),
                          node_list=opr_nodes,
                          products=self.db.query(models.TrProduct),
                          **self.get_params())
        elif self.request.path == "/admin/customer/export":
            data = Dataset()
            data.append((
                u'区域', u'姓名', u'证件号', u'邮箱', u'联系电话', u'地址',
                u'用户账号', u'密码', u'资费', u'过期时间', u'余额(元)',
                u'时长(小时)', u'流量(MB)', u'并发数', u'ip地址', u'状态', u'创建时间'
            ))
            for i, j, _product_name, _node_desc in _query:
                data.append((
                    _node_desc, i.realname, i.idcard, i.email, i.mobile, i.address,
                    j.account_number, self.aes.decrypt(j.password), _product_name,
                    j.expire_date, utils.fen2yuan(j.balance),
                    utils.sec2hour(j.time_length), utils.kb2mb(j.flow_length), j.user_concur_number, j.ip_address,
                    customer_forms.user_state[j.status], j.create_time
                ))
            name = u"RADIUS-USER-" + datetime.datetime.now().strftime("%Y%m%d-%H%M%S") + ".xls"
            self.export_file(name, data)

@permit.route(r"/admin/customer/export", u"用户导出",MenuUser, order=1.0001)
class CustomerExportHandler(CustomerListHandler):
    pass

