#!/usr/bin/env python
# coding=utf-8
import datetime
import os
import six
from twisted.internet import protocol
from twisted.internet import reactor
from twisted.python import log
from twisted.internet import defer
from toughlib import utils
from toughlib import mcache
from toughlib import logger
from toughlib import db_cache as cache
from toughlib.dbengine import get_engine
from txradius.radius import dictionary
from txradius.radius import packet
from txradius import message
from toughlib.utils import timecast
from toughradius.manage import models
from toughradius.manage.settings import *
from toughradius.manage.radius.plugins import mac_parse,vlan_parse, rate_process
from toughradius.manage.radius.radius_authorize import RadiusAuth
from toughradius.manage.radius.radius_acct_start import RadiusAcctStart
from toughradius.manage.radius.radius_acct_update import RadiusAcctUpdate
from toughradius.manage.radius.radius_acct_stop import RadiusAcctStop
from toughradius.manage.radius.radius_acct_onoff import RadiusAcctOnoff
import toughradius

class PacketError(Exception):
    """ packet exception"""
    pass

###############################################################################
# Basic RADIUS                                                            ####
###############################################################################

class RADIUS(protocol.DatagramProtocol):
    def __init__(self, config, log):
        self.config = config
        self.dict = dictionary.Dictionary(
            os.path.join(os.path.dirname(toughradius.__file__), 'dictionarys/dictionary'))
        self.syslog = log
        self.db_engine = get_engine(config)
        self.mcache = mcache.Mcache()

    def get_nas(self,ip_addr):
        def fetch_result():
            table = models.TrBas.__table__
            with self.db_engine.begin() as conn:
                return conn.execute(table.select().where(table.c.ip_addr==ip_addr)).first()
        return self.mcache.aget(bas_cache_key(ip_addr),fetch_result, expire=600)

    def processPacket(self, pkt,bas=None):
        pass

    def createPacket(self, **kwargs):
        raise NotImplementedError('Attempted to use a pure base class')

    @timecast
    def datagramReceived(self, datagram, (host, port)):
        try:
            bas = self.get_nas(host)
            if not bas:
                self.syslog.info('[Radiusd] :: Dropping packet from unknown host ' + host)
                return

            secret, vendor_id = bas['bas_secret'], bas['vendor_id']
            radius_request = self.createPacket(packet=datagram, 
                dict=self.dict, secret=six.b(str(secret)),vendor_id=vendor_id)

            self.syslog.info("[Radiusd] :: Received radius request: %s" % (repr(radius_request)))
            if self.config.system.debug:
                log.msg(radius_request.format_str())

            reply = self.processPacket(radius_request, bas)
            self.reply(reply, (host, port))
        except Exception as err:
            errstr = 'RadiusError:Dropping invalid packet from {0} {1},{2}'.format(
                host, port, utils.safeunicode(err))
            self.syslog.error(errstr)
            


    def reply(self, reply, (host, port)):
        self.syslog.info("[Radiusd] :: Send radius response: %s" % repr(reply))

        if self.config.system.debug:
            log.msg(reply.format_str())

        self.transport.write(reply.ReplyPacket(), (host, port))



