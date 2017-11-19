#!/usr/bin/env python
#coding=utf-8
import os
import logging
from toughradius.pyrad.radius import packet
from toughradius.pyrad import message

logger = logging.getLogger(__name__)

def log_accept(req,reply):
    logger.info('RadiusAccessAccept send to the access device %s:%s'%req.source)
    if os.environ.get('TOUGHRADIUS_DEBUG_ENABLED', "0") == "1":
        logger.debug(message.format_packet_str(reply))


def log_reject(req,reply):
    logger.info('RadiusAccessReject send to the access device %s:%s'%req.source)
    if os.environ.get('TOUGHRADIUS_DEBUG_ENABLED', "0") == "1":
        logger.debug(message.format_packet_str(reply))


def log_acct(req,reply):
    logger.info('RadiusAccountingResponse send to the access device %s:%s'%req.source)
    if os.environ.get('TOUGHRADIUS_DEBUG_ENABLED', "0") == "1":
        logger.debug(message.format_packet_str(reply))


def handle_radius(req,reply):
    try:
        if reply.code == packet.AccessAccept:
            log_accept(req,reply)
        elif reply.code == packet.AccessReject:
            log_reject(req,reply)
        elif reply.code == packet.AccountingResponse:
            log_acct(req,reply)
    except:
        logger.exception("response log error")

    return reply
