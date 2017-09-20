#!/usr/bin/env python
#coding=utf-8
import os
from toughradius.common import six
from twisted.python import log
from twisted.internet import protocol
from twisted.internet import reactor, defer
from toughradius.txradius.radius import packet
from toughradius.txradius.ext import ikuai
from toughradius.txradius import message
from toughradius.txradius.radius import dictionary
from toughradius import txradius

RADIUS_DICT = dictionary.Dictionary(os.path.join(os.path.dirname(txradius.__file__), 'dictionary/dictionary'))

def get_dm_packet(vendor_id, nas_secret, nas_addr, coa_port=3799, **kwargs):
    coa_request = message.CoAMessage(
        code=packet.DisconnectRequest,
        dict=RADIUS_DICT,
        secret=six.b(str(nas_secret)),
        **kwargs
    )
    username = coa_request["User-Name"][0]
    if int(vendor_id) == ikuai.VENDOR_ID:
        pkg = ikuai.create_dm_pkg(six.b(str(nas_secret)), username)
        return (pkg,nas_addr,coa_port)
    else:
        return (coa_request.RequestPacket(),nas_addr,coa_port)


class CoAClient(protocol.DatagramProtocol):
    
    def __init__(self, vendor_id, dictionary, nas_secret, nas_addr, coa_port=3799, debug=False):
        self.dictionary = dictionary
        self.secret = six.b(str(nas_secret))
        self.addr = nas_addr
        self.port = int(coa_port)
        self.vendor_id = int(vendor_id)
        self.debug=debug
        self.uport = reactor.listenUDP(0, self)

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
        
    def sendDisconnect(self, **kwargs):
        timeout_sec = kwargs.pop('timeout',5) 
        coa_req = message.CoAMessage(
            code=packet.DisconnectRequest, dict=self.dictionary, secret=self.secret, **kwargs)   
        username = coa_req["User-Name"][0]
        if self.vendor_id == ikuai.VENDOR_ID:
            pkg = ikuai.create_dm_pkg(self.secret, username)
            if self.debug:
                log.msg("send ikuai radius Coa Request to (%s:%s) [username:%s]: %s"%(
                    self.addr, self.port, username, repr(pkg)))
            self.transport.write(pkg,(self.addr, self.port))
        else:
            if self.debug:
                log.msg("send radius Coa Request to (%s:%s) [username:%s] : %s"%(
                    self.addr, self.port, username, coa_req))
            self.transport.write(coa_req.RequestPacket(),(self.addr, self.port))
        self.deferrd = defer.Deferred()
        self.deferrd.addCallbacks(self.onResult,self.onError)
        reactor.callLater(timeout_sec, self.onTimeout,)
        return self.deferrd

    def datagramReceived(self, datagram, (host, port)):
        try:
            response = packet.Packet(packet=datagram)
            if self.debug:
                log.msg("Received Radius Response from (%s:%s): %s" % (host, port, repr(response)))
            self.deferrd.callback(response.code)
        except Exception as err:
            log.err('Invalid Response packet from %s: %s' % ((host, port), str(err)))
            self.deferrd.errback(err)


def disconnect(vendor_id, dictionary, nas_secret, nas_addr, coa_port=3799, debug=False, **kwargs):
    return CoAClient(vendor_id, dictionary, nas_secret, nas_addr, coa_port, debug).sendDisconnect(**kwargs)



