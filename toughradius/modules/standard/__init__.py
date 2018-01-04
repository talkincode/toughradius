#coding=utf-8

from toughradius.modules import parse_std_vlan
from toughradius.modules import get_radius_attr
from toughradius.modules import VENDOR_STD
import logging

logger = logging.getLogger(__name__)

class RequestParse(object):

    parse_vlan = parse_std_vlan

    @staticmethod
    def parse_mac(req):
        mac_addr = get_radius_attr(req, 'Calling-Station-Id')
        if mac_addr:
            req.client_mac = mac_addr.replace('-', ':')
        return req


    @staticmethod
    def handle_radius(req,debug=False):
        try:
            if int(req.vendor_id) == VENDOR_STD:
                req = RequestParse.parse_mac(req)
                req = RequestParse.parse_vlan(req)
        except Exception as err:
            logger.error("request parse error {}".format(err.message),exc_info=debug)

        return req

request_parse = RequestParse

__all__ = ['request_parse']