#!/usr/bin/env python
# coding:utf-8
import logging
from toughradius.pyrad.radius import dictionary
from toughradius.pyrad import message
from toughradius.common import six, tools
from toughradius.pyrad.radius import packet
import importlib
import os

class BasicAdapter(object):

    def __init__(self, settings):
        self.settings = settings
        self.logger = logging.getLogger(__name__)
        self.dictionary = dictionary.Dictionary(self.settings.RADIUSD['dictionary'])
        self.xdebug = os.environ.get('TOUGHRADIUS_TRACE_ENABLED',"0")  == '1'
        self.auth_pre = [self.load_module(m) for m in self.settings.MODULES["auth_pre"] if m is not None]
        self.acct_pre = [self.load_module(m) for m in self.settings.MODULES["acct_pre"] if m is not None]
        self.auth_post = [self.load_module(m) for m in self.settings.MODULES["auth_post"] if m is not None]
        self.acct_post = [self.load_module(m) for m in self.settings.MODULES["acct_post"] if m is not None]


    def load_module(self, mdl):
        try:
            self.logger.info('load module %s' % mdl)
            return importlib.import_module(mdl)
        except:
            self.logger.info('load module error, %s' % mdl, exc_info=self.xdebug)


    @tools.timecast
    def handleAuth(self, data, address):
        """
        auth request handle

        :param resp_que:
        :param socket:
        :param data:
        :param address:

        :return:
        """
        try:
            req = self.parseAuthPacket(data,address)
            try:
                prereply = self.processAuth(req)
                reply = self.authReply(req, prereply)
                return reply.ReplyPacket()
            except Exception as e:
                # import pdb;pdb.set_trace()
                errstr = "Handle Radius Auth error {}".format(e.message)
                self.logger.error( errstr,exc_info=self.xdebug)
                reply = self.rejectReply(req,errmsg=errstr)
                return reply.ReplyPacket()
        except Exception as e:
            self.logger.error( "Parse Radius Auth Message error {}".format(e.message),exc_info=self.xdebug)



    @tools.timecast
    def handleAcct(self, data, address):
        """
        acct request handle

        :param resp_que:
        :param socket:
        :param data:
        :param address:

        :return:
        """
        try:
            req = self.parseAcctPacket(data,address)
            prereply = self.processAcct(req)
            reply = self.acctReply(req, prereply)
            return reply.ReplyPacket()
        except Exception as e:
            self.logger.error("Handle Radius Acct error {}".format(e.message),exc_info=self.xdebug)

    def getClient(self, nasip=None, nasid=None):
        """
        fetch nas clients

        Usage example::

            def getClient(self,nasip=None,nasid=None):
                return dict(
                    status=1,
                    nasid='toughac',
                    name='toughac',
                    vendor=0,
                    ipaddr='127.0.0.1',
                    secret='testing123',
                    coaport=3799
                )

        :return: nas dict
        """
        raise NotImplementedError('Attempted to use a pure base class')


    @staticmethod
    def verifyAcctRequest(req):
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

    @staticmethod
    def freeReply(req, **params):
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
        reply_attrs['input_rate'] = params.pop("free_auth_input_limit", 1048576)
        reply_attrs['output_rate'] = params.pop("free_auth_output_limit", 4194304)
        reply_attrs['rate_code'] = params.pop("free_auth_rate_code", "")
        reply_attrs['domain'] = params.pop("free_auth_domain", "")
        reply_attrs['attrs']['Session-Timeout'] = params.pop("max_session_timeout", 86400)
        reply.resp_attrs = reply_attrs
        return reply

    @staticmethod
    def rejectReply(req, errmsg=''):
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

    # @tools.timecast
    def parseAuthPacket(self, datagram, (host, port)):
        """
        parse radius auth request

        :param datagram:

        :return:  pyrad.message
        """
        request = message.AuthMessage(packet=datagram, dict=self.dictionary, secret=six.b(''))
        nas_id = request.get_nas_id()
        client = self.getClient(nasip=host, nasid=nas_id)
        if client:
            request.vendor_id = client['vendor']
            request.secret = six.b(tools.safestr(client['secret']))
        else:
            raise packet.PacketError("Unauthorized Radius Access Device [%s] (%s:%s)" % (nas_id, host, port))

        if request.code != packet.AccessRequest:
            errstr = u'Invalid authenticator request code=%s' % request.code
            raise packet.PacketError(errstr)
        request.source = (host, port)
        for _module in self.auth_pre:
            request = _module.handle_radius(request)
        return request

    # @tools.timecast
    def parseAcctPacket(self, datagram, (host, port)):
        """
        parse radius accounting request

        :param datagram:

        :return: pyrad.message
        """
        vendors = self.settings.VENDORS
        request = message.AcctMessage(packet=datagram, dict=self.dictionary, secret=six.b(''))
        nas_id = request.get_nas_id()
        client = self.getClient(nasip=host, nasid=nas_id)
        if client:
            request.vendor_id = vendors.get(client['vendor'])
            request.secret = tools.safestr(client['secret'])
        else:
            raise packet.PacketError("Unauthorized Radius Access Device [%s] (%s:%s)" % (nas_id, host, port))
        self.verifyAcctRequest(request)
        request.source = (host, port)
        for _module in self.acct_pre:
            request = _module.handle_radius(request)
        return request

    # @tools.timecast
    def authReply(self, req, prereply):
        """
        process radius auth response

        :rtype: object
        :param req:

        :param prereply: dict
        :return: radius reply
        """
        try:
            if not isinstance(prereply,dict):
                raise packet.PacketError("Invalid prereply response, must dict")

            if 'code' not in prereply:
                raise packet.PacketError("Invalid response, no code attr")

            if prereply['code'] > 0:
                raise packet.PacketError("radius authentication failure, %s" % prereply.get("msg", ""))

            reply = req.CreateReply()
            reply.vendor_id = req.vendor_id
            reply.resp_attrs = prereply
            for _module in self.auth_post:
                reply = _module.handle_radius(req, reply)
                if reply is None:
                    raise packet.PacketError("radius authentication message discarded")

                if reply.code == packet.AccessReject:
                    return reply

                if not req.VerifyReply(reply):
                    errstr = u'The authentication message failed to check. \
                    Check that the shared key is consistent'
                    raise packet.PacketError(errstr)
            return reply
        except Exception as e:
            errmsg = "handle radius response error {}".format(e.message)
            logging.error(errmsg, exc_info=self.xdebug)
            return self.rejectReply(req, errmsg)

    # @tools.timecast
    def acctReply(self, req, prereply):
        """
        process radius accounting response

        :param req:
        :param prereply:

        :return:
        """
        try:
            if not isinstance(prereply,dict):
                raise packet.PacketError("Invalid prereply response, must dict")

            if 'code' not in prereply:
                raise packet.PacketError("Invalid response, no code attr")

            if prereply['code'] > 0:
                raise packet.PacketError("radius accounting failure, %s" % prereply.get("msg", ""))

            reply = req.CreateReply()
            for _module in self.acct_post:
                reply = _module.handle_radius(req, reply)
                if reply is None:
                    raise packet.PacketError("radius accounting message discarded")

                if not req.VerifyReply(reply):
                    errstr = '[User:%s] The accounting message failed to check. \
                    Check that the shared key is consistent'
                    raise packet.PacketError(errstr)
            return reply
        except Exception as err:
            raise packet.PacketError("handle radius accounting response error")


    def processAuth(self, req):
        """
        Function delivery to subclass implementation

        :param req:

        :return:
        """
        raise NotImplementedError('Attempted to use a pure base class')

    def processAcct(self, req):
        """
        Function delivery to subclass implementation

        :param req:

        :return:
        """
        raise NotImplementedError('Attempted to use a pure base class')

