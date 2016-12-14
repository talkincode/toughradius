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
from toughradius.common import utils
from toughradius.common import mcache
from toughradius.common import logger,dispatch
from toughradius.common.dbengine import get_engine
from txradius.radius import dictionary
from txradius.radius import packet
from txradius.radius.packet import PacketError
from txradius import message
from toughradius.radiusd.radius_authorize import RadiusAuth
from toughradius.radiusd.server import DICTIONARY
from toughradius.common import log_trace
from toughradius.radiusd.server import WorkerBasic
from txzmq import (
    ZmqEndpoint, ZmqFactory, ZmqPushConnection, 
    ZmqPullConnection,ZmqREPConnection,ZmqREQConnection
)

class RADIUSAuthWorker(WorkerBasic):
    """ The authentication process, processing authentication authorization logic, 
    push the result to the radius protocol to process the main process
    """

    def __init__(self, config, dbengine, radcache=None):
        self.config = config
        self.load_plugins(load_types=['radius_auth_req','radius_accept'])
        self.dict = dictionary.Dictionary(config.radiusd.get('dictionary',DICTIONARY))
        self.db_engine = dbengine or get_engine(config)
        self.aes = utils.AESCipher(key=config.system.secret)
        self.mcache = radcache
        self.stat_pusher = ZmqPushConnection(ZmqFactory())
        self.zmqrep = ZmqREPConnection(ZmqFactory())
        self.stat_pusher.tcpKeepalive = 1
        self.zmqrep.tcpKeepalive = 1
        self.stat_pusher.addEndpoints([ZmqEndpoint('connect',config.mqproxy.task_connect)])
        self.zmqrep.addEndpoints([ZmqEndpoint('connect',config.mqproxy.auth_connect)])
        self.zmqrep.gotMessage = self.process       
        self.reject_debug = int(self.get_param_value('radius_reject_debug',0)) == 1 
        logger.info("radius auth worker %s start"%os.getpid())
        logger.info("init auth worker : %s " % (self.zmqrep))
        logger.info("init auth stat pusher : %s " % (self.stat_pusher))

    def process(self, msgid, message):
        datagram, host, port =  msgpack.unpackb(message)
        reply = self.processAuth(datagram, host, port)

        if not reply:
            return

        if self.config.system.debug:
            logger.debug(reply.format_str())

        self.zmqrep.reply(msgid, msgpack.packb([reply.ReplyPacket(),host,port]))
        self.do_auth_stat(reply.code)

    def createAuthPacket(self, **kwargs):
        vendor_id = kwargs.pop('vendor_id',0)
        auth_message = message.AuthMessage(**kwargs)
        auth_message.vendor_id = vendor_id
        for plugin in self.auth_req_plugins:
            auth_message = plugin.plugin_func(auth_message)
        return auth_message

    def freeReply(self,req):
        reply = req.CreateReply()
        reply.vendor_id = req.vendor_id
        reply['Reply-Message'] = u'User:%s (free authentication) success' % req.get_user_name()
        reply.code = packet.AccessAccept        
        reply_attrs = {'attrs':{}}
        reply_attrs['input_rate'] = int(self.get_param_value("radius_free_input_rate",1048576))
        reply_attrs['output_rate'] = int(self.get_param_value("radius_free_output_rate",4194304))
        reply_attrs['rate_code'] = self.get_param_value("radius_free_rate_code","freerate")
        reply_attrs['domain'] = self.get_param_value("radius_free_domain","freedomain")
        reply_attrs['attrs']['Session-Timeout'] = int(self.get_param_value("radius_max_session_timeout",86400))
        for plugin in self.auth_accept_plugins:
            reply = plugin.plugin_func(reply, reply_attrs)
        return reply

    def rejectReply(self,req,errmsg=''):
        reply = req.CreateReply()
        reply.vendor_id = req.vendor_id
        reply['Reply-Message'] = errmsg
        reply.code = packet.AccessReject
        return reply

    def processAuth(self, datagram, host, port):
        try:
            bas = self.find_nas(host)
            if not bas:
                raise PacketError(u'Unauthorized access router %s' % host)
                
            secret, vendor_id = bas['bas_secret'], bas['vendor_id']
            req = self.createAuthPacket(
                packet=datagram, 
                dict=self.dict, 
                secret=six.b(str(secret)),
                vendor_id=vendor_id
            )

            username = req.get_user_name()
            bypass = int(self.get_param_value('radius_bypass',1)) 
            
            if req.code != packet.AccessRequest:
                errstr = u'Illegal authentication request, code=%s' % req.code
                logger.error(errstr, tag="radius_auth_drop", trace="radius",username=username) 
                return

            self.log_trace(host,port,req)
            self.do_auth_stat(req.code)        

            if bypass == 2:
                reply = self.freeReply(req)
                self.log_trace(host,port,req,reply)
                return reply               

            if not self.user_exists(username):
                errmsg = u'Authentication User:%s Non-existent' % username
                reply = self.rejectReply(req,errmsg)
                self.log_trace(host,port,req,reply)
                return reply                

            bind_nas_list = self.get_account_bind_nas(username)
            if not bind_nas_list or host not in bind_nas_list:
                errmsg = u'Nas:%s Not bound user:%s node' % (host,username)
                reply = self.rejectReply(req,errmsg)
                self.log_trace(host,port,req,reply)
                return reply

            aaa_request = dict(
                account_number=username,
                domain=req.get_domain(),
                macaddr=req.client_mac,
                nasaddr=req.get_nas_addr(),
                vlanid1=req.vlanid1,
                vlanid2=req.vlanid2,
                bypass=bypass,
                radreq=req
            )

            auth_resp = RadiusAuth(self.db_engine,self.mcache,self.aes,aaa_request).authorize()

            if auth_resp['code'] > 0:
                reply = self.rejectReply(req,auth_resp['msg'])
                self.log_trace(host,port,req,reply)
                return reply

            reply = req.CreateReply()
            reply.code = packet.AccessAccept            
            reply.vendor_id = req.vendor_id
            extmsg = u'Domain=%s;' % auth_resp['domain'] if 'domain' in auth_resp else ''
            extmsg += u'Limiting=%s;' % auth_resp['rate_code'] if 'rate_code' in auth_resp else ''
            reply['Reply-Message'] = u'User:%s successful authentication; %s'%(username,extmsg)
            for plugin in self.auth_accept_plugins:
                reply = plugin.plugin_func(reply, auth_resp)

            if not req.VerifyReply(reply):
                errstr = u'User:%s Authentication message authentication failed, please check the shared key is consistent'% username
                logger.error(errstr, tag="radius_auth_drop", trace="radius",username=username)
                return

            self.log_trace(host,port,req,reply)
            return reply

        except Exception as err:
            self.do_auth_stat(0)
            logger.exception(err,tag="radius_auth_error")


