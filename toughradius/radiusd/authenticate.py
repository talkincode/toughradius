#!/usr/bin/env python
#coding:utf-8

import gevent
import logging
import six
from toughradius.txradius import message
from toughradius.txradius.radius import packet
import importlib

logger = logging.getLogger(__name__)


class Handler(object):

    def __init__(self,server):
        self.server = server
        self.config = self.server.config


    def createAuthPacket(self, datagram,(host,port)):
        if host in self.server.clients.ips:
            client = self.server.clients.ips[host]
            auth_message = message.AuthMessage(packet=datagram, 
                dict=self.server.dictionary, 
                secret=six.b(client.secret),
                vendor_id=self.server.clients.vendors.get(client.vendor))
        elif self.server.clients.ids:
            auth_message = message.AuthMessage(packet=datagram, 
                dict=self.server.dictionary, secret=six.b(''),vendor_id=0)
            nas_id = auth_message.get_nas_id()
            if nas_id in self.server.clients.ids:
                client = self.server.clients.ids[nas_id]
                auth_message.vendor_id = self.server.clients.vendors.get(client.vendor)
            else:
                raise packet.PacketError("Unauthorized Radius Access Device [%s] (%s:%s)"%(nas_id,host,port))
        else:
            raise packet.PacketError("Unauthorized Radius Access Device (%s:%s) "%(host,port))

        return auth_message

    def freeReply(self,req):
        reply = req.CreateReply()
        reply.vendor_id = req.vendor_id
        reply['Reply-Message'] = u'User:%s (Free)Authenticate Success' % req.get_user_name()
        reply.code = packet.AccessAccept        
        reply_attrs = {'attrs':{}}
        reply_attrs['input_rate'] = self.server.config.radiusd.get("free_auth_input_limit",1048576)
        reply_attrs['output_rate'] = self.server.config.radiusd.get("free_auth_output_limit",4194304)
        reply_attrs['rate_code'] = self.server.config.radiusd.get("free_auth_limit_code","")
        reply_attrs['domain'] = self.server.config.radiusd.get("free_auth_limit_code","")
        reply_attrs['attrs']['Session-Timeout'] = self.server.config.radiusd.get("max_session_timeout",86400)
        reply.resp_attrs = reply_attrs
        with gevent.Timeout(5, True) as timeout:
            for module_cls in self.modules.authorization:
                mod = self.get_module(module_cls)
                if mod: 
                    reply = mod.handle_radius(req,reply)
            return reply

    def rejectReply(self,req,errmsg=''):
        reply = req.CreateReply()
        reply.vendor_id = req.vendor_id
        reply['Reply-Message'] = errmsg
        reply.code = packet.AccessReject
        return reply

    def sendReply(self,reply, (host,port)):
        self.server.socket.sendto(reply.ReplyPacket(), (host,port))     


    def handle(self,data, (host,port)):
        req = self.createAuthPacket(data,(host,port))
        # pass user and password
        if self.server.config.radiusd.pass_userpwd:
            reply = self.freeReply(req)
            gevent.spawn(self.sendReply,reply,(host,port))
            return
            
        reply = req.CreateReply()
        reply.vendor_id = req.vendor_id

        with gevent.Timeout(self.config.radiusd.request_timeout, True) as timeout:
            # process radius access request
            for module_cls in self.server.modules.authentication:
                mod = self.server.get_module(module_cls)
                if mod: 
                    try:
                        req = mod.handle_radius(req)       
                    except:
                        errmsg = "server handle radius authentication error"
                        logger.exception(errmsg)
                        reply = self.self.rejectReply(req,errmsg)
                        gevent.spawn(self.sendReply,reply,(host,port))
                        return

            for module_cls in self.server.modules.authorization:
                mod = self.server.get_module(module_cls)
                if mod: 
                    try:
                        reply = mod.handle_radius(req,reply)       
                    except:
                        errmsg = "server handle radius authorization error"
                        logger.exception(errmsg)
                        reply = self.self.rejectReply(req,errmsg)
                        gevent.spawn(self.sendReply,reply,(host,port))
                        return

        if not req.VerifyReply(reply):
            errstr = u'[User:%s] The authentication message failed to check. \
            Check that the shared key is consistent'% username
            logger.error(errstr)
            return

        gevent.spawn(self.sendReply,reply,(host,port))

        



















