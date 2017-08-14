#!/usr/bin/env python
#coding:utf-8

from toughradius.txradius.radius import dictionary
from toughradius.radiusd.radutils import parse_auth_packet
from toughradius.radiusd.radutils import parse_acct_packet
from toughradius.radiusd.radutils import process_auth_reply
from toughradius.radiusd.radutils import process_acct_reply
import gevent
import logging

class BasicAdapter(object):

    def __init__(self,config):
        self.config = config
        self.logger = logging.getLogger(__name__)
        self.dictionary = dictionary.Dictionary(config.radiusd.dictionary)
        self.clients = self.config.clients

    def handleAuth(self,socket, data, address):
        try:
            req = parse_auth_packet(data,address,self.clients,dictionary)
            prereply = self.send(req)
            gevent.sleep(0)
            reply = process_auth_reply(req, prereply)
            socket.sendto(reply.ReplyPacket(), address )
        except:
            self.logger.exception( "Handle Radius Auth error" )

    def handleAcct(self,socket, data, address):
        try:
            req = parse_acct_packet(data,address,self.clients,dictionary)
            prereply = self.send(req)
            gevent.sleep(0)
            reply = process_acct_reply(req, prereply)
            socket.sendto(reply.ReplyPacket(), address )
        except:
            self.logger.exception("Handle Radius Acct error")

    def send(self,req):
        raise NotImplementedError()



