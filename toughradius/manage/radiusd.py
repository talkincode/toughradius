#!/usr/bin/env python
# coding=utf-8
import datetime
import os
import six
import msgpack
import toughradius
from txzmq import ZmqEndpoint, ZmqFactory, ZmqPushConnection, ZmqPullConnection
from twisted.internet import protocol
from twisted.internet import reactor
from twisted.internet.threads import deferToThread
from twisted.internet import defer
from toughlib import utils
from toughlib import mcache
from toughlib import logger,dispatch
from toughlib import db_cache as cache
from toughlib.redis_cache import CacheManager
from toughlib.dbengine import get_engine
from txradius.radius import dictionary
from txradius.radius import packet
from txradius.radius.packet import PacketError
from txradius import message
from toughlib.utils import timecast
from toughradius.common import log_trace
from toughradius.manage import models
from toughradius.manage.settings import *
from toughradius.manage.radius.plugins import mac_parse,vlan_parse, rate_process
from toughradius.manage.radius.radius_authorize import RadiusAuth
from toughradius.manage.radius.radius_acct_start import RadiusAcctStart
from toughradius.manage.radius.radius_acct_update import RadiusAcctUpdate
from toughradius.manage.radius.radius_acct_stop import RadiusAcctStop
from toughradius.manage.radius.radius_acct_onoff import RadiusAcctOnoff



class RADIUSMaster(protocol.DatagramProtocol):
    def __init__(self, config, service='auth'):
        self.config = config
        self.service = service
        self.pusher = ZmqPushConnection(ZmqFactory(), ZmqEndpoint('bind', 'ipc:///tmp/radiusd-%s-message' % service))
        self.puller = ZmqPullConnection(ZmqFactory(), ZmqEndpoint('bind', 'ipc:///tmp/radiusd-%s-result' % service))
        self.puller.onPull = self.reply
        logger.info("init %s master pusher : %s " % (self.service, self.pusher))
        logger.info("init %s master puller : %s " % (self.service, self.puller))


    def datagramReceived(self, datagram, (host, port)):
        message = msgpack.packb([datagram, host, port])
        self.pusher.push(message)
        
    def reply(self, result):
        data, host, port = msgpack.unpackb(result[0])
        self.transport.write(data, (host, int(port)))

