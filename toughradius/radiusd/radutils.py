#!/usr/bin/env python
#coding:utf-8
# from __future__ import unicode_literals
import logging
import datetime
from toughradius.common import six
from toughradius.txradius.radius import packet
from toughradius.txradius import message
from toughradius.radiusd.modules import (
    request_logger,
    request_mac_parse,
    request_vlan_parse,
    response_logger,
    accept_rate_process
)

Acct_Status_Start = 1
Acct_Status_Stop = 2
Acct_Status_Update = 3
Acct_Status_On = 7
Acct_Status_Off = 8

def parse_auth_packet(datagram,(host,port),vendors,clients,dictionary=None,plugins=[]):
    """
    parse radius auth request
    :param datagram:
    :param vendors:
    :param clients:
    :param dictionary:
    :param plugins:
    :return:
    """
    if host in clients:
        client = clients[host]
        request = message.AuthMessage(packet=datagram, dict=dictionary,secret=str(client['secret']))
        request.vendor_id=vendors.get(client['vendor'])
    else:
        request = message.AuthMessage(packet=datagram,dict=dictionary, secret=six.b(''))
        nas_id = request.get_nas_id()
        if nas_id in clients:
            client = clients[nas_id]
            request.vendor_id = vendors.get(client['vendor'])
            request.secret = six.b(client['secret'])
        else:
            raise packet.PacketError("Unauthorized Radius Access Device [%s] (%s:%s)"%(nas_id,host,port))

    request.source = (host,port)
    request = request_logger.handle_radius(request)
    request = request_mac_parse.handle_radius(request)
    request = request_vlan_parse.handle_radius(request)
    for pg in plugins:
        try:
            request = pg.handle_radius(request)
        except:
            pass
    return request

def parse_acct_packet(datagram,(host,port),vendors,clients,dictionary=None,plugins=[]):
    """
    parse radius accounting request
    :param datagram:
    :param vendors:
    :param clients:
    :param dictionary:
    :param plugins:
    :return: txradius.message
    """
    if host in clients:
        client = clients[host]
        request = message.AcctMessage(packet=datagram, dict=dictionary,secret=str(client['secret']))
        request.vendor_id=vendors.get(client['vendor'])
    else:
        request = message.AcctMessage(packet=datagram,dict=dictionary, secret=six.b(''))
        nas_id = request.get_nas_id()
        if nas_id in clients:
            client = clients[nas_id]
            request.vendor_id = vendors.get(client['vendor'])
            request.secret = six.b(client['secret'])
        else:
            raise packet.PacketError("Unauthorized Radius Access Device [%s] (%s:%s)"%(nas_id,host,port))

    request.source = (host,port)
    request = request_logger.handle_radius(request)
    request = request_mac_parse.handle_radius(request)
    request = request_vlan_parse.handle_radius(request)
    return request

def process_auth_reply(req, prereply={}):
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
            raise packet.PacketError("radius authentication failure, %s" % prereply.get("msg",""))

        reply = req.CreateReply()
        reply.vendor_id = req.vendor_id
        reply.resp_attrs = prereply
        for module in (response_logger,accept_rate_process):
            reply = module.handle_radius(req,reply)
            if reply is None:
                raise packet.PacketError("radius authentication message discarded")

            if not req.VerifyReply(reply):
                errstr = u'The authentication message failed to check. \
                Check that the shared key is consistent'
                raise packet.PacketError(errstr)
        return reply
    except:
        errmsg="handle radius response error"
        logging.exception(errmsg)
        return reject_reply(req,errmsg)



def process_acct_reply(req, prereply):
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
            raise packet.PacketError("radius accounting failure, %s" % prereply.get("msg",""))

        reply = req.CreateReply()
        for module in (response_logger,):
            reply = module.handle_radius(req,reply)
            if reply is None:
                raise packet.PacketError("radius accounting message discarded")

            if not req.VerifyReply(reply):
                errstr = '[User:%s] The accounting message failed to check. \
                Check that the shared key is consistent'
                raise packet.PacketError(errstr)
        return reply
    except:
        raise packet.PacketError("handle radius accounting response error")


def verify_acct_request(req):
    """
    verify radius accounting request
    :param req:
    """
    if req.code != packet.AccountingRequest:
        errstr = u'Invalid accounting request code=%s'%req.code
        raise packet.PacketError(errstr)

    if not req.VerifyAcctRequest():
        errstr = u'The accounting response check failed. Check that the shared key is consistent'
        raise packet.PacketError(errstr)


def free_reply(req, params={}):
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
    reply_attrs['input_rate'] = params.get("free_auth_input_limit",1048576)
    reply_attrs['output_rate'] = params.get("free_auth_output_limit",4194304)
    reply_attrs['rate_code'] = params.get("free_auth_rate_code","")
    reply_attrs['domain'] = params.get("free_auth_domain","")
    reply_attrs['attrs']['Session-Timeout'] = params.get("max_session_timeout",86400)
    reply.resp_attrs = reply_attrs
    return reply

def reject_reply(req,errmsg=''):
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
















