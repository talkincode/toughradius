#!/usr/bin/env python
#coding:utf-8
from __future__ import unicode_literals
import logging
import six
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

def parse_auth_packet(datagram,(host,port),client_config,dictionary=None):
    """
    parse radius auth request
    :param datagram:
    :param client_config:
    :param dictionary:
    :return:
    """
    if host in client_config.defaults:
        client = client_config.defaults[host]
        authreq = message.AuthMessage(packet=datagram, dict=dictionary,secret=str(client['secret']))
        authreq.vendor_id=client_config.vendors.get(client['vendor'])
    else:
        authreq = message.AuthMessage(packet=datagram,dict=dictionary, secret=six.b(''))
        nas_id = authreq.get_nas_id()
        _client = [c for c in client_config.defaults.itervalues() if c['nasid'] == nas_id]
        if _client:
            client = _client[0]
            authreq.vendor_id = client_config.vendors.get(client['vendor'])
            authreq.secret = six.b(client['secret'])
        else:
            raise packet.PacketError("Unauthorized Radius Access Device [%s] (%s:%s)"%(nas_id,host,port))

    authreq.source = (host,port)
    authreq = request_logger.handle_radius(authreq)
    authreq = request_mac_parse.handle_radius(authreq)
    authreq = request_vlan_parse.handle_radius(authreq)

    return authreq

def parse_acct_packet(datagram,(host,port),client_config,dictionary=None):
    """
    parse radius accounting request
    :param datagram:
    :param client_config:
    :param dictionary:
    :return: txradius.message
    """
    if host in client_config.defaults:
        client = client_config.defaults[host]
        acctreq = message.AcctMessage(packet=datagram, dict=dictionary, secret=str(client['secret']))
        acctreq.vendor_id=client_config.vendors.get(client['vendor'])
    else:
        acctreq = message.AcctMessage(packet=datagram,dict=dictionary, secret=six.b(''))
        nas_id = acctreq.get_nas_id()
        _client = [c for c in client_config.defaults.itervalues() if c['nasid'] == nas_id]
        if _client:
            client = _client[0]
            acctreq.vendor_id = client_config.vendors.get(client['vendor'])
            acctreq.secret = six.b(client['secret'])
        else:
            raise packet.PacketError("Unauthorized Radius Access Device [%s] (%s:%s)"%(nas_id,host,port))


    acctreq.source = (host,port)
    acctreq = request_logger.handle_radius(acctreq)
    acctreq = request_mac_parse.handle_radius(acctreq)
    acctreq = request_vlan_parse.handle_radius(acctreq)
    return acctreq



def process_auth_reply(req, prereply):
    """
    process radius auth response
    :rtype: object
    :param req:
    :param prereply:
    :return:
    """
    reply = req.CreateReply()
    reply.vendor_id = req.vendor_id
    reply.resp_attrs = prereply

    try:
        for module in (response_logger,accept_rate_process):
            reply = module.handle_radius(req,reply)
            if reply is None:
                raise packet.PacketError("radius authentication message discarded")
        
            if not req.VerifyReply(reply):
                errstr = u'The authentication message failed to check. \
                Check that the shared key is consistent'
                raise packet.PacketError(errstr)
    except:
        errmsg="handle radius response error"
        logging.exception(errmsg)
        return reject_reply(req,errmsg)

    return reply

def process_acct_reply(req, prereply):
    """
    process radius accounting response
    :param req:
    :param prereply:
    :return:
    """
    reply = req.CreateReply()
    try:
        for module in (response_logger,):
            reply = module.handle_radius(req,reply)
            if reply is None:
                raise packet.PacketError("radius accounting message discarded")
        
            if not req.VerifyReply(reply):
                errstr = '[User:%s] The accounting message failed to check. \
                Check that the shared key is consistent'
                raise packet.PacketError(errstr)
    except:
        raise packet.PacketError("handle radius accounting response error")
    return reply




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

