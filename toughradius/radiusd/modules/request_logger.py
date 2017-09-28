#!/usr/bin/env python
#coding=utf-8
import os
import logging
from toughradius.txradius.radius import packet
from toughradius.txradius import message

TOUGHRADIUS_DEBUG_ENABLE = int(os.environ.get('TOUGHRADIUS_DEBUG_ENABLE','0'))


def log_auth(req):
    logging.info('RadiusAccessRequest received from the access device %s:%s'%req.source)
    if TOUGHRADIUS_DEBUG_ENABLE == 1:
        logging.debug(message.format_packet_str(req))


def log_acct(req):
    logging.info('RadiusAccountingRequest received from the access device %s:%s'%req.source)
    if TOUGHRADIUS_DEBUG_ENABLE == 1:
        logging.debug(message.format_packet_str(req))


def handle_radius(req):
    try:
        if req.code == packet.AccessRequest:
            log_auth(req)
        elif req.code == packet.AccountingRequest:
            log_acct(req)
    except:
        logging.exception("request log error")
        
    return req
