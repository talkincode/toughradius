#coding=utf-8
import re

__version__ = "0.0.1"
__license__ = 'Apache License 2.0'

VENDOR_STD = 0
VENDOR_ALCATEL = 3041
VENDOR_CISCO = 9
VENDOR_H3C = 25506
VENDOR_HUAWEI = 2011
VENDOR_JUNIPER = 2636
VENDOR_MICROSOFT = 311
VENDOR_MIKROTIC = 14988
VENDOR_XSPEEDER = 26732
VENDOR_RADBACK = 2352
VENDOR_ZTE = 3902
VENDOR_IKUAI = 10055

vlan_fmt = re.compile(r'\w+\s\d+/\d+/\d+:(\d+).(\d+)\s')

def get_radius_attr(req,key):
    if key not in req:
        return ''
    attr = req[key]
    if isinstance(attr,list) and len(attr) > 0:
        return attr[0]
    else:
        return attr

def parse_std_mac(req):
    mac_addr = get_radius_attr(req, 'Calling-Station-Id')
    if mac_addr:
        req.client_mac = mac_addr.replace('-', ':')
    return req


def parse_std_vlan(req):
    nasportid = req.get_nas_portid()
    if not nasportid:
        return req
    nasportid = nasportid.lower()

    def parse_vlanid():
        ind = nasportid.find('vlanid=')
        if ind == -1: return
        ind2 = nasportid.find(';', ind)
        if ind2 == -1:
            req.vlanid = int(nasportid[ind + 7])
        else:
            req.vlanid = int(nasportid[ind + 7:ind2])

    def parse_vlanid2():
        ind = nasportid.find('vlanid2=')
        if ind == -1: return
        ind2 = nasportid.find(';', ind)
        if ind2 == -1:
            req.vlanid2 = int(nasportid[ind + 8])
        else:
            req.vlanid2 = int(nasportid[ind + 8:ind2])

    parse_vlanid()
    parse_vlanid2()
    return req


def parse_cisco_vlan(req):
    """'phy_slot/phy_subslot/phy_port:XPI.XCI"""
    nasportid = req.get_nas_portid()
    if not nasportid:
        return req
    matchs = vlan_fmt.search(nasportid.lower())
    if matchs:
        req.vlanid1 = matchs.group(1)
        req.vlanid2 = matchs.group(2)
    return req