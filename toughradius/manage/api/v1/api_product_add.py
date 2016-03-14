#!/usr/bin/env python
#coding=utf-8
import time
import traceback
from toughlib.btforms import dataform
from toughlib.btforms import rules
from toughlib import utils, apiutils, dispatch
from toughlib.permit import permit
from toughradius.manage.api.apibase import ApiHandler
from toughradius.manage import models

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
        try:
            request = self.parse_form_request()
            product_id = request.get('product_id')
            products = self.db.query(models.TrProduct)
            if product_id:
                products = products.filter_by(id=product_id)

            product_datas = []
            excludes = ['fee_period']
            for product in products:
                product_data = { c.name : getattr(product, c.name) \
                        for c in product.__table__.columns if c.name not in excludes}
                product_datas.append(product_data)

            self.render_success(products=product_datas)
        except Exception as err:
            self.render_unknow(err)















