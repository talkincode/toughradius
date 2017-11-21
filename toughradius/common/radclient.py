#!/usr/bin/env python
# coding: utf-8
import os
import sys
from gevent import socket
from toughradius.pyrad.radius import packet
from toughradius.pyrad.radius import dictionary
from toughradius.pyrad import message
from toughradius.common import six
from toughradius.pyrad.ext import ikuai
import toughradius
import logging
import gevent

logger = logging.getLogger(__name__)

def get_dictionary(dictfile=None):
    '''
    Instantiated radius dictionary, if dictfile not exists, use default dictionary file path

    :param dictfile:

    :return:
    '''
    if dictfile and os.path.exists(dictfile):
        return dictionary.Dictionary(dictfile)
    else:
        return dictionary.Dictionary(os.path.join(os.path.dirname(toughradius.__file__),'dictionarys/dictionary'))

def send_auth(server, port=1812, secret=six.b("testing123"), debug=False, dictfile=None, stat=None, retry=3,**kwargs):
    """
    send auth request

    :param server: radius server ipaddr
    :param port: auth port, default 1812

    :param secret: nas share secret
    :param debug: logging level

    :param dictfile: dictionary file path
    :param kwargs: request params

    :return: auth response
    """
    for i in range(retry):
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
                logger.debug("Send radius auth request %s times to (%s:%s): %s" % (i+1, server, port, request.format_str()))

            sock = socket.socket(type=socket.SOCK_DGRAM)
            sock.settimeout(timeout_sec)
            sock.connect((server,port))
            reqdata = request.RequestPacket()
            sock.send(reqdata)
            if stat:
                stat.incr('auth_req')
                stat.incr('req_bytes', len(reqdata))
            if result:
                data, address = sock.recvfrom(8192)
                reply = request.CreateReply(packet=data)
                if stat:
                    stat.incr('resp_bytes', len(data))
                    if reply.code == packet.AccessReject:
                        stat.incr('auth_reject')
                    elif reply.code == packet.AccessAccept:
                        stat.incr('auth_accept')
                if debug:
                    logger.debug("Recv radius auth response from (%s:%s): %s" % (server, port, reply.format_str()))
                return reply
        except Exception as e:
            if stat:
                stat.incr('auth_drop')
            logger.error("authenticator error {}".format(e.message), exc_info=debug)
            if i < retry-1:
                gevent.sleep((i + 1) * 3)


def send_acct(server, port=1813, secret=six.b("testing123"), debug=False, dictfile=None,stat=None, retry=3, **kwargs):
    """
    send accounting request

    :param server: radius server ipaddr
    :param port: acct port, default 1813

    :param secret: nas share secret
    :param debug: logging level

    :param dictfile: dictionary file path
    :param kwargs: request params

    :return: acct response
    """
    for i in range(retry):
        try:
            radius_dictionary = get_dictionary(dictfile=dictfile)
            timeout_sec = kwargs.pop('timeout', 5)
            result = kwargs.pop('result', True)
            request = message.AcctMessage(dict=radius_dictionary, secret=secret, **kwargs)
            if debug:
                logger.debug("Send radius acct request %s times to (%s:%s): %s" % (i+1, server, port, request.format_str()))

            sock = socket.socket(type=socket.SOCK_DGRAM)
            sock.settimeout(timeout_sec)
            sock.connect((server,port))
            reqdata = request.RequestPacket()
            sock.send(reqdata)
            if stat:
                stat.incr('req_bytes', len(reqdata))
                if request.get_acct_status_type() == message.STATUS_TYPE_START:
                    stat.incr('acct_start')
                elif request.get_acct_status_type() == message.STATUS_TYPE_UPDATE:
                    stat.incr('acct_update')
                elif request.get_acct_status_type() == message.STATUS_TYPE_START:
                    stat.incr('acct_update')

            if result:
                data, address = sock.recvfrom(8192)
                reply = request.CreateReply(packet=data)
                if stat:
                    stat.incr('resp_bytes', len(data))
                    if reply.code == packet.AccountingResponse:
                        stat.incr('acct_resp')
                        if request.get_acct_status_type() == message.STATUS_TYPE_START:
                            stat.incr('online')
                        elif request.get_acct_status_type() == message.STATUS_TYPE_STOP:
                            stat.incr('online', -1)
                if debug:
                    logger.debug("Recv radius auth response from (%s:%s): %s" % (server, port, reply.format_str()))
                return reply
        except Exception as e:
            if stat:
                stat.incr('acct_drop')
            logger.error("accounting error {}".format(e.message), exc_info=debug)
            if i < retry-1:
                gevent.sleep((i + 1) * 3)



def send_coadm(server, port=3799, secret=six.b("testing123"), debug=False, dictfile=None, retry=3,**kwargs):
    """
    send coa disconnect request to nas

    :param server: nas server ipaddr
    :param port: coa port, default 3799

    :param secret: nas share secret
    :param debug: logging level

    :param dictfile: dictionary file path
    :param kwargs: request params

    :return: coa response
    """
    for i in range(retry):
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
                    logger.debug( "Send ikuai radius CoaDmRequest %s times to (%s:%s) [username:%s]: %s" % (i+1, server, port, username, repr(pkg)))
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
            logger.error("accounting error {}".format(e.message), exc_info=debug)
            if i < retry-1:
                gevent.sleep((i+1)*3)
