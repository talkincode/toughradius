#!/usr/bin/env python
# coding=utf-8

from twisted.python import log
from twisted.internet import protocol
from twisted.internet import reactor, defer
from toughradius.txradius.radius import packet
from toughradius.txradius import message
from toughradius.common import six
import time

class RadiusClient(protocol.DatagramProtocol):
    def __init__(self, secret, dictionary, server, authport=1812, acctport=1813,  debug=False):
        self.dict = dictionary
        self.secret = six.b(secret)
        self.server = server
        self.authport = authport
        self.acctport = acctport
        self.debug = debug
        reactor.listenUDP(0, self)

    def close(self):
        if self.transport is not None:
            self.transport.stopListening()
            self.transport = None

    def onError(self, err):
        log.err('Packet process error：%s' % str(err))
        reactor.callLater(0.01, self.close,)
        return err

    def onResult(self, resp):
        reactor.callLater(0.01, self.close,)
        return resp

    def onTimeout(self):
        if not self.deferrd.called:
            defer.timeout(self.deferrd)

    def sendAuth(self, **kwargs):
        timeout_sec = kwargs.pop('timeout',10) 
        User_Password = kwargs.pop("User-Password",None)
        CHAP_Password = kwargs.pop("CHAP-Password",None)
        CHAP_Password_Plaintext = kwargs.pop("CHAP-Password-Plaintext",None)
        CHAP_Challenge = kwargs.get("CHAP-Challenge")
        request = message.AuthMessage(dict=self.dict, secret=self.secret, **kwargs)
        if User_Password:
            request['User-Password'] = request.PwCrypt(User_Password)
        if CHAP_Password:
            if CHAP_Challenge: 
                request['CHAP-Challenge'] = CHAP_Challenge
            request['CHAP-Password'] = CHAP_Password
        if CHAP_Password_Plaintext:
            request['CHAP-Password'] = request.ChapEcrypt(CHAP_Password_Plaintext)

        if self.debug:
            log.msg("Send radius Auth Request to (%s:%s): %s" % (self.server, self.authport, request.format_str()))
        self.transport.write(request.RequestPacket(), (self.server, self.authport))
        self.deferrd = defer.Deferred()
        self.deferrd.addCallbacks(self.onResult,self.onError)
        reactor.callLater(timeout_sec, self.onTimeout,)
        return self.deferrd

    def sendAcct(self, **kwargs):
        timeout_sec = kwargs.pop('timeout',10) 
        request = message.AcctMessage(dict=self.dict, secret=self.secret, **kwargs)
        if self.debug:
            log.msg("Send radius Acct Request to (%s:%s): %s" % (self.server, self.acctport, request.format_str()))
        self.transport.write(request.RequestPacket(), (self.server, self.acctport))
        self.deferrd = defer.Deferred()
        self.deferrd.addCallbacks(self.onResult,self.onError)
        reactor.callLater(timeout_sec, self.onTimeout,)
        return self.deferrd

    def datagramReceived(self, datagram, (host, port)):
        try:
            response = packet.Packet(packet=datagram,dict=self.dict, secret=self.secret)
            if self.debug:
                log.msg("Received Radius Response from %s: %s" % ((host, port), message.format_packet_str(response)))
            self.deferrd.callback(response)
        except Exception as err:
            log.err('Invalid Response packet from %s: %s' % ((host, port), str(err)))
            self.deferrd.errback(err)


def send_auth(secret, dictionary, server, authport=1812, acctport=1813, debug=False, **kwargs):
    return RadiusClient(secret, dictionary, server, authport, acctport, debug).sendAuth(**kwargs)

def send_acct(secret, dictionary, server, authport=1812, acctport=1813, debug=False, **kwargs):
    return RadiusClient(secret, dictionary, server, authport, acctport, debug).sendAcct(**kwargs)









