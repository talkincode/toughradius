#!/usr/bin/env python
#coding=utf-8
from twisted.internet.defer import Deferred
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

class PacketError(Exception):pass

class RADIUS(host.Host, protocol.DatagramProtocol):
    def __init__(self, config="config.yaml",dict=dictionary.Dictionary("res/dictionary"),debug=False):
        host.Host.__init__(self,dict=dict)
        self.debug = debug
        self.bas_ip_pool = {bas['ip_addr']:bas for bas in store.list_bas()}
        with open(config,'rb') as cf:
            self.config = yaml.load(cf)

    def processPacket(self, pkt):
        pass

    def createPacket(self, **kwargs):
        raise NotImplementedError('Attempted to use a pure base class')

    def datagramReceived(self, datagram, (host, port)):
        bas = self.bas_ip_pool.get(host)
        if self.config['bas_host_check']  and not bas:
            return log.msg('Dropping packet from unknown host ' + host)
        secret = bas['secret'] or self.config['bas_default_secret']
        try:
            _packet = self.createPacket(packet=datagram,dict=self.dict,secret=six.b(str(secret)))
            _packet.deferred.addCallbacks(self.reply,self.on_exception)
            _packet.source = (host, port)
            log.msg("::Received an radius request from %s : %s"%((host, port),str(_packet)))
            if self.debug:
                log.msg(_packet.format_str())    
            self.processPacket(_packet)
        except packet.PacketError as err:
            log.msg('::Dropping invalid packet from %s: %s'%((host, port),str(err)))

    def reply(self,reply):
        log.msg("send radius response to %s : %s"%(reply.source,reply))
        if self.debug:
            log.msg(reply.format_str())
        self.transport.write(reply.ReplyPacket(), reply.source)  
 
    def on_exception(self,err):
        log.msg('Packet process error：%s' % str(err))   


class RADIUSAccess(RADIUS):

    def createPacket(self, **kwargs):
        return utils.AuthPacket2(**kwargs)

    def processPacket(self, req):
        if req.code != packet.AccessRequest:
            raise PacketError('non-AccessRequest packet on authentication socket')

        reply = req.CreateReply()
        reply.source = req.source
        user = store.get_user(req.get_user_name())
        # middleware execute
        for mcls in middlewares.auth_objs:
            middle_ware = mcls(req,reply,user)
            if hasattr(middle_ware,'on_auth'):
                if self.debug:
                    log.msg(mcls.__doc__)
                reply = middle_ware.on_auth()
                if reply.code == packet.AccessReject:
                    return req.deferred.callback(reply)
                    
        # send accept
        reply['Reply-Message'] = 'success!'
        reply.code=packet.AccessAccept
        req.deferred.callback(reply)
           

class RADIUSAccounting(RADIUS):

    def createPacket(self, **kwargs):
        return utils.AcctPacket2(**kwargs)

    def processPacket(self, req):
        if req.code != packet.AccountingRequest:
            raise PacketError(
                    'non-AccountingRequest packet on authentication socket')

        def do_acct(req):
            user = store.get_user(req.get_user_name())
            for mcls in middlewares.acct_objs:
                middle_ware = mcls(req,user)
                if hasattr(middle_ware,'on_acct'):
                    middle_ware.on_acct() 

        reply = req.CreateReply()
        reply.source = req.source
        reply.deferred.addCallbacks(do_acct,self.on_exception)    
        req.deferred.callback(reply)
        reply.deferred.callback(req)


if __name__ == '__main__':
    log.startLogging(sys.stdout, 0)
    reactor.listenUDP(1812, RADIUSAccess(debug=True))
    reactor.listenUDP(1813, RADIUSAccounting())
    reactor.run()