class RADIUSAuthWorker(protocol.DatagramProtocol):

    def __init__(self, config, dbengine, radcache=None):
        self.config = config
        self.dict = dictionary.Dictionary(
            os.path.join(os.path.dirname(toughradius.__file__), 'dictionarys/dictionary'))
        self.db_engine = dbengine or get_engine(config)
        self.aes = utils.AESCipher(key=self.config.system.secret)
        self.mcache = radcache
        self.pusher = ZmqPushConnection(ZmqFactory(), ZmqEndpoint('connect', 'ipc:///tmp/radiusd-auth-result'))
        self.stat_pusher = ZmqPushConnection(ZmqFactory(), ZmqEndpoint('connect', 'ipc:///tmp/radiusd-stat-task'))
        self.puller = ZmqPullConnection(ZmqFactory(), ZmqEndpoint('connect', 'ipc:///tmp/radiusd-auth-message'))
        self.puller.onPull = self.process
        reactor.listenUDP(0, self)
        logger.info("init auth worker pusher : %s " % (self.pusher))
        logger.info("init auth worker puller : %s " % (self.puller))
        logger.info("init auth stat pusher : %s " % (self.stat_pusher))

    def find_nas(self,ip_addr):
        def fetch_result():
            table = models.TrBas.__table__
            with self.db_engine.begin() as conn:
                return conn.execute(table.select().where(table.c.ip_addr==ip_addr)).first()
        return self.mcache.aget(bas_cache_key(ip_addr),fetch_result, expire=600)

    def do_stat(self,code):
        try:
            stat_msg = []
            if code == packet.AccessRequest:
                stat_msg.append('auth_req')
            elif code == packet.AccessAccept:
                stat_msg.append('auth_accept')
            elif  code == packet.AccessReject:
                stat_msg.append('auth_reject')
            else:
                stat_msg = ['auth_drop']
            self.stat_pusher.push(msgpack.packb(stat_msg))
        except:
            pass

    def process(self, message):
        datagram, host, port =  msgpack.unpackb(message[0])
        reply = self.processAuth(datagram, host, port)
        if not reply:
            return
        logger.info("[Radiusd] :: Send radius response: %s" % repr(reply))
        if self.config.system.debug:
            logger.debug(reply.format_str())
        self.pusher.push(msgpack.packb([reply.ReplyPacket(),host,port]))
        # self.transport.write(reply.ReplyPacket(), (host,port))
        self.do_stat(reply.code)

    def createAuthPacket(self, **kwargs):
        vendor_id = kwargs.pop('vendor_id',0)
        auth_message = message.AuthMessage(**kwargs)
        auth_message.vendor_id = vendor_id
        auth_message = mac_parse.process(auth_message)
        auth_message = vlan_parse.process(auth_message)
        return auth_message

    def processAuth(self, datagram, host, port):
        try:
            bas = self.find_nas(host)
            if not bas:
                raise PacketError('[Radiusd] :: Dropping packet from unknown host %s' % host)

            secret, vendor_id = bas['bas_secret'], bas['vendor_id']
            req = self.createAuthPacket(packet=datagram, 
                dict=self.dict, secret=six.b(str(secret)),vendor_id=vendor_id)

            # if 'trbtest' in req.get_user_name():
            #     reply = req.CreateReply()
            #     reply.vendor_id = req.vendor_id
            #     reply['Reply-Message'] = 'trbtest success!'
            #     reply.code = packet.AccessAccept
            #     return reply

            self.do_stat(req.code)

            logger.info("[Radiusd] :: Received radius request: %s" % (repr(req)))
            if self.config.system.debug:
                logger.debug(req.format_str())

            if req.code != packet.AccessRequest:
                raise PacketError('non-AccessRequest packet on authentication socket')

            reply = req.CreateReply()
            reply.vendor_id = req.vendor_id

            aaa_request = dict(
                account_number=req.get_user_name(),
                domain=req.get_domain(),
                macaddr=req.client_mac,
                nasaddr=req.get_nas_addr() or host,
                vlanid1=req.vlanid1,
                vlanid2=req.vlanid2
            )

            auth_resp = RadiusAuth(self.db_engine,self.mcache,self.aes,aaa_request).authorize()

            if auth_resp['code'] > 0:
                reply['Reply-Message'] = auth_resp['msg']
                reply.code = packet.AccessReject
                return reply

            if 'bypass' in auth_resp and int(auth_resp['bypass']) == 0:
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
                        logger.error(errstr)

                for attr, attr_val in req.resp_attrs.iteritems():
                    reply[attr] = attr_val

            reply['Reply-Message'] = 'success!'
            reply.code = packet.AccessAccept
            if not req.VerifyReply(reply):
                raise PacketError('VerifyReply error')
            return reply
        except Exception as err:
            self.do_stat(0)
            errstr = 'RadiusError:Dropping invalid auth packet from {0} {1},{2}'.format(
                host, port, utils.safeunicode(err))
            logger.error(errstr)
            import traceback
            traceback.print_exc()



