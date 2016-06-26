#!/usr/bin/env python
#coding=utf-8
import time
import traceback
from toughlib import utils, apiutils
from toughlib import logger
from toughlib.permit import permit
from toughradius.manage.api.apibase import ApiHandler
from toughradius.manage import models


@permit.route(r"/api/v1/account/billing")
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

            if not account_number:
                return self.render_verify_err(msg="account_number must")

            _query = self.db.query(
                models.TrTicket,
            ).filter(
                models.TrTicket.account_number == models.TrAccount.account_number,
                models.TrCustomer.customer_id == models.TrAccount.customer_id
            )
            _query = _query.filter(models.TrBilling.account_number.like('%' + account_number + '%')) 
            _query = _query.order_by(models.TrTicket.acct_start_time.desc())
            billing_datas = []
            for history in _query:
                billing_data = {c.name: getattr(history, c.name) for c in history.__table__.columns if c.name in ['acct_start_time','acct_stop_time','acct_output_octets','acct_input_octets','mac_addr']}
                billing_datas.append(billing_data)

            self.render_success(history=billing_datas)
        except Exception as err:
            logger.error(u"api query online error, %s" % utils.safeunicode(traceback.format_exc()))
            self.render_result(code=1, msg=u"api error")



