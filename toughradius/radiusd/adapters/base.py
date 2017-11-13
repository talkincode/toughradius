#!/usr/bin/env python
#coding:utf-8
import gevent
import logging
from toughradius.txradius.radius import dictionary
from toughradius.txradius import message
from toughradius.common import six
from toughradius.txradius.radius import packet
from gevent.pool import Pool
from toughradius.radiusd.modules import (
    request_logger,
    request_mac_parse,
    request_vlan_parse,
    response_logger,
    accept_rate_process
)


class BasicAdapter(object):

    def __init__(self, config):
        self.config = config
        self.pool = Pool(self.config.pool_size)
        self.logger = logging.getLogger(__name__)
        self.dictionary = dictionary.Dictionary(config.radiusd.dictionary)
        self.plugins = []

    def handleAuth(self,socket, data, address):
        try:
            req = self.parseAuthPacket(data,address)
            prereply = self.processAuth(req)
            reply = self.authReply(req, prereply)
            self.pool.spawn(socket.sendto,reply.ReplyPacket(),address)
        except:
            self.logger.error( "Handle Radius Auth error",exc_info=True)

    def handleAcct(self,socket, data, address):
        try:
            req = self.parseAcctPacket(data,address)
            prereply = self.processAcct(req)
            reply = self.acctReply(req, prereply)
            self.pool.spawn(socket.sendto, reply.ReplyPacket(), address)
        except:
            self.logger.error("Handle Radius Acct error",exc_info=True)

    def getClients(self):
        nas = dict(status=1, nasid='toughac', name='toughac', vendor=0, ipaddr='127.0.0.1', secret='secret', coaport=3799)
        return {
            'toughac' : nas,
            '127.0.0.1' : nas
        }

    def verify_acct_request(self, req):
        """
        verify radius accounting request
        :param req:
        """
        if req.code != packet.AccountingRequest:
            errstr = u'Invalid accounting request code=%s' % req.code
            raise packet.PacketError(errstr)

        if not req.VerifyAcctRequest():
            errstr = u'The accounting response check failed. Check that the shared key is consistent'
            raise packet.PacketError(errstr)

    def freeReply(self, req, params=None):
        """
        gen free auth response
        :param req:
        :param params:
        :return:
        """
        reply = req.CreateReply()
        reply.vendor_id = req.vendor_id
        reply['Reply-Message'] = u'User:%s (Free)Authenticate Success' % req.get_user_name()
        reply.code = packet.AccessAccept
        reply_attrs = dict(attrs={})
        reply_attrs['input_rate'] = params.get("free_auth_input_limit", 1048576)
        reply_attrs['output_rate'] = params.get("free_auth_output_limit", 4194304)
        reply_attrs['rate_code'] = params.get("free_auth_rate_code", "")
        reply_attrs['domain'] = params.get("free_auth_domain", "")
        reply_attrs['attrs']['Session-Timeout'] = params.get("max_session_timeout", 86400)
        reply.resp_attrs = reply_attrs
        return reply

    def rejectReply(self, req, errmsg=''):
        """
        gen reject radius auth response
        :param req:
        :param errmsg:
        :return:
        """
        reply = req.CreateReply()
        reply.vendor_id = req.vendor_id
        reply['Reply-Message'] = errmsg
        reply.code = packet.AccessReject
        return reply

    def parseAuthPacket(self, datagram, (host, port)):
        """
        parse radius auth request
        :param datagram:
        :param dictionary:
        :param plugins:
        :return:
        """
        clients = self.getClients()
        vendors = self.config.vendors
        if host in clients:
            client = clients[host]
            request = message.AuthMessage(packet=datagram, dict=self.dictionary, secret=str(client['secret']))
            request.vendor_id = vendors.get(client['vendor'])
        else:
            request = message.AuthMessage(packet=datagram, dict=self.dictionary, secret=six.b(''))
            nas_id = request.get_nas_id()
            if nas_id in clients:
                client = clients[nas_id]
                request.vendor_id = vendors.get(client['vendor'])
                request.secret = six.b(client['secret'])
            else:
                raise packet.PacketError("Unauthorized Radius Access Device [%s] (%s:%s)" % (nas_id, host, port))

        request.source = (host, port)
        request = request_logger.handle_radius(request)
        request = request_mac_parse.handle_radius(request)
        request = request_vlan_parse.handle_radius(request)
        for pg in self.plugins:
            try:
                request = pg.handle_radius(request)
            except:
                pass
        return request

    def parseAcctPacket(self, datagram, (host, port)):
        """
        parse radius accounting request
        :param datagram:
        :param dictionary:
        :param plugins:
        :return: txradius.message
        """
        clients = self.getClients()
        vendors = self.config.vendors
        if host in clients:
            client = clients[host]
            request = message.AcctMessage(packet=datagram, dict=self.dictionary, secret=str(client['secret']))
            request.vendor_id = vendors.get(client['vendor'])
        else:
            request = message.AcctMessage(packet=datagram, dict=self.dictionary, secret=six.b(''))
            nas_id = request.get_nas_id()
            if nas_id in clients:
                client = clients[nas_id]
                request.vendor_id = vendors.get(client['vendor'])
                request.secret = six.b(client['secret'])
            else:
                raise packet.PacketError("Unauthorized Radius Access Device [%s] (%s:%s)" % (nas_id, host, port))

        request.source = (host, port)
        request = request_logger.handle_radius(request)
        request = request_mac_parse.handle_radius(request)
        request = request_vlan_parse.handle_radius(request)
        return request

    def authReply(self, req, prereply=None):
        """
        process radius auth response
        :rtype: object
        :param req:
        :param prereply:
        :return:
        """
        try:
            if 'code' not in prereply:
                raise packet.PacketError("Invalid response, no code attr")

            if prereply['code'] > 0:
                raise packet.PacketError("radius authentication failure, %s" % prereply.get("msg", ""))

            reply = req.CreateReply()
            reply.vendor_id = req.vendor_id
            reply.resp_attrs = prereply
            for module in (response_logger, accept_rate_process):
                reply = module.handle_radius(req, reply)
                if reply is None:
                    raise packet.PacketError("radius authentication message discarded")

                if not req.VerifyReply(reply):
                    errstr = u'The authentication message failed to check. \
                    Check that the shared key is consistent'
                    raise packet.PacketError(errstr)
            return reply
        except:
            errmsg = "handle radius response error"
            logging.exception(errmsg)
            return self.rejectReply(req, errmsg)

    def acctReply(self, req, prereply):
        """
        process radius accounting response
        :param req:
        :param prereply:
        :return:
        """
        try:
            if 'code' not in prereply:
                raise packet.PacketError("Invalid response, no code attr")

            if prereply['code'] > 0:
                raise packet.PacketError("radius accounting failure, %s" % prereply.get("msg", ""))

            reply = req.CreateReply()
            for module in (response_logger,):
                reply = module.handle_radius(req, reply)
                if reply is None:
                    raise packet.PacketError("radius accounting message discarded")

                if not req.VerifyReply(reply):
                    errstr = '[User:%s] The accounting message failed to check. \
                    Check that the shared key is consistent'
                    raise packet.PacketError(errstr)
            return reply
        except:
            raise packet.PacketError("handle radius accounting response error")




    def processAuth(self, req):
        raise NotImplementedError

    def processAcct(self, req):
        raise NotImplementedError

