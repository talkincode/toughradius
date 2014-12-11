#!/usr/bin/env python
#coding=utf-8

from twisted.internet import protocol
from twisted.internet import reactor
from twisted.python import log
from pyrad import dictionary
from pyrad import host
from pyrad import packet
from store import store
import middlewares
import yaml
import six
import sys
import pprint
import utils

class PacketError(Exception):
    """Exception class for bogus packets

    PacketError exceptions are only used inside the Server class to
    abort processing of a packet.
    """

class RADIUS(host.Host, protocol.DatagramProtocol):
    def __init__(self, config="config.yaml", dict=dictionary.Dictionary("res/dictionary")):
        host.Host.__init__(self,dict=dict)
        with open(config) as cf:
            self.config = yaml.load(cf)
        self.debug = self.config['radiusd']['debug']
        self.hosts = self.config['radiusd']['hosts']

    def processPacket(self, pkt):
        pass

    def createPacket(self, **kwargs):
        raise NotImplementedError('Attempted to use a pure base class')


    def datagramReceived(self, datagram, (host, port)):
        bas_host = [ vh for vh in self.hosts.values() if vh['addr']==host ]
        bas_host = bas_host and bas_host[0]
        if not bas_host:
            log.msg('Dropping packet from unknown host ' + host)
            return

        try:
            pkt = self.createPacket(packet=datagram,dict=self.dict,secret=six.b(str(bas_host['secret'])))
            pkt.source = (host, port)
            log.msg("::Received an radius request from %s:%s : %s"%(host,port,str(pkt)))
            if self.debug:
                log.msg(pkt.format_str())    

        except packet.PacketError as err:
            log.msg('::Dropping invalid packet: ' + str(err))
            return

        try:
            
            reply = self.processPacket(pkt)
            log.msg("::Send an radius response to %s:%s : %s"%(host,port,reply))
            if self.debug:
                log.msg(reply.format_str())
            self.transport.write(reply.ReplyPacket(), reply.source)  
    

        except PacketError as err:
            log.msg('::Dropping packet from %s: %s' % (host, str(err)))

      


class RADIUSAccess(RADIUS):

    def createPacket(self, **kwargs):
        return utils.AuthPacket2(**kwargs)

    def processPacket(self, req):
        if req.code != packet.AccessRequest:
            raise PacketError(
                    'non-AccessRequest packet on authentication socket')

        reply = req.CreateReply()
        reply.source = req.source

        # domain check
        domain_user = req.get_user_name()
        domain = None
        username = domain_user
        if "@" in domain_user:
            username = domain_user[:domain_user.index("@")]
            self.req["User-Name"] = username
            domain = domain_user[domain_user.index("@")+1:]

        user = store.get_user(username)
        if not user:
            reply.code = packet.AccessReject
            reply['Reply-Message'] = 'user %s not exists'%username
            return reply

        if domain and domain not in user.get('domain'):
            reply.code = packet.AccessReject
            reply['Reply-Message'] = 'user domain %s not match'%domain         
            return reply
            
        # middleware execute
        for mcls in middlewares.auth_objs:
            middle_ware = mcls(req,reply,user)
            if hasattr(middle_ware,'on_auth'):
                if self.debug:
                    log.msg(mcls.__doc__)
                reply = middle_ware.on_auth()
                if reply.code == packet.AccessReject:
                    return reply
                    
        # send accept
        reply['Reply-Message'] = 'success!'
        reply.code=packet.AccessAccept
        return reply
           

class RADIUSAccounting(RADIUS):

    def createPacket(self, **kwargs):
        return utils.AcctPacket2(**kwargs)

    def processPacket(self, req):
        if req.code != packet.AccountingRequest:
            raise PacketError(
                    'non-AccountingRequest packet on authentication socket')

        user = store.get_user(req.get_user_name())

        def do_acct(req,user):
            for mcls in middlewares.acct_objs:
                middle_ware = mcls(req,user)
                if hasattr(middle_ware,'on_acct'):
                    middle_ware.on_acct() 

        reactor.callInThread(do_acct,req,user)           

        reply = req.CreateReply()
        reply.source = req.source
        return reply


if __name__ == '__main__':
    log.startLogging(sys.stdout, 0)
    reactor.listenUDP(1812, RADIUSAccess())
    reactor.listenUDP(1813, RADIUSAccounting())
    reactor.run()
