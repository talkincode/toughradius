#coding=utf-8
from toughradius.modules import parse_cisco_vlan as parse_vlan
from toughradius.modules import parse_std_mac as parse_mac
from toughradius.modules import VENDOR_MIKROTIC
import logging

logger = logging.getLogger(__name__)

def handle_radius(req, debug=False):
    try:
        if int(req.vendor_id) == VENDOR_MIKROTIC:
            req = parse_mac(req)
            req = parse_vlan(req)
    except Exception as err:
        logger.error("request parse error {}".format(err.message),exc_info=debug)

    return req