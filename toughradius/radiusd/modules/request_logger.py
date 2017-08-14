#!/usr/bin/env python
#coding=utf-8

import logging
from toughradius.txradius.radius import packet
from toughradius.txradius import message


def log_auth(req):
    logging.debug('RadiusAccessRequest received from the access device %s:%s'%req.source)
    logging.debug(message.format_packet_str(req))


def log_acct(req):
    logging.debug('RadiusAccountingRequest received from the access device %s:%s'%req.source)
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
