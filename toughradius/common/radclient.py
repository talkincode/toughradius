#!/usr/bin/env python
# coding: utf-8
import os
import sys
from gevent import socket
from toughradius.txradius.radius import packet
from toughradius.txradius.radius import dictionary
from toughradius.txradius import message
from toughradius.common import six
from toughradius.txradius.ext import ikuai
import toughradius
import logging

logger = logging.getLogger(__name__)

def get_dictionary(dictfile=None):
    if dictfile and os.path.exists(dictfile):
        return dictionary.Dictionary(dictfile)
    else:
        return dictionary.Dictionary(os.path.join(os.path.dirname(toughradius.__file__),'dictionarys/dictionary'))

def send_auth(server, port=1812, secret=six.b("testing123"), debug=False, dictfile=None, **kwargs):
    try:
        radius_dictionary = get_dictionary(dictfile=dictfile)
        timeout_sec = kwargs.pop('timeout', 5)
        result = kwargs.pop('result', True)
        User_Password = kwargs.pop("User-Password", None)
        CHAP_Password = kwargs.pop("CHAP-Password", None)
        CHAP_Password_Plaintext = kwargs.pop("CHAP-Password-Plaintext", None)
        CHAP_Challenge = kwargs.get("CHAP-Challenge")
        request = message.AuthMessage(dict=radius_dictionary, secret=secret, **kwargs)
        if User_Password:
            request['User-Password'] = request.PwCrypt(User_Password)
        if CHAP_Password:
            if CHAP_Challenge:
                request['CHAP-Challenge'] = CHAP_Challenge
            request['CHAP-Password'] = CHAP_Password
        if CHAP_Password_Plaintext:
            request['CHAP-Password'] = request.ChapEcrypt(CHAP_Password_Plaintext)

        if debug:
            logger.debug("Send radius auth request to (%s:%s): %s" % (server, port, request.format_str()))

        sock = socket.socket(type=socket.SOCK_DGRAM)
        sock.settimeout(timeout_sec)
        sock.connect((server,port))
        sock.send(request.RequestPacket())
        if result:
            data, address = sock.recvfrom(8192)
            reply = request.CreateReply(packet=data)
            if debug:
                logger.debug("Recv radius auth response from (%s:%s): %s" % (server, port, reply.format_str()))
            return reply
    except Exception as e:
        logger.error("authenticator error {}".format(e.message), exc_info=True)


def send_acct(server, port=1813, secret=six.b("testing123"), debug=False, dictfile=None, **kwargs):
    try:
        radius_dictionary = get_dictionary(dictfile=dictfile)
        timeout_sec = kwargs.pop('timeout', 5)
        result = kwargs.pop('result', True)
        request = message.AcctMessage(dict=radius_dictionary, secret=secret, **kwargs)
        if debug:
            logger.debug("Send radius acct request to (%s:%s): %s" % (server, port, request.format_str()))

        sock = socket.socket(type=socket.SOCK_DGRAM)
        sock.settimeout(timeout_sec)
        sock.connect((server,port))
        sock.send(request.RequestPacket())
        if result:
            data, address = sock.recvfrom(8192)
            reply = request.CreateReply(packet=data)
            if debug:
                logger.debug("Recv radius auth response from (%s:%s): %s" % (server, port, reply.format_str()))
            return reply
    except Exception as e:
        logger.error("accounting error {}".format(e.message), exc_info=True)


def send_coadm(server, port=3799, secret=six.b("testing123"), debug=False, dictfile=None, **kwargs):
    try:
        radius_dictionary = get_dictionary(dictfile=dictfile)
        timeout_sec = kwargs.pop('timeout', 5)
        result = kwargs.pop('result', True)
        vendor_id = kwargs.pop('vendor_id', 0)
        request = message.CoAMessage(code=packet.DisconnectRequest, dict=radius_dictionary, secret=secret, **kwargs)
        username = request["User-Name"][0]
        if vendor_id == ikuai.VENDOR_ID:
            pkg = ikuai.create_dm_pkg(secret, username)
            if debug:
                logger.debug( "Send ikuai radius CoaDmRequest to (%s:%s) [username:%s]: %s" % (server, port, username, repr(pkg)))
        else:
            pkg = request.RequestPacket()
            if debug:
                logger.debug("Send radius CoaDmRequest to (%s:%s) [username:%s]: %s" % (server, port, username, request.format_str()))

        sock = socket.socket(type=socket.SOCK_DGRAM)
        sock.settimeout(timeout_sec)
        sock.connect((server,port))
        sock.send(pkg)
        if result:
            data, address = sock.recvfrom(8192)
            if vendor_id != ikuai.VENDOR_ID:
                reply = request.CreateReply(packet=data)
                if debug:
                    logger.debug("Recv radius coa dm response from (%s:%s): %s" % (server, port, reply.format_str()))
                return reply
            else:
                if debug:
                    logger.debug("Recv radius ik coa dm response from (%s:%s): %s" % (server, port, repr(data)))
                return data
    except Exception as e:
        logger.error("accounting error {}".format(e.message), exc_info=True)