class RADIUSAcctWorker:

    def __init__(self, config, dbengine,radcache=None):
        self.config = config
        self.dict = dictionary.Dictionary(
            os.path.join(os.path.dirname(toughradius.__file__), 'dictionarys/dictionary'))
        self.db_engine = dbengine or get_engine(config)
        self.mcache = radcache
        self.pusher = ZmqPushConnection(ZmqFactory(), ZmqEndpoint('connect', 'ipc:///tmp/radiusd-acct-result'))
        self.stat_pusher = ZmqPushConnection(ZmqFactory(), ZmqEndpoint('connect', 'ipc:///tmp/radiusd-stat-task'))
        self.puller = ZmqPullConnection(ZmqFactory(), ZmqEndpoint('connect', 'ipc:///tmp/radiusd-acct-message'))
        self.puller.onPull = self.process
        logger.info("init acct worker pusher : %s " % (self.pusher))
        logger.info("init acct worker puller : %s " % (self.puller))
        logger.info("init auth stat pusher : %s " % (self.stat_pusher))
        self.acct_class = {
            STATUS_TYPE_START: RadiusAcctStart,
            STATUS_TYPE_STOP: RadiusAcctStop,
            STATUS_TYPE_UPDATE: RadiusAcctUpdate,
            STATUS_TYPE_ACCT_ON: RadiusAcctOnoff,
            STATUS_TYPE_ACCT_OFF: RadiusAcctOnoff
        }

    def find_nas(self,ip_addr):
        def fetch_result():
            table = models.TrBas.__table__
            with self.db_engine.begin() as conn:
                return conn.execute(table.select().where(table.c.ip_addr==ip_addr)).first()
        return self.mcache.aget(bas_cache_key(ip_addr),fetch_result, expire=600)

    def do_stat(self,code, status_type=0):
        try:
            stat_msg = ['acct_drop']
            if code  in (4,5):
                stat_msg = []
                if code == packet.AccountingRequest:
                    stat_msg.append('acct_req')
                elif code == packet.AccountingResponse:
                    stat_msg.append('acct_resp')

                if status_type == 1:
                    stat_msg.append('acct_start')
                elif status_type == 2:
                    stat_msg.append('acct_stop')        
                elif status_type == 3:
                    stat_msg.append('acct_update')        
                elif status_type == 7:
                    stat_msg.append('acct_on')        
                elif status_type == 8:
                    stat_msg.append('acct_off')
            self.stat_pusher.push(msgpack.packb(stat_msg))
        except:
            pass

    def process(self, message):
        datagram, host, port =  msgpack.unpackb(message[0])
        self.processAcct(datagram, host, port)
        
    def createAcctPacket(self, **kwargs):
        vendor_id = 0
        if 'vendor_id' in kwargs:
            vendor_id = kwargs.pop('vendor_id')
        acct_message = message.AcctMessage(**kwargs)
        acct_message.vendor_id = vendor_id
        acct_message = mac_parse.process(acct_message)
        acct_message = vlan_parse.process(acct_message)
        return acct_message

    def processAcct(self, datagram, host, port):
        try:
            bas = self.find_nas(host)
            if not bas:
                raise PacketError('[Radiusd] :: Dropping packet from unknown host %s' % host)

            secret, vendor_id = bas['bas_secret'], bas['vendor_id']
            req = self.createAcctPacket(packet=datagram, 
                dict=self.dict, secret=six.b(str(secret)),vendor_id=vendor_id)

            self.do_stat(req.code, req.get_acct_status_type())

            logger.info("[Radiusd] :: Received radius request: %s" % (repr(req)))
            if self.config.system.debug:
                logger.debug(req.format_str())

            if req.code != packet.AccountingRequest:
                raise PacketError('non-AccountingRequest packet on authentication socket')

            if not req.VerifyAcctRequest():
                raise PacketError('VerifyAcctRequest error')

            reply = req.CreateReply()
            self.pusher.push(msgpack.packb([reply.ReplyPacket(),host,port]))
            self.do_stat(reply.code)
            logger.info("[Radiusd] :: Send radius response: %s" % repr(reply))
            if self.config.system.debug:
                logger.debug(reply.format_str())

            status_type = req.get_acct_status_type()
            if status_type in self.acct_class:
                ticket = req.get_ticket()
                if not ticket.get('nas_addr'):
                    ticket['nas_addr'] = host
                acct_func = self.acct_class[status_type](
                        self.db_engine,self.mcache,None,ticket).acctounting
                reactor.callLater(0.1,acct_func)
            else:
                logger.error('status_type <%s> not support' % status_type)
        except Exception as err:
            self.do_stat(0)
            errstr = 'RadiusError:Dropping invalid acct packet from {0} {1},{2}'.format(
                host, port, utils.safeunicode(err))
            logger.error(errstr)
            import traceback
            traceback.print_exc()

def run_auth(config):
    auth_protocol = RADIUSMaster(config, service='auth')
    reactor.listenUDP(int(config.radiusd.auth_port), auth_protocol, interface=config.radiusd.host)

def run_acct(config):
    acct_protocol = RADIUSMaster(config,service='acct')
    reactor.listenUDP(int(config.radiusd.acct_port), acct_protocol, interface=config.radiusd.host)

def run_worker(config,dbengine,**kwargs):
    _cache = kwargs.pop("cache",CacheManager(redis_conf(config),cache_name='RadiusWorkerCache-%s'%os.getpid()))
    _cache.print_hit_stat(120)
    # app event init
    if not kwargs.get('standalone'):
        dispatch.register(log_trace.LogTrace(redis_conf(config)),check_exists=True)
        event_params= dict(dbengine=dbengine, mcache=_cache, aes=kwargs.pop('aes',None))
        event_path = os.path.abspath(os.path.dirname(toughradius.manage.events.__file__))
        dispatch.load_events(event_path,"toughradius.manage.events",event_params=event_params)
    logger.info('start radius worker: %s' % RADIUSAuthWorker(config,dbengine,radcache=_cache))
    logger.info('start radius worker: %s' % RADIUSAcctWorker(config,dbengine,radcache=_cache))


