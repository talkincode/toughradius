#!/usr/bin/env python
#coding=utf-8
from twisted.python import log
from toughradius.radiusd.settings import *

#  vlan parse          
def parse_cisco(req):
    '''phy_slot/phy_subslot/phy_port:XPI.XCI'''
    nasportid = req.get('NAS-Port-Id')
    if not nasportid:return
    nasportid = nasportid.lower()
    def parse_vlanid():
        ind = nasportid.find(':')
        if ind == -1:return
        ind2 = nasportid.find('.',ind)
        if ind2 == -1:
            req.vlanid = int(nasportid[ind+1])
        else:
            req.vlanid = int(nasportid[ind+1:ind2])
    def parse_vlanid2():
        ind = nasportid.find('.')
        if ind == -1:return
        ind2 = nasportid.find(' ',ind)
        if ind2 == -1:
            req.vlanid2 = int(nasportid[ind+1])
        else:
            req.vlanid2 = int(nasportid[ind+1:ind2])
            
    parse_vlanid()
    parse_vlanid2()


def parse_std(req):
    ''''''
    nasportid = req.get('NAS-Port-Id')
    if not nasportid:return
    nasportid = nasportid.lower()
    def parse_vlanid():
        ind = nasportid.find('vlanid=')
        if ind == -1:return
        ind2 = nasportid.find(';',ind)
        if ind2 == -1:
            req.vlanid = int(nasportid[ind+7])
        else:
            req.vlanid = int(nasportid[ind+7:ind2])
            
    def parse_vlanid2():
        ind = nasportid.find('vlanid2=')
        if ind == -1:return
        ind2 = nasportid.find(';',ind)
        if ind2 == -1:
            req.vlanid2 = int(nasportid[ind+8])
        else:
            req.vlanid2 = int(nasportid[ind+8:ind2])
            
    parse_vlanid()
    parse_vlanid2() 

def parse_ros(req):
    ''''''
    nasportid = req.get('NAS-Port-Id')
    if not nasportid:return
    nasportid = nasportid.lower()
    def parse_vlanid():
        ind = nasportid.find(':')
        if ind == -1:return        
        ind2 = nasportid.find(' ',ind)
        if ind2 == -1:return
        req.vlanid = int(nasportid[ind+1:ind2])
        
    def parse_vlanid2():
        ind = nasportid.find(':')
        if ind == -1:return
        ind2 = nasportid.find('.',ind)
        if ind2 == -1:return
        req.vlanid2 = int(nasportid[ind+1:ind2])   
    parse_vlanid()
    parse_vlanid2()               
  
parse_radback = parse_ros
parse_zte = parse_ros

_parses = {
    '0' : parse_std,
    '9' : parse_cisco,
    '3041' : parse_cisco,
    '2352' : parse_radback,
    '2011' : parse_std,
    '25506' : parse_std,
    '3902' : parse_zte,
    '14988' : parse_ros
}

def process(req=None,resp=None,user=None,**kwargs):
    if req.vendor_id in _parses:
        _parses[req.vendor_id](req)

