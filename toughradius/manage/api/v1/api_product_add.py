#!/usr/bin/env python
# coding=utf-8

from toughradius.common.btforms import dataform
from toughradius.common.btforms import rules
from toughradius.common import utils, apiutils, dispatch
from toughradius.common.permit import permit
import datetime
from toughradius.manage.api.apibase import ApiHandler
from toughradius import models

""" 添加产品套餐
"""
product_add_vform = dataform.Form(
    dataform.Item("product_name", rules.len_of(4, 64), description=u"资费名称"),
    dataform.Item("product_policy", description=u"计费策略"),
    dataform.Item("fee_months", rules.is_number, description=u"买断授权月数"),
    dataform.Item("fee_times", rules.is_number3, description=u"买断时长(小时)"),
    dataform.Item("fee_flows", rules.is_number3, description=u"买断流量(MB)"),
    dataform.Item("fee_price", rules.is_rmb, description=u"资费价格(元)"),
    # dataform.Hidden("fee_period", rules.is_period, description=u"开放认证时段", **input_style),
    dataform.Item("concur_number", rules.is_numberOboveZore, description=u"并发数控制(0表示不限制)"),
    dataform.Item("bind_mac", description=u"是否绑定MAC "),
    dataform.Item("bind_vlan", description=u"是否绑定VLAN "),
    dataform.Item("input_max_limit", rules.is_number3, description=u"最大上行速率(Mbps)"),
    dataform.Item("output_max_limit", rules.is_number3, description=u"最大下行速率(Mbps)"),
    dataform.Item("product_status", description=u"资费状态"),
    title="api product add"
)


@permit.route(r"/api/v1/product/add")
class ProductAddHandler(ApiHandler):

    def get(self):
        self.post()

    def post(self):

        form = product_add_vform()
        try:
            request = self.parse_form_request()
        except apiutils.SignError, err:
            return self.render_sign_err(err)
        except Exception as err:
            return self.render_parse_err(err)

        try:
            if not form.validates(**request):
                return self.render_verify_err(form.errors)

            if self.db.query(models.TrProduct).filter_by(product_name=form.d.product_name).count() > 0:
                return self.render_verify_err(msg=u"product name already exists")
        except Exception, err:
            return self.render_verify_err(err)

        try:
            product = models.TrProduct()
            product.product_name = form.d.product_name
            product.product_policy = form.d.product_policy
            product.product_status = form.d.product_status
            product.fee_months = int(form.d.get("fee_months", 0))
            product.fee_times = utils.hour2sec(form.d.get("fee_times", 0))
            product.fee_flows = utils.mb2kb(form.d.get("fee_flows", 0))
            product.bind_mac = form.d.bind_mac
            product.bind_vlan = form.d.bind_vlan
            product.concur_number = form.d.concur_number
            product.fee_price = utils.yuan2fen(form.d.fee_price)
            product.fee_period = ''  # form.d.fee_period or ''
            product.input_max_limit = utils.mbps2bps(form.d.input_max_limit)
            product.output_max_limit = utils.mbps2bps(form.d.output_max_limit)
            _datetime = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")
            product.create_time = _datetime
            product.update_time = _datetime
            self.db.add(product)
            self.add_oplog(u'API新增资费信息:%s' % utils.safeunicode(product.product_name))
            self.db.commit()
            return self.render_success(msg=u'资费添加成功')
        except Exception as e:
            return self.render_verify_err(e)














