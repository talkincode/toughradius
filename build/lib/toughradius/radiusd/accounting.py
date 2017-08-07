#!/usr/bin/env python
#coding:utf-8

import gevent
import logging
from toughradius.txradius import message
from toughradius.txradius.radius import packet
import importlib

logger = logging.getLogger(__name__)

Acct_Status_Start = 1
Acct_Status_Stop = 2
Acct_Status_Update = 3
Acct_Status_On = 7
Acct_Status_Off = 8


class Handler(object):

    def __init__(self,server):
        self.server = server
        self.config = self.server.config
        self.acct_filters = {
            Acct_Status_Start: self.server.modules.acctounting.start,
            Acct_Status_Stop: self.server.modules.acctounting.stop,
            Acct_Status_Update: self.server.modules.acctounting.update,
            Acct_Status_On: self.server.modules.acctounting.on,
            Acct_Status_Off: self.server.modules.acctounting.off,
        }

    def sendReply(self,reply, (host,port)):
        self.server.socket.sendto(reply.ReplyPacket(), (host,port))            

    def createAcctPacket(self, datagram,(host,port)):
        if host in self.server.clients.ips:
            client = self.server.clients.ips[host]
            acct_message = message.AcctMessage(packet=datagram, 
                dict=self.server.dictionary, 
                secret=six.b(client['secret']),
                vendor_id=self.server.clients.vendors.get(client['vendor']))
        elif self.server.clients.ids:
            acct_message = message.AcctMessage(packet=datagram, 
                dict=self.server.dictionary, secret=six.b(''),vendor_id=0)
            nas_id = acct_message.get_nas_id()
            if nas_id in self.server.clients.ids:
                client = self.server.clients.ids[nas_id]
                acct_message.vendor_id = self.server.clients.vendors.get(client['vendor'])
                acct_message.secret = six.b(client['secret'])
            else:
                raise packet.PacketError("Unauthorized Radius Access Device [%s] (%s:%s)"%(nas_id,host,port))
        else:
            raise packet.PacketError("Unauthorized Radius Access Device (%s:%s) "%(host,port))

        return acct_message


    def handle(self,data, (host,port)):
        req = self.createAcctPacket(data,(host,port))
        with gevent.Timeout(self.config.radiusd.request_timeout, True) as timeout:
            # process radius access request
            for module_cls in self.server.modules.acctounting.parse:
                mod = self.server.get_module(module_cls)
                if mod: 
                    try:
                        req = mod.handle_radius(req)       
                    except:
                        errmsg = "server handle radius acctounting parse error"
                        logger.exception(errmsg)
                        return

            for module_cls in self.acct_filters.get(req.get_acct_status_type(),[]):
                mod = self.server.get_module(module_cls)
                if mod: 
                    try:
                        req = mod.handle_radius(req)       
                    except:
                        errmsg = "server handle radius accounting error"
                        logger.exception(errmsg)
                        return

        gevent.spawn(self.sendReply,req.CreateReply(),(host,port))









