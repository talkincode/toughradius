#!/usr/bin/env python
#coding=utf-8
import time
import traceback
from toughradius.common import utils, apiutils
from toughradius.common import logger
from toughradius.common.permit import permit
from toughradius.manage.api.apibase import ApiHandler
from toughradius import models


@permit.route(r"/api/v1/online/query")
class OnlineQueryHandler(ApiHandler):

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
            account_number = request.get('account_number')
            customer_name = request.get('customer_name')

            if not any([customer_name, account_number]):
                return self.render_verify_err(msg="customer_name,account_number must one")

            onlines=[]
            if account_number:
                onlines = self.db.query(models.TrOnline)\
                    .filter(models.TrCustomer.customer_id == models.TrAccount.customer_id,
                            models.TrAccount.account_number == models.TrOnline.account_number,
                            models.TrOnline.account_number == account_number)

            else:
                onlines = self.db.query(models.TrOnline)\
                    .filter(models.TrCustomer.customer_id == models.TrAccount.customer_id,
                            models.TrAccount.account_number == models.TrOnline.account_number,
                            models.TrCustomer.customer_name == customer_name)

            online_datas = []
            for online in onlines:
                online_data = {c.name: getattr(online, c.name) for c in online.__table__.columns}
                online_datas.append(online_data)

            self.render_success(onlines=online_datas)
        except Exception as err:
            logger.error(u"api query online error, %s" % utils.safeunicode(traceback.format_exc()))
            self.render_result(code=1, msg=u"api error")



