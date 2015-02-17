#!/usr/bin/env python
#coding=utf-8
from twisted.python import log
from toughradius.radiusd.settings import *


def parse_cisco(req):
    for attr in req:
        if attr not in 'Cisco-AVPair':
            continue
        attr_val = req[attr]
        if attr_val.startswith('client-mac-address'):
            mac_addr = attr_val[len("client-mac-address="):]
            mac_addr = mac_addr.replace('.','')
            _mac = (mac_addr[0:2],mac_addr[2:4],mac_addr[4:6],mac_addr[6:8],mac_addr[8:10],mac_addr[10:])
            req.client_macaddr =  ':'.join(_mac)


def parse_radback(req):
    mac_addr = req.get('Mac-Addr')
    if mac_addr:req.client_macaddr = mac_addr.replace('-',':')


def parse_zte(req):
    mac_addr = req.get('Calling-Station-Id')
    if mac_addr:
        mac_addr = mac_addr[12:] 
        _mac = (mac_addr[0:2],mac_addr[2:4],mac_addr[4:6],mac_addr[6:8],mac_addr[8:10],mac_addr[10:])
        req.client_macaddr =  ':'.join(_mac)

  
def parse_h3c(req):
    mac_addr = req.get('H3C-Ip-Host-Addr')
    if mac_addr and len(mac_addr) > 17:
        req.client_macaddr = mac_addr[:-17]
    else:
        req.client_macaddr = mac_addr


_parses = {
            '9' : parse_cisco,
            '2352' : parse_radback,
            '3902' : parse_zte,
            '25506' : parse_h3c
        }


def process(req=None,resp=None,user=None,**kwargs):
    if req.vendor_id in _parses:
        _parses[req.vendor_id](req)






