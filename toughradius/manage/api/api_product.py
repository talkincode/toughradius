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

""" 产品套餐查询
"""

@permit.route(r"/api/product/query")
class ProductQueryHandler(ApiHandler):
    """ @param: 
        product_id: str
    """

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

            self.render_result(code=0, msg='success',products=product_datas)

        except Exception as err:
            self.render_result(code=1, msg=utils.safeunicode(err.message))
            return