###############################################################################
# Auth Server                                                              ####
###############################################################################
class RADIUSAccess(RADIUS):
    """ Radius Access Handler
    """

    def createPacket(self, **kwargs):
        vendor_id = kwargs.pop('vendor_id',0)
        auth_message = message.AuthMessage(**kwargs)
        auth_message.vendor_id = vendor_id
        auth_message = mac_parse.process(auth_message)
        auth_message = vlan_parse.process(auth_message)
        return auth_message

    def processPacket(self, req,bas=None):
        if req.code != packet.AccessRequest:
            raise PacketError('non-AccessRequest packet on authentication socket')

        try:
            reply = req.CreateReply()
            reply.vendor_id = req.vendor_id

            aaa_request = dict(
                username=req.get_user_name(),
                domain=req.get_domain(),
                macaddr=req.client_mac,
                nasaddr=req.get_nas_addr(),
                vlanid1=req.vlanid1,
                vlanid2=req.vlanid2
            )

            auth_resp = RadiusAuth(self,aaa_request).authorize()

            if auth_resp['code'] > 0:
                reply['Reply-Message'] = auth_resp['msg']
                reply.code = packet.AccessReject
                return reply

            if 'bypass' in auth_resp and auth_resp['bypass'] == 0:
                is_pwd_ok = True
            else:
                is_pwd_ok = req.is_valid_pwd(auth_resp.get('passwd'))

            if not is_pwd_ok:
                reply['Reply-Message'] =  "password not match"
                reply.code = packet.AccessReject
                return reply
            else:
                if u"input_rate" in auth_resp and u"output_rate" in auth_resp:
                    reply = rate_process.process(
                        reply, input_rate=auth_resp['input_rate'], output_rate=auth_resp['output_rate'])

                attrs = auth_resp.get("attrs") or {}
                for attr_name in attrs:
                    try:
                        # todo: May have a type matching problem
                        reply.AddAttribute(utils.safestr(attr_name), attrs[attr_name])
                    except Exception as err:
                        errstr = "RadiusError:current radius cannot support attribute {0},{1}".format(
                            attr_name,utils.safestr(err.message))
                        self.syslog.error(errstr)

                for attr, attr_val in req.resp_attrs.iteritems():
                    reply[attr] = attr_val

            reply['Reply-Message'] = 'success!'
            reply.code = packet.AccessAccept
            if not req.VerifyReply(reply):
                raise PacketError('VerifyReply error')
            return reply
        except Exception as err:
            reply['Reply-Message'] =  "auth failure, %s" % utils.safeunicode(err.message)
            reply.code = packet.AccessReject
            return reply


###############################################################################
# Acct Server                                                              ####
############################################################################### 

class RADIUSAccounting(RADIUS):
    """ Radius Accounting Handler
    """
    acct_class = {
        STATUS_TYPE_START: RadiusAcctStart,
        STATUS_TYPE_STOP: RadiusAcctStop,
        STATUS_TYPE_UPDATE: RadiusAcctUpdate,
        STATUS_TYPE_ACCT_ON: RadiusAcctOnoff,
        STATUS_TYPE_ACCT_OFF: RadiusAcctOnoff
    }

    def createPacket(self, **kwargs):

        vendor_id = 0
        if 'vendor_id' in kwargs:
            vendor_id = kwargs.pop('vendor_id')

        acct_message = message.AcctMessage(**kwargs)
        acct_message.vendor_id = vendor_id
        acct_message = mac_parse.process(acct_message)
        acct_message = vlan_parse.process(acct_message)
        return acct_message

    def processPacket(self, req, bas=None):
        if req.code != packet.AccountingRequest:
            raise PacketError('non-AccountingRequest packet on authentication socket')

        if not req.VerifyAcctRequest():
            raise PacketError('VerifyAcctRequest error')

        reply = req.CreateReply()
        acct_request = dict(
            acct_status_type=req.get_acct_status_type(),
            username=req.get_user_name(),
            session_id=req.get_acct_sessionid(),
            session_time=req.get_acct_sessiontime(),
            session_timeout=req.get_session_timeout(),
            macaddr=req.get_mac_addr(),
            nasaddr=req.get_nas_addr(),
            ipaddr=req.get_framed_ipaddr(),
            input_octets=req.get_input_total(),
            output_octets=req.get_output_total(),
            input_pkts=req.get_acct_input_packets(),
            output_pkts=req.get_acct_output_packets(),
            nas_port=req.get_nas_port(),
            event_timestamp=req.get_event_timestamp(),
            nas_port_type=req.get_nas_port_type(),
            nas_port_id=req.get_nas_portid()
        )
        if acct_request['acct_status_type'] in RADIUSAccounting.acct_class:
            RADIUSAccounting.acct_class[acct_request['acct_status_type']](self,acct_request).acctounting()
        return reply

###############################################################################
# Radius  Run                                                              ####
###############################################################################

def run_auth(config, log=None):
    auth_protocol = RADIUSAccess(config, log)
    reactor.listenUDP(int(config.radiusd.auth_port), auth_protocol, interface=config.radiusd.host)

def run_acct(config, log=None):
    acct_protocol = RADIUSAccounting(config, log)
    reactor.listenUDP(int(config.radiusd.acct_port), acct_protocol, interface=config.radiusd.host)

