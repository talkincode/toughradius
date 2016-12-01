#!/usr/bin/env python
#coding=utf-8

from toughradius.common.btforms import rules
from toughradius.common.btforms import dataform
from toughradius.common import utils
from toughradius.common.permit import permit
from toughradius.manage.api.apibase import ApiHandler
from toughradius import models

""" 产品套餐修改
"""

product_update_vform = dataform.Form(
    dataform.Item("product_name", rules.len_of(4, 64), description=u"资费名称"),
    dataform.Item("fee_price", rules.is_rmb, description=u"资费价格(元)"),
    dataform.Item("concur_number", rules.is_numberOboveZore, description=u"并发数控制(0表示不限制)"),
    dataform.Item("bind_mac", description=u"是否绑定MAC "),
    dataform.Item("bind_vlan", description=u"是否绑定VLAN "),
    dataform.Item("input_max_limit", rules.is_number3, description=u"最大上行速率(Mbps)"),
    dataform.Item("output_max_limit", rules.is_number3, description=u"最大下行速率(Mbps)"),
    dataform.Item("product_status", description=u"资费状态"),
    title="api product update"
)

@permit.route(r"/api/v1/product/update")
class ProductUpdateHandler(ApiHandler):
    """ @param: 
        product_id: str
    """

    def get(self):
        self.post()

    def post(self):
        form = product_update_vform()
        try:
            request = self.parse_form_request()
            if not form.validates(**request):
                return self.render_verify_err(form.errors)
        except Exception, err:
            return self.render_verify_err(err)

        try:
            product_id = request.get('product_id')
            if not product_id:
                return self.render_verify_err(msg="product_id must not be NULL")

            product = self.db.query(models.TrProduct).get(product_id)

            if not product:
                return self.render_verify_err(msg="product is not exist")

            if form.d.product_name:
                product.product_name = form.d.product_name
            if form.d.fee_price:
                product.fee_price = utils.yuan2fen(form.d.fee_price)

            if form.d.concur_number:
                product.concur_number = form.d.concur_number
            if form.d.bind_mac:
                product.bind_mac = form.d.bind_mac
            if form.d.bind_vlan:
                product.bind_vlan = form.d.bind_vlan

            if form.d.input_max_limit:
                product.input_max_limit = form.d.input_max_limit
            if form.d.output_max_limit:
                product.output_max_limit = form.d.output_max_limit
            if form.d.product_status:
                product.product_status = form.d.product_status

            product.update_time = utils.get_currtime()

            self.db.commit()
            self.add_oplog(u'API修改资费信息:%s' % utils.safeunicode(product_id))
            self.render_success(msg=u'product update success')
        except Exception as err:
            self.render_unknow(err)















