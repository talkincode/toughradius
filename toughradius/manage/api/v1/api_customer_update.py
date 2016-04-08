#!/usr/bin/env python
#coding=utf-8
import traceback
from toughlib import utils, apiutils
from hashlib import md5
from toughlib.permit import permit
from toughradius.manage.api.apibase import ApiHandler
from toughradius.manage import models

""" 客户资料修改，修改客户资料
"""

@permit.route(r"/api/v1/customer/update")
class CustomerUpdateHandler(ApiHandler):
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

            if not customer_name:
                return self.render_verify_err(msg="customer_name is empty")

            customer = self.db.query(models.TrCustomer).filter_by(customer_name=customer_name).first()

            if not customer:
                return self.render_verify_err(msg="customer %s is not existed" % customer_name)

            password = request.get("password")
            realname = request.get("realname")
            idcard = request.get("idcard")
            email = request.get("email")
            mobile = request.get("mobile")
            address = request.get("address")

            if password:
                customer.password = md5(password.encode()).hexdigest()

            if realname:
                customer.realname = realname

            if idcard:
                customer.idcard = idcard

            if email:
                customer.email = email

            if mobile:
                customer.mobile = mobile

            if address:
                customer.address = address

            self.add_oplog(u'修改用户资料 %s' % customer_name)
            self.db.commit()
            self.render_success()
        except Exception as err:
            self.render_unknow(err)
            import traceback
            traceback.print_exc()
            return















