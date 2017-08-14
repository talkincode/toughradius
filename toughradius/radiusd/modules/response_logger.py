#!/usr/bin/env python
#coding=utf-8

import logging
from toughradius.txradius.radius import packet
from toughradius.txradius import message


def log_accept(req,reply):
    logging.debug('RadiusAccessAccept send to the access device %s:%s'%req.source)
    logging.debug(message.format_packet_str(reply))


def log_reject(req,reply):
    logging.debug('RadiusAccessReject send to the access device %s:%s'%req.source)
    logging.debug(message.format_packet_str(reply))


def log_acct(req,reply):
    logging.debug('RadiusAccountingResponse send to the access device %s:%s'%req.source)
    logging.debug(message.format_packet_str(reply))


def handle_radius(req,reply):
    try:
        if reply.code == packet.AccessAccept:
            log_accept(req,reply)
        elif reply.code == packet.AccessReject:
            log_reject(req,reply)
        elif reply.code == packet.AccountingResponse:
            log_acct(req,reply)
    except:
        logging.exception("response log error")

    return reply
