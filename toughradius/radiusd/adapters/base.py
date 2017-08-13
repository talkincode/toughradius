#!/usr/bin/env python
#coding:utf-8

from toughradius.txradius.radius import dictionary
from toughradius.radiusd.radutils import parse_auth_packet
from toughradius.radiusd.radutils import parse_acct_packet
from toughradius.radiusd.radutils import process_auth_reply
from toughradius.radiusd.radutils import process_acct_reply

import logging

class BasicAdapter(object):

    def __init__(self,config):
        self.config = config
        self.logger = logging.getLogger(__name__)
        self.dictionary = dictionary.Dictionary(config.radiusd.dictionary)
        self.clients = self.config.clients

    def sendReply(self,,socket, reply, (host,port)):
        self.logger.info("send reply to %s:%s"%(host,port))
        socket.sendto(reply.ReplyPacket(), (host,port))     

    def handleAuth(self,data,address):
        req = parse_auth_packet(data,address,dictionary,self.config.clients)
        prereply = self.send_rest(req)
        return process_auth_reply(req, prereply)

    def handleAcct(self,data,address):
        req = parse_acct_packet(data,address,dictionary,self.config.clients)
        prereply = self.send_rest(req)
        return process_acct_reply(req, prereply)

    def send(self,req):
        raise NotImplementedError()



