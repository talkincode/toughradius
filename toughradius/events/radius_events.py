#!/usr/bin/env python
# coding=utf-8
from twisted.internet import reactor
from toughlib import  utils,dispatch,logger
from txradius import statistics
from toughlib.dbutils import make_db
from toughradius import models
from sqlalchemy.orm import scoped_session, sessionmaker
from toughradius.manage.settings import *
from toughlib import db_cache as cache
from toughlib.storage import Storage
from txradius.radius import dictionary
from toughradius.events.event_basic import BasicEvent
from toughradius.radius import radius_acct_stop
from txradius import authorize
import toughradius
import decimal
import datetime
import json
import os


class RadiusEvents(BasicEvent):

    dictionary = dictionary.Dictionary(
        os.path.join(os.path.dirname(toughradius.__file__), 'dictionarys/dictionary'))

    def get_request(self, online):
        if not online:
            return None
        session_time = (datetime.datetime.now() - datetime.datetime.strptime(
            online.acct_start_time,"%Y-%m-%d %H:%M:%S")).total_seconds()
        return Storage(
            account_number = online.account_number,
            mac_addr = online.mac_addr,
            nas_addr = online.nas_addr,
            nas_port = 0,
            service_type = 2,
            framed_ipaddr = online.framed_ipaddr,
            framed_netmask = '',
            nas_class = '',
            session_timeout = 0,
            calling_station_id = '',
            acct_status_type = STATUS_TYPE_STOP,
            acct_input_octets = 0,
            acct_output_octets = 0,
            acct_terminate_cause=6,
            acct_session_id = online.acct_session_id,
            acct_session_time = session_time,
            acct_input_packets = online.input_total * 1024,
            acct_output_packets = online.output_total * 1024,
            acct_input_gigawords = 0,
            acct_output_gigawords = 0,
            event_timestamp = datetime.datetime.now().strftime( "%Y-%m-%d %H:%M:%S"),
            nas_port_type=16,
            nas_port_id=online.nas_port_id
        )

    def onSendResp(self, resp, disconnect_req):
        if disconnect_req and self.db.query(models.TrOnline).filter_by(
                nas_addr=disconnect_req.nas_addr,
                acct_session_id=disconnect_req.acct_session_id).count() > 0:
            radius_acct_stop.RadiusAcctStop(
                self.dbengine,self.mcache,self.aes,disconnect_req).acctounting()
        logger.info(u"send disconnect ok! coa resp : %s" % resp)

    def onSendError(self,err, disconnect_req):
        if disconnect_req and self.db.query(models.TrOnline).filter_by(
                nas_addr=disconnect_req.nas_addr,
                acct_session_id=disconnect_req.acct_session_id).count() > 0:
            radius_acct_stop.RadiusAcctStop(
                self.dbengine,self.mcache,self.aes, disconnect_req).acctounting()
        logger.error(u"send disconnect done! %s" % err.getErrorMessage())

    def event_unlock_online(self, account_number, nas_addr, acct_session_id):
        logger.info("event unlock online [username:{0}] {1} {2}".format(account_number, nas_addr, acct_session_id))
        nas = self.db.query(models.TrBas).filter_by(ip_addr=nas_addr).first()
        if nas_addr and  not nas:
            self.db.query(models.TrOnline).filter_by(
                nas_addr=nas_addr,acct_session_id=acct_session_id).delete()
            self.db.commit()
            return

        online = self.db.query(models.TrOnline).filter_by(
                nas_addr=nas_addr, acct_session_id=acct_session_id).first()

        authorize.disconnect(
            int(nas.vendor_id or 0),
            self.dictionary,
            nas.bas_secret,
            nas.ip_addr,
            coa_port=int(nas.coa_port or 3799),
            debug=True,
            User_Name=account_number,
            NAS_IP_Address=nas.ip_addr,
            Acct_Session_Id=acct_session_id,
            Framed_IP_Address=online.framed_ipaddr,
        ).addCallback(
            self.onSendResp, self.get_request(online)).addErrback(
            self.onSendError, self.get_request(online))



    def event_clear_online(self, account_number,nas_addr, acct_session_id):
        try:
            with make_db(self.db) as db:
                logger.info("event clear expire online [username:{0}] {1} {2}".format(account_number, nas_addr, acct_session_id))
                nas = db.query(models.TrBas).filter_by(ip_addr=nas_addr).first()
                if nas_addr and  not nas:
                    db.query(models.TrOnline).filter_by(
                        nas_addr=nas_addr,acct_session_id=acct_session_id).delete()
                    db.commit()                
                    return

                online = db.query(models.TrOnline).filter_by(
                        nas_addr=nas_addr, acct_session_id=acct_session_id).first()            
                clear_req = self.get_request(online)
                radius_acct_stop.RadiusAcctStop(self.dbengine,self.mcache,self.aes,clear_req).acctounting()
                logger.info(u"系统触发用户过期清理成功: [username:%s] OnlineInfo: %s "%(str(account_number), json.dumps(clear_req)),trace="event" )
        except Exception as err:
            logger.exception(err)    


def __call__(dbengine=None, mcache=None, **kwargs):
    return RadiusEvents(dbengine=dbengine, mcache=mcache, **kwargs)

