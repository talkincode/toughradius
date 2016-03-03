#!/usr/bin/env python
# coding=utf-8

from toughlib import utils,apiutils
from toughlib.permit import permit
from toughradius.manage.api.apibase import ApiHandler
from toughradius.manage import models
from toughradius.manage.radius.radius_acct_start import RadiusAcctStart
from toughradius.manage.radius.radius_acct_update import RadiusAcctUpdate
from toughradius.manage.radius.radius_acct_stop import RadiusAcctStop
from toughradius.manage.radius.radius_acct_onoff import RadiusAcctOnoff
from toughradius.manage.settings import *


@permit.route(r"/api/v1/acctounting")
class AcctountingHandler(ApiHandler):

    acct_class = {
        STATUS_TYPE_START: RadiusAcctStart,
        STATUS_TYPE_STOP: RadiusAcctStop,
        STATUS_TYPE_UPDATE: RadiusAcctUpdate,
        STATUS_TYPE_ACCT_ON: RadiusAcctOnoff,
        STATUS_TYPE_ACCT_OFF: RadiusAcctOnoff
    }


    def post(self):
        try:
            req_msg = self.parse_request()
            if req_msg['acct_status_type'] in AcctountingHandler.acct_class:
                acctcls =  AcctountingHandler.acct_class[req_msg.acct_status_type] 
                app = self.application
                acctcls(app.db_engine,app.mcache,app.aes,req_msg).acctounting()
            self.render_result(code=0,msg='done')
        except Exception as err:
            self.render_result(code=1, msg=utils.safeunicode(err))
            return

        



