# coding=utf-8

from toughradius.modules import get_radius_attr
from toughradius.modules import parse_std_vlan
from toughradius.modules import VENDOR_H3C
import logging
logger = logging.getLogger(__name__)


parse_vlan = parse_std_vlan

def parse_mac(req):
    mac_addr = get_radius_attr(req, 'H3C-Ip-Host-Addr')
    if mac_addr and len(mac_addr) > 17:
        req.client_mac = mac_addr[:-17]
    else:
        req.client_mac = mac_addr

    return req

def handle_radius(req, debug=False):
    try:
        if int(req.vendor_id) == VENDOR_H3C:
            req = parse_mac(req)
            req = parse_vlan(req)
    except Exception as err:
        logger.error("request parse error {}".format(err.message), exc_info=debug)

    return req