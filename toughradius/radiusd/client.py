#!/usr/bin/env python
# coding:utf-8
from toughradius.radiusd.pyrad import packet
from toughradius.radiusd.pyrad import dictionary
from toughradius.radiusd import utils
from twisted.python import log
from twisted.internet import defer
from twisted.internet import protocol
from twisted.internet import reactor
import time
import six
import os
import logging


class Timeout(Exception):
    """Simple exception class which is raised when a timeout occurs
    while waiting for a ac server to respond."""


__vis_filter = """................................ !"#$%&\'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[.]^_`abcdefghijklmnopqrstuvwxyz{|}~................................................................................................................................."""


def hexdump(buf, length=16):
    """Return a hexdump output string of the given buffer."""
    n = 0
    res = []
    while buf:
        line, buf = buf[:length], buf[length:]
        hexa = ' '.join(['%02x' % ord(x) for x in line])
        line = line.translate(__vis_filter)
        res.append('  %04d:  %-*s %s' % (n, length * 3, hexa, line))
        n += length
    return '\n'.join(res)


def sleep(secs):
    d = defer.Deferred()
    reactor.callLater(secs, d.callback, None)
    return d


class RadiusClient(protocol.DatagramProtocol):
    results = {}

    def __init__(self, dictfile=None, timeout=5, retry=5, debug=True):
        self.dict = dictionary.Dictionary(dictfile)
        self.timeout = timeout
        self.retry = retry
        self.debug = debug
        self.port = reactor.listenUDP(0, self)

    def close(self):
        self.transport = None
        self.results.clear()
        self.port.stopListening()

    def CreateAcctPacket(self, secret=six.b(''), **kwargs):
        return utils.AcctPacket2(code=packet.AccessRequest, dict=self.dict, secret=secret, **kwargs)

    def CreateAuthPacket(self, secret=six.b(''), **kwargs):
        return utils.AuthPacket2(code=packet.AccountingRequest, dict=self.dict, secret=secret, **kwargs)

    def CreateDmPacket(self, secret=six.b(''), **kwargs):
        return utils.CoAPacket2(code=packet.DisconnectRequest, dict=self.dict, secret=secret, **kwargs)


    @defer.inlineCallbacks
    def sendto(self, req, (host, port), recv=True):
        if self.debug:
            log.msg(req.format_str(), level=logging.DEBUG)

        log.msg(":: Send Radius Request To (%s:%s) >> %s" % (host, port, repr(req)))

        if not recv:
            self.transport.write(str(req), (host, port))
            return

        for attempt in range(self.retry):
            if attempt and req.code == packet.AccountingRequest:
                if "Acct-Delay-Time" in req:
                    req["Acct-Delay-Time"] = \
                        req["Acct-Delay-Time"][0] + self.timeout
                else:
                    req["Acct-Delay-Time"] = self.timeout

            self.transport.write(str(req), (host, port))
            now = time.time()
            waitto = now + self.timeout
            while now < waitto:
                if req.id in self.results:
                    defer.returnValue(self.results.pop(req.id))
                    return
                else:
                    now = time.time()
                    yield sleep(0.0001)
                    continue
        raise Timeout

    def datagramReceived(self, datagram, (host, port)):
        try:
            resp = self.createPacket(packet=datagram)
            resp.source = (host, port)
            self.results[resp.id] = resp
            log.msg("::Received Radius Response From %s: %s" % ((host, port), str(resp)), level=logging.INFO)
            if self.debug:
                log.msg(resp.format_str(), level=logging.DEBUG)
        except packet.PacketError as err:
            log.err(err, '::Dropping invalid Radius  Response  From %s: %s' % ((host, port), str(err)))


client = RadiusClient(dictfile=os.path.join(os.path.split(__file__)[0], 'dicts/dictionary'))