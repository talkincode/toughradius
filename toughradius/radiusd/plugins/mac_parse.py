#!/usr/bin/env python
#coding=utf-8

from toughlib import logger

def get_radius_attr(req,key):
    attr = req[key]
    if isinstance(attr,list) and len(attr) > 0:
        return attr[0]
    else:
        return attr


def parse_cisco(req):
    for attr in req:
        if attr not in 'Cisco-AVPair':
            continue
        attr_val = req[attr]
        if attr_val.startswith('client-mac-address'):
            mac_addr = attr_val[len("client-mac-address="):]
            mac_addr = mac_addr.replace('.','')
            _mac = (mac_addr[0:2],mac_addr[2:4],mac_addr[4:6],mac_addr[6:8],mac_addr[8:10],mac_addr[10:])
            req.client_mac =  ':'.join(_mac)
    return req


def parse_radback(req):
    mac_addr = get_radius_attr(req,'Mac-Addr')
    if mac_addr:
        req.client_mac = mac_addr.replace('-',':')
    return req


def parse_zte(req):
    mac_addr = get_radius_attr(req,'Calling-Station-Id')
    if mac_addr:
        mac_addr = mac_addr[12:] 
        _mac = (mac_addr[0:2],mac_addr[2:4],mac_addr[4:6],mac_addr[6:8],mac_addr[8:10],mac_addr[10:])
        req.client_mac =  ':'.join(_mac)
    return req

def parse_normal(req):
    mac_addr = get_radius_attr(req,'Calling-Station-Id')
    if mac_addr:
        req.client_mac = mac_addr.replace('-', ':')
    return req

def parse_h3c(req):
    mac_addr = get_radius_attr(req,'H3C-Ip-Host-Addr')
    if mac_addr and len(mac_addr) > 17:
        req.client_mac = mac_addr[:-17]
    else:
        req.client_mac = mac_addr

    return req


_parses = {
    '0' : parse_normal,
    '9' : parse_cisco,
    '2352' : parse_radback,
    '3902' : parse_zte,
    '14988' : parse_normal,
    '25506' : parse_h3c,
    '39999' : parse_normal,
}

def radius_parse(req):
    try:
        vendorid = str(req.vendor_id)
        if vendorid in _parses:
            _parses[vendorid](req)
        else:
            parse_normal(req)
    except Exception as err:
        logger.exception(err,trace="radius",tag="radius_mac_parse_error")

    return req


plugin_name = 'radius mac parse'
plugin_types = ['radius_auth_req','radius_acct_req']
plugin_priority = 100
plugin_func = radius_parse


