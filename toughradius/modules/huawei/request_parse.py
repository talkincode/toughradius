# coding=utf-8

from toughradius.modules import parse_std_vlan
from toughradius.modules import parse_std_mac
from toughradius.modules import VENDOR_HUAWEI
import logging

logger = logging.getLogger(__name__)


parse_vlan = parse_std_vlan
parse_mac = parse_std_mac


def handle_radius(req, debug=False):
    try:
        if int(req.vendor_id) == VENDOR_HUAWEI:
            req = parse_mac(req)
            req = parse_vlan(req)
    except Exception as err:
        logger.error("request parse error {}".format(err.message), exc_info=debug)

    return req
