#!/usr/bin/env python
# coding=utf-8
import os
import six
import msgpack
import toughradius
import importlib
import traceback
from twisted.internet import protocol
from twisted.internet import reactor
from twisted.internet import defer
from toughlib import utils
from toughlib import mcache
from toughlib import logger
from toughlib.dbengine import get_engine
from txradius.radius import dictionary
from txradius.radius import packet
from txradius.radius.packet import PacketError
from txradius import message
from toughradius.manage import models
from toughradius.manage import settings
from toughradius.radiusd.radius_acct_start import RadiusAcctStart
from toughradius.radiusd.radius_acct_update import RadiusAcctUpdate
from toughradius.radiusd.radius_acct_stop import RadiusAcctStop
from toughradius.radiusd.radius_acct_onoff import RadiusAcctOnoff
from toughradius.radiusd.server import DICTIONARY
from toughradius.radiusd.server import WorkerBasic
from toughradius.common import log_trace
from txzmq import (
    ZmqEndpoint, ZmqFactory, ZmqPushConnection, 
    ZmqPullConnection,ZmqREPConnection,ZmqREQConnection
)

class RADIUSAcctWorker(WorkerBasic):
    """ Billing process, billing logic, push the results of the radius protocol to 
    deal with the main process,
    Accounting is an asynchronous process, that is, every time a message is received, 
    the message is sent immediately, and then in the background to deal with billing logic.
    """

    def __init__(self, config, dbengine,radcache=None):
        self.config = config
        self.load_plugins(load_types=['radius_acct_req'])
        self.db_engine = dbengine or get_engine(config)
        self.mcache = radcache
        self.dict = dictionary.Dictionary(config.radiusd.get('dictionary',DICTIONARY))
        self.stat_pusher = ZmqPushConnection(ZmqFactory())
        self.zmqrep = ZmqREPConnection(ZmqFactory())
        self.stat_pusher.tcpKeepalive = 1
        self.zmqrep.tcpKeepalive = 1
        self.stat_pusher.addEndpoints([ZmqEndpoint('connect',config.mqproxy.task_connect)])
        self.zmqrep.addEndpoints([ZmqEndpoint('connect',config.mqproxy.acct_connect)])
        self.zmqrep.gotMessage = self.process    
        self.acct_class = {
            settings.STATUS_TYPE_START: RadiusAcctStart,
            settings.STATUS_TYPE_STOP: RadiusAcctStop,
            settings.STATUS_TYPE_UPDATE: RadiusAcctUpdate,
            settings.STATUS_TYPE_ACCT_ON: RadiusAcctOnoff,
            settings.STATUS_TYPE_ACCT_OFF: RadiusAcctOnoff
        }
        logger.info("radius acct worker %s start"%os.getpid())
        logger.info("init acct worker : %s " % (self.zmqrep))
        logger.info("init acct stat pusher : %s " % (self.stat_pusher))        



    def process(self, msgid, message):
        datagram, host, port =  msgpack.unpackb(message)
        reply = self.processAcct(datagram, host, port)
        self.zmqrep.reply(msgid, msgpack.packb([reply.ReplyPacket(),host,port]))
        
    def createAcctPacket(self, **kwargs):
        vendor_id = kwargs.pop('vendor_id',0)
        acct_message = message.AcctMessage(**kwargs)
        acct_message.vendor_id = vendor_id
        for plugin in self.acct_req_plugins:
            acct_message = plugin.plugin_func(acct_message)
        return acct_message

    def processAcct(self, datagram, host, port):
        try:
            bas = self.find_nas(host)
            if not bas:
                raise PacketError(u'Unauthorized access router %s' % host)

            secret, vendor_id = bas['bas_secret'], bas['vendor_id']
            req = self.createAcctPacket(packet=datagram, dict=self.dict, secret=six.b(str(secret)),vendor_id=vendor_id)
            self.log_trace(host,port,req)
            self.do_acct_stat(req.code, req.get_acct_status_type(),req=req)

            if req.code != packet.AccountingRequest:
                errstr = u'Invalid accounting request code=%s'%req.code
                logger.error(errstr, tag="radius_acct_drop", trace="radius",username=req.get_user_name())
                return

            if not req.VerifyAcctRequest():
                errstr = u'Account response check failed, please check the shared key is consistent'
                logger.error(errstr, tag="radius_acct_drop",trace="radius",username=req.get_user_name())
                return

            status_type = req.get_acct_status_type()
            if status_type in self.acct_class:
                ticket = req.get_ticket()
                if not ticket.get('nas_addr'):
                    ticket['nas_addr'] = host
                acct_func = self.acct_class[status_type](self.db_engine,self.mcache,None,ticket).acctounting
                reactor.callLater(0.05,acct_func)
            else:
                errstr = u'Billing type <%s> not supported' % status_type
                logger.error(errstr, tag="radius_acct_drop", trace="radius",username=req.get_user_name())
                return
                
            reply = req.CreateReply()
            reactor.callLater(0.05,self.log_trace, host,port,req,reply)
            reactor.callLater(0.05,self.do_acct_stat, reply.code)
            return reply
        except Exception as err:
            self.do_acct_stat(0)
            logger.exception(err,tag="radius_acct_drop")








