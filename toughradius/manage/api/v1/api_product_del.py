#!/usr/bin/env python
#coding=utf-8
import time
import traceback
from toughradius.common.btforms import dataform
from toughradius.common.btforms import rules
from toughradius.common import utils, apiutils, dispatch
from toughradius.common.permit import permit
from toughradius.manage.api.apibase import ApiHandler
from toughradius import models

""" 删除产品套餐
"""

@permit.route(r"/api/v1/product/delete")
class ProductDeleteHandler(ApiHandler):
    """ @param: 
        product_id: str
    """

    def get(self):
        self.post()

    def post(self):
        try:
            request = self.parse_form_request()
            product_id = request.get('product_id')
            if not product_id:
                return self.render_verify_err(msg="product_id must not be NULL")

            product_qry = self.db.query(models.TrProduct)
            product = product_qry.get(product_id)

            if not product:
                return self.render_verify_err(msg="product is not exist")

            if self.db.query(models.TrAccount.product_id).filter_by(product_id=product_id).count() > 0:
                return self.render_verify_err(msg="product is using by some user,can not be deleted")

            product_name = product.product_name
            product_qry.filter_by(id=product_id).delete()
            self.add_oplog(u'API删除资费套餐:%s' % utils.safeunicode(product_name))
            self.db.commit()
            self.render_success(msg=u'API删除资费套餐:%s' % product_name)
        except Exception as err:
            self.render_unknow(err)















