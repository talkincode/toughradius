#!/usr/bin/env python
#coding=utf-8

import traceback
from toughlib import apiutils
from txradius.radius import dictionary
import datetime
import toughradius
import os
from txradius import authorize
from toughlib.permit import permit
from toughlib.storage import Storage
from toughradius.manage.radius import radius_acct_stop
from toughradius.manage.api.apibase import ApiHandler
from toughradius.manage import models
from toughradius.manage.settings import *

""" 客户上网账号释放MAC,VLAN绑定
"""

@permit.route(r"/api/v1/online/unlock")
class OnlineUnlockHandler(ApiHandler):
    """ @param:
        account_number: str,
        nas_addr: str,
        acct_session_id: str
    """

    dictionary = None

    def get_request(self, online):
        session_time = (datetime.datetime.now() - datetime.datetime.strptime(
            online.acct_start_time,"%Y-%m-%d %H:%M:%S")).total_seconds()
        return Storage(
            account_number = online.account_number,
            mac_addr = online.mac_addr,
            nas_addr = online.nas_addr,
            nas_port = 0,
            service_type = '',
            framed_ipaddr = online.framed_ipaddr,
            framed_netmask = '',
            nas_class = '',
            session_timeout = 0,
            calling_station_id = '',
            acct_status_type = STATUS_TYPE_STOP,
            acct_input_octets = 0,
            acct_output_octets = 0,
            acct_session_id = online.acct_session_id,
            acct_session_time = session_time,
            acct_input_packets = 0,
            acct_output_packets = 0,
            acct_terminate_cause = '',
            acct_input_gigawords = 0,
            acct_output_gigawords = 0,
            event_timestamp = datetime.datetime.now().strftime( "%Y-%m-%d %H:%M:%S"),
            nas_port_type='',
            nas_port_id=online.nas_port_id
        )

    def onSendResp(self, resp, disconnect_req):
        if self.db.query(models.TrOnline)\
                .filter_by(nas_addr=disconnect_req.nas_addr,
                           acct_session_id=disconnect_req.acct_session_id).count() > 0:
            radius_acct_stop.RadiusAcctStop(self.application.db_engine,
                                            self.application.mcache,
                                            self.application.aes,disconnect_req).acctounting()
        self.render_success(msg=u"send disconnect ok! coa resp : %s" % resp)

    def onSendError(self,err, disconnect_req):
        if self.db.query(models.TrOnline).filter_by(nas_addr=disconnect_req.nas_addr,
                                                    acct_session_id=disconnect_req.acct_session_id).count() > 0:
            radius_acct_stop.RadiusAcctStop(self.application.db_engine,
                                            self.application.mcache,
                                            self.application.aes, disconnect_req).acctounting()
        self.self.render_success(msg=u"send disconnect done! %s" % err.getErrorMessage())

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
            nas_addr = request.get('nas_addr')
            acct_session_id = request.get('acct_session_id')

            nas = self.db.query(models.TrBas).filter_by(ip_addr=nas_addr).first()
            if nas_addr and not nas:
                self.db.query(models.TrOnline).filter_by(
                    nas_addr=nas_addr,
                    acct_session_id=acct_session_id
                ).delete()
                self.db.commit()
                self.render_verify_err(msg=u"nasaddr not exists, online clear!")
                return

            online = self.db.query(models.TrOnline).filter_by(
                nas_addr=nas_addr, acct_session_id=acct_session_id).first()

            if not online:
                self.render_verify_err(msg=u"online not exists")
                return

            if not OnlineUnlockHandler.dictionary:
                OnlineUnlockHandler.dictionary = dictionary.Dictionary(
                    os.path.join(os.path.dirname(toughradius.__file__), 'dictionarys/dictionary'))

            deferd = authorize.disconnect(
                int(nas.vendor_id or 0),
                OnlineUnlockHandler.dictionary,
                nas.bas_secret,
                nas.ip_addr,
                coa_port=int(nas.coa_port or 3799),
                debug=True,
                User_Name=account_number,
                NAS_IP_Address=nas.ip_addr,
                Acct_Session_Id=acct_session_id)

            deferd.addCallback(
                self.onSendResp, self.get_request(online)).addErrback(
                self.onSendError, self.get_request(online))

            self.add_oplog(u'API强制用户下线%s:%s' % (account_number, acct_session_id))
            self.db.commit()
            self.render_success()
            return deferd

        except Exception as err:
            self.render_unknow(err)
            import traceback
            traceback.print_exc()















