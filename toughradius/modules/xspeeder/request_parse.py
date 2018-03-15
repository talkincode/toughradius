#coding=utf-8
from toughradius.modules import parse_std_vlan
from toughradius.modules import parse_std_mac as parse_mac
from toughradius.modules import VENDOR_XSPEEDER
import logging
import re

logger = logging.getLogger(__name__)

xspeeder_fmt =  re.compile(r'\d+/\d+/\d+:(\d+).(\d+)')


#  vlan parse
def parse_xs_vlan(req):
    '''phy_slot/phy_subslot/phy_port:XPI.XCI'''
    nasportid = req.get_nas_portid()
    if not nasportid: return
    matchs = xspeeder_fmt.search(nasportid.lower())
    if matchs:
        req.vlanid1 = matchs.group(1)
        req.vlanid2 = matchs.group(2)
    else:
        req = parse_std_vlan(req)
    return req


def handle_radius(req, debug=False):
    try:
        if int(req.vendor_id) == VENDOR_XSPEEDER:
            req = parse_mac(req)
            req = parse_xs_vlan(req)
    except Exception as err:
        logger.error("request parse error {}".format(err.message),exc_info=debug)

    return req