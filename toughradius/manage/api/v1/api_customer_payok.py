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
from hashlib import md5

""" 客户交易确认
"""

@permit.route(r"/api/v1/order/payok")
class CustomerAuthHandler(ApiHandler):
    """ @param: 
        order_id: str, 
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
            order_id = request.get('order_id')
            if not order_id:
                return self.render_verify_err(msg=u"order_id is empty")

            order = self.db.query(models.TrCustomerOrder).get(order_id)
            if not order:
                return self.render_verify_err(msg=u'order not exists')
            account = self.db.query(models.TrAccount).get(order.account_number)
            if not account:
                return self.render_verify_err(msg=u'account not exists')

            order.pay_status = 1
            account.status = 1
            order.order_desc = order.order_desc + u"paytime:" + utils.get_currtime()
            self.db.commit()
        except Exception as err:
            self.render_unknow(err)
            import traceback
            traceback.print_exc()
         















