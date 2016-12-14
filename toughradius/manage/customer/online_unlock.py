#!/usr/bin/env python
#coding=utf-8
import os
import cyclone.web
import datetime
from toughradius import models
from toughradius.manage.base import BaseHandler
from toughradius.common.permit import permit
from toughradius.common import utils
from toughradius.common.storage import Storage
from toughradius.radiusd import radius_acct_stop
from toughradius import settings 
from txradius import authorize
from txradius.radius import dictionary
import toughradius

@permit.route(r"/admin/customer/online/unlock", u"用户在线解锁",settings.MenuUser, order=4.0001)
class CustomerOnlineUnlockHandler(BaseHandler):

    dictionary = None

    def get_request(self,online):
        session_time = (datetime.datetime.now() - datetime.datetime.strptime(
            online.acct_start_time,"%Y-%m-%d %H:%M:%S")).total_seconds()
        return Storage(
            account_number = online.account_number,
            mac_addr = online.mac_addr,
            nas_addr = online.nas_addr,
            nas_port = 0,
            service_type = 0,
            framed_ipaddr = online.framed_ipaddr,
            framed_netmask = '',
            nas_class = '',
            session_timeout = 0,
            calling_station_id = '00:00:00:00:00:00',
            acct_status_type = STATUS_TYPE_STOP,
            acct_input_octets = 0,
            acct_output_octets = 0,
            acct_session_id = online.acct_session_id,
            acct_session_time = session_time,
            acct_input_packets = 0,
            acct_output_packets = 0,
            acct_terminate_cause = 1,
            acct_input_gigawords = 0,
            acct_output_gigawords = 0,
            event_timestamp = datetime.datetime.now().strftime( "%Y-%m-%d %H:%M:%S"),
            nas_port_type=0,
            nas_port_id=online.nas_port_id
        )

    def onSendResp(self, resp, disconnect_req):
        if self.db.query(models.TrOnline).filter_by(
            nas_addr=disconnect_req.nas_addr, acct_session_id=disconnect_req.acct_session_id).count() > 0:
            radius_acct_stop.RadiusAcctStop(self.application.db_engine,
                                            self.application.mcache,
                                            self.application.aes,disconnect_req).acctounting()
        self.render_json(code=0, msg=u"send disconnect ok! coa resp : %s" % resp)

    def onSendError(self,err, disconnect_req):
        if self.db.query(models.TrOnline).filter_by(
            nas_addr=disconnect_req.nas_addr, acct_session_id=disconnect_req.acct_session_id).count() > 0:
            radius_acct_stop.RadiusAcctStop(self.application.db_engine,
                                            self.application.mcache,
                                            self.application.aes, disconnect_req).acctounting()
        self.render_json(code=0, msg=u"send disconnect done! %s" % err.getErrorMessage())

    @cyclone.web.authenticated
    def get(self):
        self.post()

    @cyclone.web.authenticated
    def post(self):
        username = self.get_argument('username',None)
        nas_addr = self.get_argument('nas_addr',None)
        session_id = self.get_argument('session_id',None)
        nas = self.db.query(models.TrBas).filter_by(ip_addr=nas_addr).first()
        if nas_addr and not nas:
            self.db.query(models.TrOnline).filter_by(
                nas_addr=nas_addr,
                acct_session_id=session_id
            ).delete()
            self.db.commit()
            self.render_json(code=1,msg=u"nasaddr not exists, online clear!")
            return

        if nas_addr and not session_id:
            onlines = self.db.query(models.TrOnline).filter_by(nas_addr=nas_addr)
            for online in onlines:
                radius_acct_stop.RadiusAcctStop(self.application.db_engine,
                                                self.application.mcache,
                                                self.application.aes, self.get_request(online)).acctounting()
            self.render_json(code=1,msg=u"unlock all online done!")
            return


        online = self.db.query(models.TrOnline).filter_by(
            nas_addr=nas_addr, acct_session_id=session_id).first()

        if not online:
            self.render_json(code=1,msg=u"online not exists")
            return


        if not CustomerOnlineUnlockHandler.dictionary:
            CustomerOnlineUnlockHandler.dictionary = dictionary.Dictionary(
            os.path.join(os.path.dirname(toughradius.__file__), 'dictionarys/dictionary'))

        deferd = authorize.disconnect(
            int(nas.vendor_id or 0), 
            CustomerOnlineUnlockHandler.dictionary, 
            nas.bas_secret, 
            nas.ip_addr, 
            coa_port=int(nas.coa_port or 3799), 
            debug=True,
            User_Name=username,
            NAS_IP_Address=nas.ip_addr,
            Acct_Session_Id=session_id)


        deferd.addCallback(
            self.onSendResp, self.get_request(online)).addErrback(
            self.onSendError, self.get_request(online))

        return deferd








