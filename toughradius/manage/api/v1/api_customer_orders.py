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

""" 客户交易记录查询
"""


@permit.route(r"/api/v1/order/query")
class CustomerAccountsHandler(ApiHandler):
    """ @param: 
        customer_name: str,
    """

    def get(self):
        self.post()

    def post(self):
        try:
            request = self.parse_form_request()
        except apiutils.SignError, err:
            return self.render_sign_err(err)
        except Exception as err:
            return self.render_parse_err(err)

        try:
            customer_name = request.get('customer_name')
            order_id = request.get('order_id')

            if order_id:
                order = self.db.query(models.TrCustomerOrder).get(order_id)
                order_data = {}
                if order:
                    order_data = { c.name : getattr(order, c.name) for c in order.__table__.columns }
                return self.render_success(order=order_data)

            if not customer_name:
                return self.render_verify_err(msg="customer_name requerd")

            order_query = self.db.query(models.TrCustomerOrder).filter(
                models.TrCustomerOrder.customer_id == models.TrCustomer.customer_id,
                models.TrCustomer.customer_name == customer_name
            ).order_by(models.TrCustomerOrder.create_time.desc())


            order_data = []
            for order in order_query:
                order_item = { c.name : getattr(order, c.name) for c in order.__table__.columns }
                order_data.append(order_item)

            self.render_success(orders=order_data)
        except Exception as err:
            self.render_unknow(err)












